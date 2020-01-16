/*
   Copyright 2020 the original author or authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package ofagent

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/ofagent-go/internal/pkg/openflow"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/voltha"
	"google.golang.org/grpc"
	"sync"
	"time"
)

var logger, _ = log.AddPackage(log.JSON, log.DebugLevel, nil)

type OFAgent struct {
	VolthaApiEndPoint    string
	OFControllerEndPoint string

	volthaConnection *grpc.ClientConn
	volthaClient     voltha.VolthaServiceClient
	packetOutChannel chan voltha.PacketOut
	mapLock          sync.Mutex
	clientMap        map[string]*openflow.OFClient
}

func NewOFAgent(config *OFAgent) (*OFAgent, error) {
	ofa := OFAgent{
		VolthaApiEndPoint:    config.VolthaApiEndPoint,
		OFControllerEndPoint: config.OFControllerEndPoint,
		packetOutChannel:     make(chan voltha.PacketOut),
		clientMap:            make(map[string]*openflow.OFClient),
	}

	return &ofa, nil
}

// Run - make the inital connection to voltha and kicks off io streams
func (ofa *OFAgent) Run() {

	if logger.V(log.DebugLevel) {
		logger.Debugw("Starting GRPC - VOLTHA client",
			log.Fields{
				"voltha-endpoint":     ofa.VolthaApiEndPoint,
				"controller-endpoint": ofa.OFControllerEndPoint})
	}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(ofa.VolthaApiEndPoint, opts...)
	defer conn.Close()
	if err != nil {
		log.Fatalw("ofagent failed to connection to voltha API",
			log.Fields{
				"endpoint": ofa.VolthaApiEndPoint,
				"error":    err})
	}
	ofa.volthaClient = voltha.NewVolthaServiceClient(conn)

	go ofa.receivePacketIn()
	go ofa.receiveChangeEvent()
	go ofa.streamPacketOut()

	for {
		if logger.V(log.DebugLevel) {
			logger.Debugln("GrpcClient entering device refresh pull")
		}
		deviceList, err := ofa.volthaClient.ListLogicalDevices(context.Background(), &empty.Empty{})
		if err != nil {
			logger.Errorw("GrpcClient getDeviceList failed", log.Fields{"Error": err})
		}
		devices := deviceList.GetItems()
		ofa.refreshDeviceList(devices)
		time.Sleep(time.Minute)
	}
}

func (ofa *OFAgent) refreshDeviceList(devices []*voltha.LogicalDevice) {
	//first find the new ones

	var toAdd []string
	var toDel []string
	var deviceIDMap = make(map[string]string)
	for i := 0; i < len(devices); i++ {
		deviceID := devices[i].GetId()
		deviceIDMap[deviceID] = deviceID
		if ofa.clientMap[deviceID] == nil {
			toAdd = append(toAdd, deviceID)
		}
	}
	for key := range ofa.clientMap {
		if deviceIDMap[key] == "" {
			toDel = append(toDel, key)
		}
	}
	if logger.V(log.DebugLevel) {
		logger.Debugw("GrpcClient refreshDeviceList", log.Fields{"ToAdd": toAdd, "ToDel": toDel})
	}
	for i := 0; i < len(toAdd); i++ {
		var client = ofa.addOFClient(toAdd[i])
		go client.Start()
	}
	for i := 0; i < len(toDel); i++ {
		ofa.clientMap[toDel[i]].End()
		ofa.mapLock.Lock()
		delete(ofa.clientMap, toDel[i])
		ofa.mapLock.Unlock()
	}
}
func (ofa *OFAgent) addOFClient(deviceID string) *openflow.OFClient {
	if logger.V(log.DebugLevel) {
		logger.Debugw("GrpcClient addClient called ", log.Fields{"device-id": deviceID})
	}
	ofa.mapLock.Lock()
	ofc := ofa.clientMap[deviceID]
	if ofc == nil {
		ofc = openflow.NewOFClient(deviceID, ofa.OFControllerEndPoint, ofa.volthaClient, ofa.packetOutChannel, true)
		go ofc.Start()
		ofa.clientMap[deviceID] = ofc
	}
	ofa.mapLock.Unlock()
	logger.Debugw("Finished with addClient", log.Fields{"deviceID": deviceID})
	return ofc
}

//GetClient Returns a pointer to the OpenFlow client
func (ofa *OFAgent) GetOFClient(deviceID string) *openflow.OFClient {
	ofc := ofa.clientMap[deviceID]
	if ofc == nil {
		ofc = ofa.addOFClient(deviceID)
	}
	return ofc
}
