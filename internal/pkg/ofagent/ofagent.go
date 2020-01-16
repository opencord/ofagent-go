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
