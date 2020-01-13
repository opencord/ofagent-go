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
	"github.com/opencord/ofagent-go/internal/pkg/openflow"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-lib-go/v2/pkg/probe"
	"github.com/opencord/voltha-protos/v2/go/voltha"
	"google.golang.org/grpc"
	"sync"
	"time"
)

var grpcDeviceID = "GRPC_CLIENT"

type ofaEvent int

const (
	ofaConnected    = ofaEvent(iota)
	ofaDisconnected = iota
)

type OFAgent struct {
	VolthaApiEndPoint    string
	OFControllerEndPoint string
	ConnectionMaxRetries int
	ConnectionRetryDelay time.Duration

	volthaConnection *grpc.ClientConn
	volthaClient     voltha.VolthaServiceClient
	packetOutChannel chan voltha.PacketOut
	mapLock          sync.Mutex
	clientMap        map[string]*openflow.OFClient
	events           chan ofaEvent
	ctx              context.Context
	done             func()
}

func NewOFAgent(in OFAgent) (*OFAgent, error) {
	o := OFAgent{
		VolthaApiEndPoint:    in.VolthaApiEndPoint,
		OFControllerEndPoint: in.OFControllerEndPoint,
		ConnectionMaxRetries: in.ConnectionMaxRetries,
		ConnectionRetryDelay: in.ConnectionRetryDelay,
	}

	o.packetOutChannel = make(chan voltha.PacketOut, 100)
	o.clientMap = make(map[string]*openflow.OFClient)
	return &o, nil
}

func (o *OFAgent) establishConnectionToVoltha(p *probe.Probe) error {
	if p != nil {
		p.UpdateStatus("voltha", probe.ServiceStatusPreparing)
	}

	if o.volthaConnection != nil {
		o.volthaConnection.Close()
	}

	o.volthaConnection = nil
	o.volthaClient = nil
	try := 1
	for o.ConnectionMaxRetries == 0 || try < o.ConnectionMaxRetries {
		conn, err := grpc.Dial(o.VolthaApiEndPoint, grpc.WithInsecure())
		if err == nil {
			svc := voltha.NewVolthaServiceClient(conn)
			if svc != nil {
				if _, err = svc.GetVoltha(context.Background(), nil); err == nil {
					log.Debugw("Established connection to Voltha",
						log.Fields{
							"VolthaApiEndPoint": o.VolthaApiEndPoint,
						})
					o.volthaConnection = conn
					o.volthaClient = svc
					p.UpdateStatus("voltha", probe.ServiceStatusRunning)
					return nil
				}
			}
		}
		log.Warnw("Failed to connect to voltha",
			log.Fields{
				"VolthaApiEndPoint": o.VolthaApiEndPoint,
				"error":             err.Error(),
			})
		if o.ConnectionMaxRetries == 0 || try < o.ConnectionMaxRetries {
			if o.ConnectionMaxRetries != 0 {
				try += 1
			}
			time.Sleep(o.ConnectionRetryDelay)
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
			"VolthaApiEndPoint":    o.VolthaApiEndPoint,
			"OFControllerEndPoint": o.OFControllerEndPoint,
		})

	// If the context contains a probe then register services
	p := probe.GetProbeFromContext(ctx)
	if p != nil {
		p.RegisterService("voltha")
	}

	// Establish connection to Voltha
	if err := o.establishConnectionToVoltha(p); err != nil {
		log.Fatal("Failed to establish connection to VOLTHA",
			log.Fields{"error": err})
	}

	// Create a context which  an be used to manage the lifecycle of
	// the child go routines
	var done func()
	var subCtx context.Context

loop:
	for {
		select {
		case <-ctx.Done():
			// Stop the children, if they have been started
			if done != nil {
				done()
				done = nil
			}
			// Close the packet out channel
			close(o.packetOutChannel)
			break loop
		case event := <-o.events:
			switch event {
			case ofaConnected:
				subCtx, done = context.WithCancel(context.Background())
				go o.receivePacketIn(subCtx)
				go o.receiveChangeEvent(subCtx)
				go o.streamPacketOut(subCtx)
				go o.refreshDeviceList(subCtx)
			case ofaDisconnected:
				// Stop the children, they will be restarted
				// on connect
				if done != nil {
					done()
				}
				done = nil
			default:
			}
		}
	}
}
