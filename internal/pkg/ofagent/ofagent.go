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
	"github.com/opencord/ofagent-go/internal/pkg/openflow"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/voltha"
	"google.golang.org/grpc"
	"sync"
	"time"
)

var logger, _ = log.AddPackage(log.JSON, log.DebugLevel, nil)

type ofaEvent byte

const (
	ofaEventConnected = ofaEvent(iota)
	ofaEventDisconnected
	ofaEventError
)

type OFAgent struct {
	VolthaApiEndPoint         string
	OFControllerEndPoint      string
	DeviceListRefreshInterval time.Duration

	volthaConnection *grpc.ClientConn
	volthaClient     voltha.VolthaServiceClient
	packetOutChannel chan voltha.PacketOut
	mapLock          sync.Mutex
	clientMap        map[string]*openflow.OFClient
	events           chan ofaEvent
}

func NewOFAgent(config *OFAgent) (*OFAgent, error) {
	ofa := OFAgent{
		VolthaApiEndPoint:         config.VolthaApiEndPoint,
		OFControllerEndPoint:      config.OFControllerEndPoint,
		DeviceListRefreshInterval: config.DeviceListRefreshInterval,
		packetOutChannel:          make(chan voltha.PacketOut),
		clientMap:                 make(map[string]*openflow.OFClient),
		events:                    make(chan ofaEvent, 10),
	}

	if ofa.DeviceListRefreshInterval <= 0 {
		logger.Warnw("device list refresh internal not valid, setting to default",
			log.Fields{
				"value":   ofa.DeviceListRefreshInterval.String(),
				"default": (1 * time.Minute).String()})
		ofa.DeviceListRefreshInterval = 1 * time.Minute
	}

	return &ofa, nil
}

// Run - make the inital connection to voltha and kicks off io streams
func (ofa *OFAgent) Run(ctx context.Context) {

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
	ofa.events <- ofaEventConnected

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case event := <-ofa.events:
			switch event {
			case ofaEventConnected:
				log.Debug("ofagent-connect-event")
				go ofa.receivePacketIn()
				go ofa.receiveChangeEvent()
				go ofa.streamPacketOut()
				go ofa.synchronizeDeviceList(context.Background())

			case ofaEventDisconnected:
			case ofaEventError:
			default:
			}
		}
	}
}
