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
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"time"

	"github.com/donNewtonAlpha/goloxi"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/ofagent-go/internal/pkg/holder"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-protos/v3/go/voltha"
)

var NoVolthaConnectionError = errors.New("no-voltha-connection")

type ofcEvent byte
type ofcState byte

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
	OFControllerEndPoint string
	Port                 uint16
	DeviceID             string
	VolthaClient         *holder.VolthaServiceClientHolder
	PacketOutChannel     chan *voltha.PacketOut
	ConnectionMaxRetries int
	ConnectionRetryDelay time.Duration
	conn                 net.Conn

	// expirimental
	events            chan ofcEvent
	sendChannel       chan Message
	lastUnsentMessage Message
}

// NewClient returns an initialized OFClient instance based on the configuration
// specified
func NewOFClient(config *OFClient) *OFClient {

	ofc := OFClient{
		DeviceID:             config.DeviceID,
		OFControllerEndPoint: config.OFControllerEndPoint,
		VolthaClient:         config.VolthaClient,
		PacketOutChannel:     config.PacketOutChannel,
		ConnectionMaxRetries: config.ConnectionMaxRetries,
		ConnectionRetryDelay: config.ConnectionRetryDelay,
		events:               make(chan ofcEvent, 10),
		sendChannel:          make(chan Message, 100),
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
	ofc.events <- ofcEventStop
}

func (ofc *OFClient) peekAtOFHeader(buf []byte) (ofp.IHeader, error) {
	header := ofp.Header{}
	header.Version = uint8(buf[0])
	header.Type = uint8(buf[1])
	header.Length = binary.BigEndian.Uint16(buf[2:4])
	header.Xid = binary.BigEndian.Uint32(buf[4:8])

	// TODO: add minimal validation of version and type

	return &header, nil
}

func (ofc *OFClient) establishConnectionToController() error {
	if ofc.conn != nil {
		logger.Debugw("closing-of-connection-to-reconnect",
			log.Fields{"device-id": ofc.DeviceID})
		ofc.conn.Close()
		ofc.conn = nil
	}
	try := 1
	for ofc.ConnectionMaxRetries == 0 || try < ofc.ConnectionMaxRetries {
		if raddr, err := net.ResolveTCPAddr("tcp", ofc.OFControllerEndPoint); err != nil {
			logger.Debugw("openflow-client unable to resolve endpoint",
				log.Fields{
					"device-id": ofc.DeviceID,
					"endpoint":  ofc.OFControllerEndPoint})
		} else {
			if connection, err := net.DialTCP("tcp", nil, raddr); err == nil {
				ofc.conn = connection
				ofc.sayHello()
				ofc.events <- ofcEventConnect
				return nil
			} else {
				logger.Warnw("openflow-client-connect-error",
					log.Fields{
						"device-id": ofc.DeviceID,
						"endpoint":  ofc.OFControllerEndPoint})
			}
		}
		if ofc.ConnectionMaxRetries == 0 || try < ofc.ConnectionMaxRetries {
			if ofc.ConnectionMaxRetries != 0 {
				try += 1
			}
			time.Sleep(ofc.ConnectionRetryDelay)
		}
	}
	return errors.New("failed-to-connect-to-of-controller")
}

// Run implementes the state machine for the OF client reacting to state change
// events and invoking actions as a reaction to those state changes
func (ofc *OFClient) Run(ctx context.Context) {

	var ofCtx context.Context
	var ofDone func()
	state := ofcStateCreated
	ofc.events <- ofcEventStart
top:
	for {
		select {
		case <-ctx.Done():
			state = ofcStateStopped
			logger.Debugw("state-transition-context-done",
				log.Fields{"device-id": ofc.DeviceID})
			break top
		case event := <-ofc.events:
			previous := state
			switch event {
			case ofcEventStart:
				logger.Debugw("ofc-event-start",
					log.Fields{"device-id": ofc.DeviceID})
				if state == ofcStateCreated {
					state = ofcStateStarted
					logger.Debug("STARTED MORE THAN ONCE")
					go func() {
						if err := ofc.establishConnectionToController(); err != nil {
							logger.Errorw("controller-connection-failed", log.Fields{"error": err})
							panic(err)
						}
					}()
				} else {
					logger.Errorw("illegal-state-transition",
						log.Fields{
							"device-id":     ofc.DeviceID,
							"current-state": state.String(),
							"event":         event.String()})
				}
			case ofcEventConnect:
				logger.Debugw("ofc-event-connected",
					log.Fields{"device-id": ofc.DeviceID})
				if state == ofcStateStarted || state == ofcStateDisconnected {
					state = ofcStateConnected
					ofCtx, ofDone = context.WithCancel(context.Background())
					go ofc.messageSender(ofCtx)
					go ofc.processOFStream(ofCtx)
				} else {
					logger.Errorw("illegal-state-transition",
						log.Fields{
							"device-id":     ofc.DeviceID,
							"current-state": state.String(),
							"event":         event.String()})
				}
			case ofcEventDisconnect:
				logger.Debugw("ofc-event-disconnected",
					log.Fields{
						"device-id": ofc.DeviceID,
						"state":     state.String()})
				if state == ofcStateConnected {
					state = ofcStateDisconnected
					if ofDone != nil {
						ofDone()
						ofDone = nil
					}
					go func() {
						if err := ofc.establishConnectionToController(); err != nil {
							logger.Errorw("controller-connection-failed", log.Fields{"error": err})
							panic(err)
						}
					}()
				} else {
					logger.Errorw("illegal-state-transition",
						log.Fields{
							"device-id":     ofc.DeviceID,
							"current-state": state.String(),
							"event":         event.String()})
				}
			case ofcEventStop:
				logger.Debugw("ofc-event-stop",
					log.Fields{"device-id": ofc.DeviceID})
				if state == ofcStateCreated || state == ofcStateConnected || state == ofcStateDisconnected {
					state = ofcStateStopped
					break top
				} else {
					logger.Errorw("illegal-state-transition",
						log.Fields{
							"device-id":     ofc.DeviceID,
							"current-state": state.String(),
							"event":         event.String()})
				}
			}
			logger.Debugw("state-transition",
				log.Fields{
					"device-id":      ofc.DeviceID,
					"previous-state": previous.String(),
					"current-state":  state.String(),
					"event":          event.String()})
		}
	}

	// If the child context exists, then cancel it
	if ofDone != nil {
		logger.Debugw("closing-child-processes",
			log.Fields{"device-id": ofc.DeviceID})
		ofDone()
	}

	// If the connection is open, then close it
	if ofc.conn != nil {
		logger.Debugw("closing-of-connection",
			log.Fields{"device-id": ofc.DeviceID})
		ofc.conn.Close()
		ofc.conn = nil
	}
	logger.Debugw("state-machine-finished",
		log.Fields{"device-id": ofc.DeviceID})
}

// processOFStream processes the OF connection from the controller and invokes
// the appropriate handler methods for each message.
func (ofc *OFClient) processOFStream(ctx context.Context) {
	fromController := bufio.NewReader(ofc.conn)

	/*
	 * We have a read buffer of a max size of 4096, so if we ever have
	 * a message larger than this then we will have issues
	 */
	headerBuf := make([]byte, 8)

top:
	// Continue until we are told to stop
	for {
		select {
		case <-ctx.Done():
			logger.Error("of-loop-ending-context-done")
			break top
		default:
			// Read 8 bytes, the standard OF header
			read, err := io.ReadFull(fromController, headerBuf)
			if err != nil {
				logger.Errorw("bad-of-header",
					log.Fields{
						"byte-count": read,
						"device-id":  ofc.DeviceID,
						"error":      err})
				break top
			}

			// Decode the header
			peek, err := ofc.peekAtOFHeader(headerBuf)
			if err != nil {
				/*
				 * Header is bad, assume stream is corrupted
				 * and needs to be restarted
				 */
				logger.Errorw("bad-of-packet",
					log.Fields{
						"device-id": ofc.DeviceID,
						"error":     err})
				break top
			}

			// Calculate the size of the rest of the packet and read it
			need := int(peek.GetLength())
			messageBuf := make([]byte, need)
			copy(messageBuf, headerBuf)
			read, err = io.ReadFull(fromController, messageBuf[8:])
			if err != nil {
				logger.Errorw("bad-of-packet",
					log.Fields{
						"byte-count": read,
						"device-id":  ofc.DeviceID,
						"error":      err})
				break top
			}

			// Decode and process the packet
			decoder := goloxi.NewDecoder(messageBuf)
			msg, err := ofp.DecodeHeader(decoder)
			if err != nil {
				// nolint: staticcheck
				js, _ := json.Marshal(decoder)
				logger.Errorw("failed-to-decode",
					log.Fields{
						"device-id": ofc.DeviceID,
						"decoder":   js,
						"error":     err})
				break top
			}
			if logger.V(log.DebugLevel) {
				js, _ := json.Marshal(msg)
				logger.Debugw("packet-header",
					log.Fields{
						"device-id": ofc.DeviceID,
						"header":    js})
			}
			ofc.parseHeader(msg)
		}
	}
	logger.Debugw("end-of-stream",
		log.Fields{"device-id": ofc.DeviceID})
	ofc.events <- ofcEventDisconnect
}

func (ofc *OFClient) sayHello() {
	hello := ofp.NewHello()
	hello.Xid = uint32(GetXid())
	elem := ofp.NewHelloElemVersionbitmap()
	elem.SetType(ofp.OFPHETVersionbitmap)
	elem.SetLength(8)
	elem.SetBitmaps([]*ofp.Uint32{{Value: 16}})
	hello.SetElements([]ofp.IHelloElem{elem})
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(hello)
		logger.Debugw("sayHello Called",
			log.Fields{
				"device-id":     ofc.DeviceID,
				"hello-message": js})
	}
	if err := ofc.SendMessage(hello); err != nil {
		logger.Fatalw("Failed saying hello to Openflow Server, unable to proceed",
			log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
	}
}

func (ofc *OFClient) parseHeader(header ofp.IHeader) {
	headerType := header.GetType()
	logger.Debugw("packet-header-type",
		log.Fields{
			"header-type": ofp.Type(headerType).String()})
	switch headerType {
	case ofp.OFPTHello:
		//x := header.(*ofp.Hello)
	case ofp.OFPTError:
		go ofc.handleErrMsg(header.(*ofp.ErrorMsg))
	case ofp.OFPTEchoRequest:
		go ofc.handleEchoRequest(header.(*ofp.EchoRequest))
	case ofp.OFPTEchoReply:
	case ofp.OFPTExperimenter:
	case ofp.OFPTFeaturesRequest:
		go func() {
			if err := ofc.handleFeatureRequest(header.(*ofp.FeaturesRequest)); err != nil {
				logger.Errorw("handle-feature-request", log.Fields{"error": err})
			}
		}()
	case ofp.OFPTFeaturesReply:
	case ofp.OFPTGetConfigRequest:
		go ofc.handleGetConfigRequest(header.(*ofp.GetConfigRequest))
	case ofp.OFPTGetConfigReply:
	case ofp.OFPTSetConfig:
		go ofc.handleSetConfig(header.(*ofp.SetConfig))
	case ofp.OFPTPacketIn:
	case ofp.OFPTFlowRemoved:
	case ofp.OFPTPortStatus:
	case ofp.OFPTPacketOut:
		go ofc.handlePacketOut(header.(*ofp.PacketOut))
	case ofp.OFPTFlowMod:
		/*
		 * Not using go routine to handle flow* messages or barrier requests
		 * onos typically issues barrier requests just before a flow* message.
		 * by handling in this thread I ensure all flow* are handled when barrier
		 * request is issued.
		 */
		switch header.(ofp.IFlowMod).GetCommand() {
		case ofp.OFPFCAdd:
			ofc.handleFlowAdd(header.(*ofp.FlowAdd))
		case ofp.OFPFCModify:
			ofc.handleFlowMod(header.(*ofp.FlowMod))
		case ofp.OFPFCModifyStrict:
			ofc.handleFlowModStrict(header.(*ofp.FlowModifyStrict))
		case ofp.OFPFCDelete:
			ofc.handleFlowDelete(header.(*ofp.FlowDelete))
		case ofp.OFPFCDeleteStrict:
			ofc.handleFlowDeleteStrict(header.(*ofp.FlowDeleteStrict))
		}
	case ofp.OFPTStatsRequest:
		go func() {
			if err := ofc.handleStatsRequest(header, header.(ofp.IStatsRequest).GetStatsType()); err != nil {
				logger.Errorw("ofpt-stats-request", log.Fields{"error": err})
			}
		}()
	case ofp.OFPTBarrierRequest:
		/* See note above at case ofp.OFPTFlowMod:*/
		ofc.handleBarrierRequest(header.(*ofp.BarrierRequest))
	case ofp.OFPTRoleRequest:
		go ofc.handleRoleRequest(header.(*ofp.RoleRequest))
	case ofp.OFPTMeterMod:
		ofc.handleMeterModRequest(header.(*ofp.MeterMod))
	}
}

// Message interface that represents an open flow message and enables for a
// unified implementation of SendMessage
type Message interface {
	Serialize(encoder *goloxi.Encoder) error
}

func (ofc *OFClient) doSend(msg Message) error {
	if ofc.conn == nil {
		return errors.New("no-connection")
	}
	enc := goloxi.NewEncoder()
	if err := msg.Serialize(enc); err != nil {
		return err
	}

	bytes := enc.Bytes()
	if _, err := ofc.conn.Write(bytes); err != nil {
		logger.Errorw("unable-to-send-message-to-controller",
			log.Fields{
				"device-id": ofc.DeviceID,
				"message":   msg,
				"error":     err})
		return err
	}
	return nil
}

func (ofc *OFClient) messageSender(ctx context.Context) {
	// first process last fail if it exists
	if ofc.lastUnsentMessage != nil {
		if err := ofc.doSend(ofc.lastUnsentMessage); err != nil {
			ofc.events <- ofcEventDisconnect
			return
		}
		ofc.lastUnsentMessage = nil
	}
top:
	for {
		select {
		case <-ctx.Done():
			break top
		case msg := <-ofc.sendChannel:
			if err := ofc.doSend(msg); err != nil {
				ofc.lastUnsentMessage = msg
				ofc.events <- ofcEventDisconnect
				logger.Debugw("message-sender-error",
					log.Fields{
						"device-id": ofc.DeviceID,
						"error":     err.Error()})
				break top
			}
			logger.Debugw("message-sender-send",
				log.Fields{
					"device-id": ofc.DeviceID})
			ofc.lastUnsentMessage = nil
		}
	}

	logger.Debugw("message-sender-finished",
		log.Fields{
			"device-id": ofc.DeviceID})
}

// SendMessage queues a message to be sent to the openflow controller
func (ofc *OFClient) SendMessage(message Message) error {
	logger.Debug("queuing-message")
	ofc.sendChannel <- message
	return nil
}
