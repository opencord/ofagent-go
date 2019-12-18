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
	"log"

	pb "github.com/opencord/voltha-protos/v2/go/voltha"
	"google.golang.org/grpc"
)

var client pb.VolthaServiceClient
var clientMap map[string]*openflow.Client
var ofAddress string
var ofPort uint16
var mapLock sync.Mutex

func StartClient(endpointAddress string, endpointPort uint16, openFlowAddress string, openFlowPort uint16) {
	ofAddress = openFlowAddress
	ofPort = openFlowPort
	clientMap = make(map[string]*openflow.Client)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	endpoint := fmt.Sprintf("%s:%d", endpointAddress, endpointPort)
	conn, err := grpc.Dial(endpoint, opts...)
	defer conn.Close()
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	client = pb.NewVolthaServiceClient(conn)

	go receivePacketIn(client)
	go receiveChangeEvent(client)
	go streamPacketOut(client)

	openflow.SetGrpcClient(&client)
	for {
		log.Println("entering device refresh pull")
		deviceList, err := client.ListLogicalDevices(context.Background(), &empty.Empty{})
		if err != nil {
			log.Printf("ERROR GET DEVICE LIST %v", err)
		}
		devices := deviceList.GetItems()
		refreshDeviceList(devices)
		time.Sleep(time.Minute)
		log.Println("waking up")
	}
}
func refreshDeviceList(devices []*pb.LogicalDevice) {
	//first find the new ones

	var toAdd []string
	var toDel []string
	var deviceIdMap = make(map[string]string)
	for i := 0; i < len(devices); i++ {
		log.Printf("Device ID %s", devices[i].GetId())
		deviceId := devices[i].GetId()
		deviceIdMap[deviceId] = deviceId
		if clientMap[deviceId] == nil {
			toAdd = append(toAdd, deviceId)
		}
	}
	for key, _ := range clientMap {
		if deviceIdMap[key] == "" {
			toDel = append(toDel, key)
		}
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
func addClient(deviceId string) *openflow.Client {
	mapLock.Lock()
	var client *openflow.Client
	client = clientMap[deviceId]
	if client == nil {
		client = openflow.NewClient(ofAddress, ofPort, deviceId, true)
		clientMap[deviceId] = client
	}
	mapLock.Unlock()
	return client
}
func GetClient(deviceId string) *openflow.Client {
	return clientMap[deviceId]
}
