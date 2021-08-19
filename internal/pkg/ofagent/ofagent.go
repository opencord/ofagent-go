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
	"sync"
	"time"

	"github.com/opencord/ofagent-go/internal/pkg/holder"
	"github.com/opencord/ofagent-go/internal/pkg/openflow"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-lib-go/v7/pkg/probe"
	"github.com/opencord/voltha-protos/v5/go/openflow_13"
	"github.com/opencord/voltha-protos/v5/go/voltha"
	"google.golang.org/grpc"
)

type ofaEvent byte
type ofaState byte

const (
	ofaEventStart = ofaEvent(iota)
	ofaEventVolthaConnected
	ofaEventVolthaDisconnected
	ofaEventError

	ofaStateConnected = ofaState(iota)
	ofaStateConnecting
	ofaStateDisconnected
)

type OFAgent struct {
	VolthaApiEndPoint         string
	OFControllerEndPoints     []string
	DeviceListRefreshInterval time.Duration
	ConnectionMaxRetries      int
	ConnectionRetryDelay      time.Duration

	volthaConnection *grpc.ClientConn
	volthaClient     *holder.VolthaServiceClientHolder
	mapLock          sync.Mutex
	clientMap        map[string]*openflow.OFClient
	events           chan ofaEvent

	packetInChannel    chan *voltha.PacketIn
	packetOutChannel   chan *voltha.PacketOut
	changeEventChannel chan *openflow_13.ChangeEvent
}

func NewOFAgent(ctx context.Context, config *OFAgent) (*OFAgent, error) {
	ofa := OFAgent{
		VolthaApiEndPoint:         config.VolthaApiEndPoint,
		OFControllerEndPoints:     config.OFControllerEndPoints,
		DeviceListRefreshInterval: config.DeviceListRefreshInterval,
		ConnectionMaxRetries:      config.ConnectionMaxRetries,
		ConnectionRetryDelay:      config.ConnectionRetryDelay,
		volthaClient:              &holder.VolthaServiceClientHolder{},
		packetInChannel:           make(chan *voltha.PacketIn),
		packetOutChannel:          make(chan *voltha.PacketOut),
		changeEventChannel:        make(chan *openflow_13.ChangeEvent),
		clientMap:                 make(map[string]*openflow.OFClient),
		events:                    make(chan ofaEvent, 100),
	}

	if ofa.DeviceListRefreshInterval <= 0 {
		logger.Warnw(ctx, "device list refresh internal not valid, setting to default",
			log.Fields{
				"value":   ofa.DeviceListRefreshInterval.String(),
				"default": (1 * time.Minute).String()})
		ofa.DeviceListRefreshInterval = 1 * time.Minute
	}

	if ofa.ConnectionRetryDelay <= 0 {
		logger.Warnw(ctx, "connection retry delay not value, setting to default",
			log.Fields{
				"value":   ofa.ConnectionRetryDelay.String(),
				"default": (3 * time.Second).String()})
		ofa.ConnectionRetryDelay = 3 * time.Second
	}

	return &ofa, nil
}

// Run - make the inital connection to voltha and kicks off io streams
func (ofa *OFAgent) Run(ctx context.Context) {

	logger.Debugw(ctx, "Starting GRPC - VOLTHA client",
		log.Fields{
			"voltha-endpoint":     ofa.VolthaApiEndPoint,
			"controller-endpoint": ofa.OFControllerEndPoints})

	// If the context contains a k8s probe then register services
	p := probe.GetProbeFromContext(ctx)
	if p != nil {
		p.RegisterService(ctx, "voltha")
	}

	ofa.events <- ofaEventStart

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
	state := ofaStateDisconnected

	for {
		select {
		case <-ctx.Done():
			if volthaDone != nil {
				volthaDone()
			}
			if hdlDone != nil {
				hdlDone()
			}
			return
		case event := <-ofa.events:
			switch event {
			case ofaEventStart:
				logger.Debug(ctx, "ofagent-voltha-start-event")

				// Start the loops that process messages
				hdlCtx, hdlDone = context.WithCancel(context.Background())
				go ofa.handlePacketsIn(hdlCtx)
				go ofa.handleChangeEvents(hdlCtx)

				// Kick off process to attempt to establish
				// connection to voltha
				state = ofaStateConnecting
				go func() {
					if err := ofa.establishConnectionToVoltha(ctx, p); err != nil {
						logger.Errorw(ctx, "voltha-connection-failed", log.Fields{"error": err})
						panic(err)
					}
				}()

			case ofaEventVolthaConnected:
				logger.Debug(ctx, "ofagent-voltha-connect-event")

				// Start the loops that poll from voltha
				if state != ofaStateConnected {
					state = ofaStateConnected
					volthaCtx, volthaDone = context.WithCancel(context.Background())
					go ofa.receiveChangeEvents(volthaCtx)
					go ofa.receivePacketsIn(volthaCtx)
					go ofa.streamPacketOut(volthaCtx)
					go ofa.synchronizeDeviceList(volthaCtx)
				}

			case ofaEventVolthaDisconnected:
				if p != nil {
					p.UpdateStatus(ctx, "voltha", probe.ServiceStatusNotReady)
				}
				logger.Debug(ctx, "ofagent-voltha-disconnect-event")
				if state == ofaStateConnected {
					state = ofaStateDisconnected
					// Clear all the OF connections to the OF controller.  These will be recreated when the
					// connection to voltha is established
					ofa.clearAllOFClient()
					ofa.volthaClient.Clear()
					volthaDone()
					volthaDone = nil
				}
				if state != ofaStateConnecting {
					state = ofaStateConnecting
					go func() {
						if err := ofa.establishConnectionToVoltha(ctx, p); err != nil {
							logger.Errorw(ctx, "voltha-connection-failed", log.Fields{"error": err})
							panic(err)
						}
					}()
				}

			case ofaEventError:
				logger.Debug(ctx, "ofagent-error-event")
			default:
				logger.Fatalw(ctx, "ofagent-unknown-event",
					log.Fields{"event": event})
			}
		}
	}
}
