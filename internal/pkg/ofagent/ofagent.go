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
	ofaEventStart = ofaEvent(iota)
	ofaEventVolthaConnected
	ofaEventVolthaDisconnected
	ofaEventError
)

type OFAgent struct {
	VolthaApiEndPoint         string
	OFControllerEndPoint      string
	DeviceListRefreshInterval time.Duration

	volthaConnection *grpc.ClientConn
	volthaClient     voltha.VolthaServiceClient
	mapLock          sync.Mutex
	clientMap        map[string]*openflow.OFClient
	events           chan ofaEvent

	packetInChannel    chan *voltha.PacketIn
	packetOutChannel   chan *voltha.PacketOut
	changeEventChannel chan *voltha.ChangeEvent
}

func NewOFAgent(config *OFAgent) (*OFAgent, error) {
	ofa := OFAgent{
		VolthaApiEndPoint:         config.VolthaApiEndPoint,
		OFControllerEndPoint:      config.OFControllerEndPoint,
		DeviceListRefreshInterval: config.DeviceListRefreshInterval,
		packetInChannel:           make(chan *voltha.PacketIn),
		packetOutChannel:          make(chan *voltha.PacketOut),
		changeEventChannel:        make(chan *voltha.ChangeEvent),
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
	ofa.events <- ofaEventStart
	ofa.events <- ofaEventVolthaConnected

	/*
	 * Two sub-contexts are created here for different purposes so we can
	 * control the lifecyle of processing loops differently.
	 *
	 * volthaCtx -  controls those processes that rely on the GRPC
	 *              GRPCconnection to voltha and will be restarted when the
	 *              GRPC connection is interrupted.
	 * hdlCtx    -  controls those processes that listen to channels and
	 *              process each message. these will likely never be
	 *              stopped until the ofagent is stopped.
	 */
	var volthaCtx, hdlCtx context.Context
	var volthaDone, hdlDone func()
	defer func() {
		if volthaDone != nil {
			volthaDone()
			volthaDone = nil
		}
		if hdlDone != nil {
			hdlDone()
			hdlDone = nil
		}
	}()

	for {
		select {
		case <-ctx.Done():
			if volthaDone != nil {
				volthaDone()
				volthaDone = nil
			}
			if hdlDone != nil {
				hdlDone()
				hdlDone = nil
			}
			return
		case event := <-ofa.events:
			switch event {
			case ofaEventStart:
				logger.Debug("ofagent-voltha-connect-event")

				volthaCtx, volthaDone = context.WithCancel(context.Background())
				// Start the loops that poll from voltha
				go ofa.receiveChangeEvents(volthaCtx)
				go ofa.receivePacketsIn(volthaCtx)
				/*
				 * No reciever for packet out as those come
				 * from the open flow client per OLT
				 */

			case ofaEventVolthaConnected:
				logger.Debug("ofagent-voltha-disconnect-event")

				hdlCtx, hdlDone = context.WithCancel(context.Background())
				// Start the loops that process messages
				go ofa.handlePacketsIn(hdlCtx)
				go ofa.handleChangeEvents(hdlCtx)
				go ofa.streamPacketOut(hdlCtx)
				go ofa.synchronizeDeviceList(hdlCtx)

			case ofaEventVolthaDisconnected:
				logger.Debug("ofagent-voltha-disconnect-event")
				volthaDone()
				volthaDone = nil
				volthaCtx = nil
			case ofaEventError:
				logger.Debug("ofagent-error-event")
			default:
				logger.Fatalw("ofagent-unknown-event",
					log.Fields{"event": event})
			}
		}
	}
}
