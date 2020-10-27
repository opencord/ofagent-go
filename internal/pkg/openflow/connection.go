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
	"sync"
	"time"

	"github.com/opencord/goloxi"
	ofp "github.com/opencord/goloxi/of13"
	"github.com/opencord/ofagent-go/internal/pkg/holder"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-protos/v3/go/voltha"
)

type OFConnection struct {
	OFControllerEndPoint string
	DeviceID             string
	VolthaClient         *holder.VolthaServiceClientHolder
	PacketOutChannel     chan *voltha.PacketOut
	ConnectionMaxRetries int
	ConnectionRetryDelay time.Duration

	conn net.Conn

	// current role of this connection
	role        ofcRole
	roleManager RoleManager

	events            chan ofcEvent
	sendChannel       chan Message
	lastUnsentMessage Message

	flowsChunkSize     int
	portsChunkSize     int
	portsDescChunkSize int
}

func (ofc *OFConnection) peekAtOFHeader(buf []byte) (ofp.IHeader, error) {
	header := ofp.Header{}
	header.Version = uint8(buf[0])
	header.Type = uint8(buf[1])
	header.Length = binary.BigEndian.Uint16(buf[2:4])
	header.Xid = binary.BigEndian.Uint32(buf[4:8])

	// TODO: add minimal validation of version and type

	return &header, nil
}

func (ofc *OFConnection) establishConnectionToController(ctx context.Context) error {
	if ofc.conn != nil {
		logger.Debugw(ctx, "closing-of-connection-to-reconnect",
			log.Fields{"device-id": ofc.DeviceID})
		ofc.conn.Close()
		ofc.conn = nil
	}
	try := 1
	for ofc.ConnectionMaxRetries == 0 || try < ofc.ConnectionMaxRetries {
		if raddr, err := net.ResolveTCPAddr("tcp", ofc.OFControllerEndPoint); err != nil {
			logger.Debugw(ctx, "openflow-client unable to resolve endpoint",
				log.Fields{
					"device-id": ofc.DeviceID,
					"endpoint":  ofc.OFControllerEndPoint})
		} else {
			if connection, err := net.DialTCP("tcp", nil, raddr); err == nil {
				ofc.conn = connection
				ofc.sayHello(ctx)
				ofc.events <- ofcEventConnect
				return nil
			} else {
				logger.Warnw(ctx, "openflow-client-connect-error",
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

// Run implements the state machine for the OF client reacting to state change
// events and invoking actions as a reaction to those state changes
func (ofc *OFConnection) Run(ctx context.Context) {

	var ofCtx context.Context
	var ofDone func()
	state := ofcStateCreated
	ofc.events <- ofcEventStart
top:
	for {
		select {
		case <-ctx.Done():
			state = ofcStateStopped
			logger.Debugw(ctx, "state-transition-context-done",
				log.Fields{"device-id": ofc.DeviceID})
			break top
		case event := <-ofc.events:
			previous := state
			switch event {
			case ofcEventStart:
				logger.Debugw(ctx, "ofc-event-start",
					log.Fields{"device-id": ofc.DeviceID})
				if state == ofcStateCreated {
					state = ofcStateStarted
					logger.Debug(ctx, "STARTED MORE THAN ONCE")
					go func() {
						if err := ofc.establishConnectionToController(ctx); err != nil {
							logger.Errorw(ctx, "controller-connection-failed", log.Fields{"error": err})
							panic(err)
						}
					}()
				} else {
					logger.Errorw(ctx, "illegal-state-transition",
						log.Fields{
							"device-id":     ofc.DeviceID,
							"current-state": state.String(),
							"event":         event.String()})
				}
			case ofcEventConnect:
				logger.Debugw(ctx, "ofc-event-connected",
					log.Fields{"device-id": ofc.DeviceID})
				if state == ofcStateStarted || state == ofcStateDisconnected {
					state = ofcStateConnected
					ofCtx, ofDone = context.WithCancel(log.WithSpanFromContext(context.Background(), ctx))
					go ofc.messageSender(ofCtx)
					go ofc.processOFStream(ofCtx)
				} else {
					logger.Errorw(ctx, "illegal-state-transition",
						log.Fields{
							"device-id":     ofc.DeviceID,
							"current-state": state.String(),
							"event":         event.String()})
				}
			case ofcEventDisconnect:
				logger.Debugw(ctx, "ofc-event-disconnected",
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
						if err := ofc.establishConnectionToController(ctx); err != nil {
							logger.Errorw(ctx, "controller-connection-failed", log.Fields{"error": err})
							panic(err)
						}
					}()
				} else {
					logger.Errorw(ctx, "illegal-state-transition",
						log.Fields{
							"device-id":     ofc.DeviceID,
							"current-state": state.String(),
							"event":         event.String()})
				}
			case ofcEventStop:
				logger.Debugw(ctx, "ofc-event-stop",
					log.Fields{"device-id": ofc.DeviceID})
				if state == ofcStateCreated || state == ofcStateConnected || state == ofcStateDisconnected {
					state = ofcStateStopped
					break top
				} else {
					logger.Errorw(ctx, "illegal-state-transition",
						log.Fields{
							"device-id":     ofc.DeviceID,
							"current-state": state.String(),
							"event":         event.String()})
				}
			}
			logger.Debugw(ctx, "state-transition",
				log.Fields{
					"device-id":      ofc.DeviceID,
					"previous-state": previous.String(),
					"current-state":  state.String(),
					"event":          event.String()})
		}
	}

	// If the child context exists, then cancel it
	if ofDone != nil {
		logger.Debugw(ctx, "closing-child-processes",
			log.Fields{"device-id": ofc.DeviceID})
		ofDone()
	}

	// If the connection is open, then close it
	if ofc.conn != nil {
		logger.Debugw(ctx, "closing-of-connection",
			log.Fields{"device-id": ofc.DeviceID})
		ofc.conn.Close()
		ofc.conn = nil
	}
	logger.Debugw(ctx, "state-machine-finished",
		log.Fields{"device-id": ofc.DeviceID})
}

// processOFStream processes the OF connection from the controller and invokes
// the appropriate handler methods for each message.
func (ofc *OFConnection) processOFStream(ctx context.Context) {
	fromController := bufio.NewReader(ofc.conn)

	/*
	 * We have a read buffer of a max size of 4096, so if we ever have
	 * a message larger than this then we will have issues
	 */
	headerBuf := make([]byte, 8)
	wg := sync.WaitGroup{}
top:
	// Continue until we are told to stop
	for {
		select {
		case <-ctx.Done():
			logger.Error(ctx, "of-loop-ending-context-done")
			break top
		default:
			// Read 8 bytes, the standard OF header
			read, err := io.ReadFull(fromController, headerBuf)
			if err != nil {
				if err == io.EOF {
					logger.Infow(ctx, "controller-disconnected",
						log.Fields{
							"device-id":  ofc.DeviceID,
							"controller": ofc.OFControllerEndPoint,
						})
				} else {
					logger.Errorw(ctx, "bad-of-header",
						log.Fields{
							"byte-count": read,
							"device-id":  ofc.DeviceID,
							"controller": ofc.OFControllerEndPoint,
							"error":      err})
				}
				break top
			}

			// Decode the header
			peek, err := ofc.peekAtOFHeader(headerBuf)
			if err != nil {
				/*
				 * Header is bad, assume stream is corrupted
				 * and needs to be restarted
				 */
				logger.Errorw(ctx, "bad-of-packet",
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
				logger.Errorw(ctx, "bad-of-packet",
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
				logger.Errorw(ctx, "failed-to-decode",
					log.Fields{
						"device-id": ofc.DeviceID,
						"decoder":   js,
						"error":     err})
				break top
			}
			if logger.V(log.DebugLevel) {
				js, _ := json.Marshal(msg)
				logger.Debugw(ctx, "packet-header",
					log.Fields{
						"device-id": ofc.DeviceID,
						"header":    js})
			}

			// We can parallelize the processing of all the operations
			// that we get before a BarrieRequest, then we need to wait.
			// What we are doing is:
			// - spawn threads until we get a Barrier
			// - when we get a barrier wait for the threads to complete before continuing

			msgType := msg.GetType()
			if msgType == ofp.OFPTBarrierRequest {
				logger.Debug(ctx, "received-barrier-request-waiting-for-pending-requests")
				wg.Wait()
				logger.Debug(ctx, "restarting-requests-processing")
			}

			wg.Add(1)
			go ofc.parseHeader(ctx, msg, &wg)
		}
	}
	logger.Debugw(ctx, "end-of-stream",
		log.Fields{"device-id": ofc.DeviceID})
	ofc.events <- ofcEventDisconnect
}

func (ofc *OFConnection) sayHello(ctx context.Context) {
	hello := ofp.NewHello()
	hello.Xid = uint32(GetXid())
	elem := ofp.NewHelloElemVersionbitmap()
	elem.SetType(ofp.OFPHETVersionbitmap)
	elem.SetLength(8)
	elem.SetBitmaps([]*ofp.Uint32{{Value: 16}})
	hello.SetElements([]ofp.IHelloElem{elem})
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(hello)
		logger.Debugw(ctx, "sayHello Called",
			log.Fields{
				"device-id":     ofc.DeviceID,
				"hello-message": js})
	}
	if err := ofc.SendMessage(ctx, hello); err != nil {
		logger.Fatalw(ctx, "Failed saying hello to Openflow Server, unable to proceed",
			log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
	}
}

func (ofc *OFConnection) parseHeader(ctx context.Context, header ofp.IHeader, wg *sync.WaitGroup) {
	defer wg.Done()
	headerType := header.GetType()
	logger.Debugw(ctx, "packet-header-type",
		log.Fields{
			"header-type": ofp.Type(headerType).String()})
	switch headerType {
	case ofp.OFPTHello:
		//x := header.(*ofp.Hello)
	case ofp.OFPTError:
		go ofc.handleErrMsg(ctx, header.(*ofp.ErrorMsg))
	case ofp.OFPTEchoRequest:
		go ofc.handleEchoRequest(ctx, header.(*ofp.EchoRequest))
	case ofp.OFPTEchoReply:
	case ofp.OFPTExperimenter:
	case ofp.OFPTFeaturesRequest:
		go func() {
			if err := ofc.handleFeatureRequest(ctx, header.(*ofp.FeaturesRequest)); err != nil {
				logger.Errorw(ctx, "handle-feature-request", log.Fields{"error": err})
			}
		}()
	case ofp.OFPTFeaturesReply:
	case ofp.OFPTGetConfigRequest:
		go ofc.handleGetConfigRequest(ctx, header.(*ofp.GetConfigRequest))
	case ofp.OFPTGetConfigReply:
	case ofp.OFPTSetConfig:
		go ofc.handleSetConfig(ctx, header.(*ofp.SetConfig))
	case ofp.OFPTPacketIn:
	case ofp.OFPTFlowRemoved:
	case ofp.OFPTPortStatus:
	case ofp.OFPTPacketOut:
		if !(ofc.role == ofcRoleMaster || ofc.role == ofcRoleEqual) {
			ofc.sendRoleSlaveError(ctx, header)
			return
		}
		go ofc.handlePacketOut(ctx, header.(*ofp.PacketOut))
	case ofp.OFPTFlowMod:
		if !(ofc.role == ofcRoleMaster || ofc.role == ofcRoleEqual) {
			ofc.sendRoleSlaveError(ctx, header)
			return
		}
		switch header.(ofp.IFlowMod).GetCommand() {
		case ofp.OFPFCAdd:
			ofc.handleFlowAdd(ctx, header.(*ofp.FlowAdd))
		case ofp.OFPFCModify:
			ofc.handleFlowMod(ctx, header.(*ofp.FlowMod))
		case ofp.OFPFCModifyStrict:
			ofc.handleFlowModStrict(ctx, header.(*ofp.FlowModifyStrict))
		case ofp.OFPFCDelete:
			ofc.handleFlowDelete(ctx, header.(*ofp.FlowDelete))
		case ofp.OFPFCDeleteStrict:
			ofc.handleFlowDeleteStrict(ctx, header.(*ofp.FlowDeleteStrict))
		}
	case ofp.OFPTStatsRequest:
		go func() {
			if err := ofc.handleStatsRequest(ctx, header, header.(ofp.IStatsRequest).GetStatsType()); err != nil {
				logger.Errorw(ctx, "ofpt-stats-request", log.Fields{"error": err})
			}
		}()
	case ofp.OFPTBarrierRequest:
		/* See note above at case ofp.OFPTFlowMod:*/
		ofc.handleBarrierRequest(ctx, header.(*ofp.BarrierRequest))
	case ofp.OFPTRoleRequest:
		go ofc.handleRoleRequest(ctx, header.(*ofp.RoleRequest))
	case ofp.OFPTMeterMod:
		if !(ofc.role == ofcRoleMaster || ofc.role == ofcRoleEqual) {
			ofc.sendRoleSlaveError(ctx, header)
			return
		}
		ofc.handleMeterModRequest(ctx, header.(*ofp.MeterMod))
	case ofp.OFPTGroupMod:
		if !(ofc.role == ofcRoleMaster || ofc.role == ofcRoleEqual) {
			ofc.sendRoleSlaveError(ctx, header)
			return
		}
		ofc.handleGroupMod(ctx, header.(ofp.IGroupMod))
	}
}

// Message interface that represents an open flow message and enables for a
// unified implementation of SendMessage
type Message interface {
	Serialize(encoder *goloxi.Encoder) error
}

func (ofc *OFConnection) doSend(ctx context.Context, msg Message) error {
	if ofc.conn == nil {
		return errors.New("no-connection")
	}
	enc := goloxi.NewEncoder()
	if err := msg.Serialize(enc); err != nil {
		return err
	}

	bytes := enc.Bytes()
	if _, err := ofc.conn.Write(bytes); err != nil {
		logger.Errorw(ctx, "unable-to-send-message-to-controller",
			log.Fields{
				"device-id": ofc.DeviceID,
				"message":   msg,
				"error":     err})
		return err
	}
	return nil
}

func (ofc *OFConnection) messageSender(ctx context.Context) {
	// first process last fail if it exists
	if ofc.lastUnsentMessage != nil {
		if err := ofc.doSend(ctx, ofc.lastUnsentMessage); err != nil {
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
			if err := ofc.doSend(ctx, msg); err != nil {
				ofc.lastUnsentMessage = msg
				ofc.events <- ofcEventDisconnect
				logger.Debugw(ctx, "message-sender-error",
					log.Fields{
						"device-id": ofc.DeviceID,
						"error":     err.Error()})
				break top
			}
			logger.Debugw(ctx, "message-sender-send",
				log.Fields{
					"device-id": ofc.DeviceID})
			ofc.lastUnsentMessage = nil
		}
	}

	logger.Debugw(ctx, "message-sender-finished",
		log.Fields{
			"device-id": ofc.DeviceID})
}

// SendMessage queues a message to be sent to the openflow controller
func (ofc *OFConnection) SendMessage(ctx context.Context, message Message) error {
	logger.Debugw(ctx, "queuing-message", log.Fields{
		"endpoint": ofc.OFControllerEndPoint,
		"role":     ofc.role,
	})
	ofc.sendChannel <- message
	return nil
}
