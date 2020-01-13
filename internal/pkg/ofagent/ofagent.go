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

package ofagent

import (
	"context"
	"errors"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/ofagent-go/internal/pkg/openflow"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-lib-go/v2/pkg/probe"
	pb "github.com/opencord/voltha-protos/v2/go/voltha"
	"google.golang.org/grpc"
	"sync"
	"time"
)

var grpcDeviceID = "GRPC_CLIENT"
var client pb.VolthaServiceClient
var clientMap map[string]*openflow.Client
var ofAddress string
var ofPort uint16
var mapLock sync.Mutex

var logger, _ = log.AddPackage(log.JSON, log.DebugLevel, nil)

type OFAgent struct {
	volthaEndPoint       string
	ofControllerEndPoint string
	connectionMaxRetries int
	connectionRetryDelay time.Duration
	volthaConnection     *grpc.ClientConn
	volthaService        pb.VolthaServiceClient
	clientMap            map[string]*openflow.Client
}

func NewOFAgent(ofControllerEndPoint, volthaEndPoint string) (*OFAgent, error) {
	o := OFAgent{
		volthaEndPoint:       volthaEndPoint,
		ofControllerEndPoint: ofControllerEndPoint,
	}

	o.connectionRetryDelay = 3 * time.Second

	clientMap = make(map[string]*openflow.Client)
	return &o, nil
}

func (o *OFAgent) establishConnectionToVoltha(p *probe.Probe) error {
	if p != nil {
		p.UpdateStatus("voltha", probe.ServiceStatusPreparing)
	}
	o.volthaConnection = nil
	try := 1
	for o.connectionMaxRetries == 0 || try < o.connectionMaxRetries {
		conn, err := grpc.Dial(o.volthaEndPoint, grpc.WithInsecure())
		if err != nil {
			log.Warnw("Failed to connect to voltha",
				log.Fields{
					"volthaEndPoint": o.volthaEndPoint,
					"error":          err.Error(),
				})
			if o.connectionMaxRetries == 0 || try < o.connectionMaxRetries {
				if o.connectionMaxRetries != 0 {
					try += 1
				}
				time.Sleep(o.connectionRetryDelay)
			}
		} else {
			log.Debugw("Established connection to Voltha",
				log.Fields{
					"volthaEndPoint": o.volthaEndPoint,
				})
			o.volthaConnection = conn
			o.volthaService = pb.NewVolthaServiceClient(conn)
			p.UpdateStatus("voltha", probe.ServiceStatusRunning)
			return nil
		}
	}
	if p != nil {
		p.UpdateStatus("voltha", probe.ServiceStatusFailed)
	}
	return errors.New("failed-to-connect-to-voltha")
}

// Start initiates connections to the OF controller as well as the voltha
// system. This method does not return until the OFAgent is stopped.
func (o *OFAgent) Run(ctx context.Context) {
	log.Debugw("Starting OFAgent",
		log.Fields{
			"volthaEndPoint":       o.volthaEndPoint,
			"ofControllerEndPoint": o.ofControllerEndPoint,
		})

	// If the context contains a probe then register services
	p := probe.GetProbeFromContext(ctx)
	if p != nil {
		p.RegisterService("voltha")
	}

	// Establish connection to Voltha
	o.establishConnectionToVoltha(p)

	go receivePacketIn(client)
	go receiveChangeEvent(client)
	go streamPacketOut(client)

	openflow.SetGrpcClient(&client)
	for {
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			log.Debugln("GrpcClient entering device refresh pull")
		}
		deviceList, err := client.ListLogicalDevices(context.Background(), &empty.Empty{})
		if err != nil {
			log.Errorw("GrpcClient getDeviceList failed", log.Fields{"Error": err})
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
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		log.Debugw("GrpcClient refreshDeviceList", log.Fields{"ToAdd": toAdd, "ToDel": toDel})
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
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		log.Debugw("GrpcClient addClient called ", log.Fields{"DeviceID": deviceID})
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
	log.Debugw("Finished with addClient", log.Fields{"deviceID": deviceID})
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
