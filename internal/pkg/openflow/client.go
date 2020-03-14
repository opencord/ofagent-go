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

package openflow

import (
	"context"
	"errors"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-protos/v3/go/voltha"
	"sync"
	"time"
)

var NoVolthaConnectionError = errors.New("no-voltha-connection")

type ofcEvent byte
type ofcState byte
type ofcRole byte

const (
	ofcEventStart = ofcEvent(iota)
	ofcEventConnect
	ofcEventDisconnect
	ofcEventStop

	ofcStateCreated = ofcState(iota)
	ofcStateStarted
	ofcStateConnected
	ofcStateDisconnected
	ofcStateStopped

	ofcRoleNone = ofcRole(iota)
	ofcRoleEqual
	ofcRoleMaster
	ofcRoleSlave
)

func (e ofcEvent) String() string {
	switch e {
	case ofcEventStart:
		return "ofc-event-start"
	case ofcEventConnect:
		return "ofc-event-connected"
	case ofcEventDisconnect:
		return "ofc-event-disconnected"
	case ofcEventStop:
		return "ofc-event-stop"
	default:
		return "ofc-event-unknown"
	}
}

func (s ofcState) String() string {
	switch s {
	case ofcStateCreated:
		return "ofc-state-created"
	case ofcStateStarted:
		return "ofc-state-started"
	case ofcStateConnected:
		return "ofc-state-connected"
	case ofcStateDisconnected:
		return "ofc-state-disconnected"
	case ofcStateStopped:
		return "ofc-state-stopped"
	default:
		return "ofc-state-unknown"
	}
}

// OFClient the configuration and operational state of a connection to an
// openflow controller
type OFClient struct {
	OFControllerEndPoints []string
	DeviceID              string
	VolthaClient          voltha.VolthaServiceClient
	PacketOutChannel      chan *voltha.PacketOut
	ConnectionMaxRetries  int
	ConnectionRetryDelay  time.Duration

	// map of endpoint to OF connection
	connections map[string]*OFConnection

	// global role state for device
	generationIsDefined bool
	generationID        uint64
	roleLock            sync.Mutex
}

type RoleManager interface {
	UpdateRoles(from string, request *ofp.RoleRequest) bool
}

func distance(a uint64, b uint64) int64 {
	return (int64)(a - b)
}

// UpdateRoles validates a role request and updates role state for connections where it changed
func (ofc *OFClient) UpdateRoles(from string, request *ofp.RoleRequest) bool {
	log.Debug("updating role", log.Fields{
		"from": from,
		"to":   request.Role,
		"id":   request.GenerationId})

	ofc.roleLock.Lock()
	defer ofc.roleLock.Unlock()

	if request.Role == ofp.OFPCRRoleEqual {
		// equal request doesn't care about generation ID and always succeeds
		connection := ofc.connections[from]
		connection.role = ofcRoleEqual
		return true
	}

	if ofc.generationIsDefined && distance(request.GenerationId, ofc.generationID) < 0 {
		// generation ID is not valid
		return false
	} else {
		ofc.generationID = request.GenerationId
		ofc.generationIsDefined = true

		if request.Role == ofp.OFPCRRoleMaster {
			// master is potentially changing, find the existing master and set it to slave
			for endpoint, connection := range ofc.connections {
				if endpoint == from {
					connection.role = ofcRoleMaster
				} else if connection.role == ofcRoleMaster {
					// the old master should be set to slave
					connection.role = ofcRoleSlave
				}
			}
			return true
		} else if request.Role == ofp.OFPCRRoleSlave {
			connection := ofc.connections[from]
			connection.role = ofcRoleSlave
			return true
		}
	}

	return false
}

// NewClient returns an initialized OFClient instance based on the configuration
// specified
func NewOFClient(config *OFClient) *OFClient {

	ofc := OFClient{
		DeviceID:              config.DeviceID,
		OFControllerEndPoints: config.OFControllerEndPoints,
		VolthaClient:          config.VolthaClient,
		PacketOutChannel:      config.PacketOutChannel,
		ConnectionMaxRetries:  config.ConnectionMaxRetries,
		ConnectionRetryDelay:  config.ConnectionRetryDelay,
		connections:           make(map[string]*OFConnection),
	}

	if ofc.ConnectionRetryDelay <= 0 {
		logger.Warnw("connection retry delay not valid, setting to default",
			log.Fields{
				"device-id": ofc.DeviceID,
				"value":     ofc.ConnectionRetryDelay.String(),
				"default":   (3 * time.Second).String()})
		ofc.ConnectionRetryDelay = 3 * time.Second
	}
	return &ofc
}

// Stop initiates a shutdown of the OFClient
func (ofc *OFClient) Stop() {
	for _, connection := range ofc.connections {
		connection.events <- ofcEventStop
	}
}

func (ofc *OFClient) Run(ctx context.Context) {

	for _, endpoint := range ofc.OFControllerEndPoints {
		connection := &OFConnection{
			OFControllerEndPoint: endpoint,
			DeviceID:             ofc.DeviceID,
			VolthaClient:         ofc.VolthaClient,
			PacketOutChannel:     ofc.PacketOutChannel,
			ConnectionMaxRetries: ofc.ConnectionMaxRetries,
			ConnectionRetryDelay: ofc.ConnectionRetryDelay,
			role:                 ofcRoleNone,
			roleManager:          ofc,
			events:               make(chan ofcEvent, 10),
			sendChannel:          make(chan Message, 100),
		}

		ofc.connections[endpoint] = connection
	}

	for _, connection := range ofc.connections {
		go connection.Run(ctx)
	}
}

func (ofc *OFClient) SendMessage(message Message) error {
	for _, connection := range ofc.connections {
		if connection.role == ofcRoleMaster || connection.role == ofcRoleEqual {
			err := connection.SendMessage(message)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
