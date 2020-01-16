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
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/ofagent-go/internal/pkg/openflow"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/voltha"
	"google.golang.org/grpc"
	"sync"
	"time"
)

var grpcDeviceID = "GRPC_CLIENT"
var client voltha.VolthaServiceClient
var packetOutChannel chan voltha.PacketOut
var clientMap map[string]*openflow.OFClient
var ofControllerEndPoint string
var mapLock sync.Mutex
var logger, _ = log.AddPackage(log.JSON, log.DebugLevel, nil)

//StartClient - make the inital connection to voltha and kicks off io streams
func StartClient(endpointAddress string, endpointPort uint16, inOFControllerEndPoint string) {

	if logger.V(log.DebugLevel) {
		logger.Debugw("Starting GRPC - VOLTHA client",
			log.Fields{
				"voltha-address":      endpointAddress,
				"voltha-port":         endpointPort,
				"controller-endpoint": inOFControllerEndPoint})
	}
	ofControllerEndPoint = inOFControllerEndPoint
	clientMap = make(map[string]*openflow.OFClient)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	endpoint := fmt.Sprintf("%s:%d", endpointAddress, endpointPort)
	conn, err := grpc.Dial(endpoint, opts...)
	defer conn.Close()
	if err != nil {
		log.Fatalw("StartClient failed opening GRPC connecton", log.Fields{"EndPoint": endpoint, "Error": err})
	}
	client = voltha.NewVolthaServiceClient(conn)
	packetOutChannel = make(chan voltha.PacketOut)

	go receivePacketIn(client)
	go receiveChangeEvent(client)
	go streamPacketOut(client)

	for {
		if logger.V(log.DebugLevel) {
			logger.Debugln("GrpcClient entering device refresh pull")
		}
		deviceList, err := client.ListLogicalDevices(context.Background(), &empty.Empty{})
		if err != nil {
			logger.Errorw("GrpcClient getDeviceList failed", log.Fields{"Error": err})
		}
		devices := deviceList.GetItems()
		refreshDeviceList(devices)
		time.Sleep(time.Minute)
	}
}
func refreshDeviceList(devices []*voltha.LogicalDevice) {
	//first find the new ones

	var toAdd []string
	var toDel []string
	var deviceIDMap = make(map[string]string)
	for i := 0; i < len(devices); i++ {
		deviceID := devices[i].GetId()
		deviceIDMap[deviceID] = deviceID
		if clientMap[deviceID] == nil {
			toAdd = append(toAdd, deviceID)
		}
	}
	for key := range clientMap {
		if deviceIDMap[key] == "" {
			toDel = append(toDel, key)
		}
	}
	if logger.V(log.DebugLevel) {
		logger.Debugw("GrpcClient refreshDeviceList", log.Fields{"ToAdd": toAdd, "ToDel": toDel})
	}
	for i := 0; i < len(toAdd); i++ {
		var client = addClient(toAdd[i])
		go client.Start()
	}
	for i := 0; i < len(toDel); i++ {
		clientMap[toDel[i]].End()
		mapLock.Lock()
		delete(clientMap, toDel[i])
		mapLock.Unlock()
	}
}
func addClient(deviceID string) *openflow.OFClient {
	if logger.V(log.DebugLevel) {
		logger.Debugw("GrpcClient addClient called ", log.Fields{"DeviceID": deviceID})
	}
	mapLock.Lock()
	ofc := clientMap[deviceID]
	if ofc == nil {
		ofc = openflow.NewOFClient(deviceID, ofControllerEndPoint, client, packetOutChannel, true)
		go ofc.Start()
		clientMap[deviceID] = ofc
	}
	mapLock.Unlock()
	logger.Debugw("Finished with addClient", log.Fields{"deviceID": deviceID})
	return ofc
}

//GetClient Returns a pointer to the OpenFlow client
func GetClient(deviceID string) *openflow.OFClient {
	client := clientMap[deviceID]
	if client == nil {
		client = addClient(deviceID)
	}
	return client
}
