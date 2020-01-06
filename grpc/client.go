/*
   Copyright 2017 the original author or authors.

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

package grpc

import (
	"context"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/opencord/ofagent-go/openflow"

	"fmt"

	"github.com/opencord/ofagent-go/settings"
	l "github.com/opencord/voltha-lib-go/v2/pkg/log"

	pb "github.com/opencord/voltha-protos/v2/go/voltha"
	"google.golang.org/grpc"
)

var grpcDeviceID = "GRPC_CLIENT"
var client pb.VolthaServiceClient
var clientMap map[string]*openflow.Client
var ofAddress string
var ofPort uint16
var mapLock sync.Mutex
var logger, _ = l.AddPackage(l.JSON, l.DebugLevel, nil)

//StartClient - make the inital connection to voltha and kicks off io streams
func StartClient(endpointAddress string, endpointPort uint16, openFlowAddress string, openFlowPort uint16) {

	if settings.GetDebug(grpcDeviceID) {
		logger.Debugw("Starting GRPC - VOLTHA client", l.Fields{"EndPointAddr": endpointAddress,
			"EndPointPort": endpointPort, "OpenFlowAddress": openFlowAddress, "OpenFlowPort": openFlowPort})
	}
	ofAddress = openFlowAddress
	ofPort = openFlowPort
	clientMap = make(map[string]*openflow.Client)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	endpoint := fmt.Sprintf("%s:%d", endpointAddress, endpointPort)
	conn, err := grpc.Dial(endpoint, opts...)
	defer conn.Close()
	if err != nil {
		l.Fatalw("StartClient failed opening GRPC connecton", l.Fields{"EndPoint": endpoint, "Error": err})
	}
	client = pb.NewVolthaServiceClient(conn)

	go receivePacketIn(client)
	go receiveChangeEvent(client)
	go streamPacketOut(client)

	openflow.SetGrpcClient(&client)
	for {
		if settings.GetDebug(grpcDeviceID) {
			logger.Debugln("GrpcClient entering device refresh pull")
		}
		deviceList, err := client.ListLogicalDevices(context.Background(), &empty.Empty{})
		if err != nil {
			logger.Errorw("GrpcClient getDeviceList failed", l.Fields{"Error": err})
		}
		devices := deviceList.GetItems()
		refreshDeviceList(devices)
		time.Sleep(time.Minute)
	}
}
func refreshDeviceList(devices []*pb.LogicalDevice) {
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
	if settings.GetDebug(grpcDeviceID) {
		logger.Debugw("GrpcClient refreshDeviceList", l.Fields{"ToAdd": toAdd, "ToDel": toDel})
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
func addClient(deviceID string) *openflow.Client {
	if settings.GetDebug(grpcDeviceID) {
		logger.Debugw("GrpcClient addClient called ", l.Fields{"DeviceID": deviceID})
	}
	mapLock.Lock()
	var client *openflow.Client
	client = clientMap[deviceID]
	if client == nil {
		client = openflow.NewClient(ofAddress, ofPort, deviceID, true)
		go client.Start()
		clientMap[deviceID] = client
	}
	mapLock.Unlock()
	logger.Debugw("Finished with addClient", l.Fields{"deviceID": deviceID})
	return client
}

//GetClient Returns a pointer to the OpenFlow client
func GetClient(deviceID string) *openflow.Client {
	client := clientMap[deviceID]
	if client == nil {
		client = addClient(deviceID)
	}
	return client
}
