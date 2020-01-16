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
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/donNewtonAlpha/goloxi"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/voltha"
	"io"
	"net"
	"time"
)

var logger, _ = log.AddPackage(log.JSON, log.DebugLevel, nil)

type ofcEvent byte
type ofcState byte

const (
	ofcEventStart = ofcEvent(iota)
	ofcEventConnected
	ofcEventDisconnected

	ofcStateConnected = ofcState(iota)
	ofcStateDisconnected
)

//Client structure to hold fields of Openflow Client
type OFClient struct {
	OFControllerEndPoint string
	Port                 uint16
	DeviceID             string
	KeepRunning          bool
	VolthaClient         voltha.VolthaServiceClient
	PacketOutChannel     chan *voltha.PacketOut
	ConnectionMaxRetries int
	ConnectionRetryDelay time.Duration
	conn                 net.Conn

	// expirimental
	events            chan ofcEvent
	sendChannel       chan Message
	lastUnsentMessage Message
}

//NewClient  contstructs a new Openflow Client and then starts up
func NewOFClient(config *OFClient) *OFClient {

	ofc := OFClient{
		DeviceID:             config.DeviceID,
		OFControllerEndPoint: config.OFControllerEndPoint,
		VolthaClient:         config.VolthaClient,
		PacketOutChannel:     config.PacketOutChannel,
		KeepRunning:          config.KeepRunning,
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

//End - set keepRunning to false so start loop exits
func (ofc *OFClient) Stop() {
	ofc.KeepRunning = false
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
				ofc.events <- ofcEventConnected
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

func (ofc *OFClient) Run(ctx context.Context) {

	var ofCtx context.Context
	var ofDone func()
	ofc.events <- ofcEventStart
	state := ofcStateDisconnected
top:
	for {
		select {
		case <-ctx.Done():
			break top
		case event := <-ofc.events:
			switch event {
			case ofcEventStart:
				logger.Debugw("ofc-event-star",
					log.Fields{"device-id": ofc.DeviceID})
				go ofc.establishConnectionToController()
			case ofcEventConnected:
				if state == ofcStateDisconnected {
					state = ofcStateConnected
					logger.Debugw("ofc-event-connected",
						log.Fields{"device-id": ofc.DeviceID})
					ofCtx, ofDone = context.WithCancel(context.Background())
					go ofc.messageSender(ofCtx)
					go ofc.processOFStream(ofCtx)
				}
			case ofcEventDisconnected:
				if state == ofcStateConnected {
					state = ofcStateDisconnected
					logger.Debugw("ofc-event-disconnected",
						log.Fields{"device-id": ofc.DeviceID})
					if ofDone != nil {
						ofDone()
						ofDone = nil
					}
					go ofc.establishConnectionToController()
				}
			}
		}
	}

	if ofDone != nil {
		ofDone()
		ofDone = nil
	}

}

// Run run loop for the openflow client
func (ofc *OFClient) processOFStream(ctx context.Context) {
	buf := make([]byte, 1500)
	var need, have int
	/*
	 * EXPLANATION
	 *
	 * The below loops reuses a byte array to read messages from the TCP
	 * connection to the OF controller. It reads messages into a large
	 * buffer in an attempt to optimize the read performance from the
	 * TCP connection. This means that on any given read there may be more
	 * than a single message in the byte array read.
	 *
	 * As the minimal size for an OF message is 8 bytes (because that is
	 * the size of the basic header) we know that if we have not read
	 * 8 bytes we need to read more before we can process a message.
	 *
	 * Once the mninium header is read, the complete length of the
	 * message is retrieved from the header and bytes are repeatedly read
	 * until we know the byte array contains at least one message.
	 *
	 * Once it is known that the buffer has at least a single message
	 * a slice (msg) is moved through the read bytes looking to process
	 * each message util the length of read data is < the length required
	 * i.e., the minimum size or the size of the next message.
	 *
	 * When no more message can be proessed from the byte array any unused
	 * bytes are moved to the front of the source array and more data is
	 * read from the TCP connection.
	 */

	/*
	 * First thing we are looking for is an openflow header, so we need at
	 * least 8 bytes
	 */
	need = 8

top:
	// Continue until we are told to stop
	for ofc.KeepRunning {
		logger.Debugw("before-read-from-controller",
			log.Fields{
				"device-id":  ofc.DeviceID,
				"have":       have,
				"need":       need,
				"buf-length": len(buf[have:])})
		read, err := ofc.conn.Read(buf[have:])
		have += read
		logger.Debugw("read-from-controller",
			log.Fields{
				"device-id":  ofc.DeviceID,
				"byte-count": read,
				"error":      err})

		/*
		 * If we have less than we need and there is no
		 * error, then continue to attempt to read more data
		 */
		if have < need && err == nil {
			// No bytes available, just continue
			logger.Debugw("continue-to-read",
				log.Fields{
					"device-id": ofc.DeviceID,
					"have":      have,
					"need":      need,
					"error":     err})
			continue
		}

		/*
		 * Single out EOF here, because if we have bytes
		 * but have an EOF we still want to process the
		 * the last meesage. A read of 0 bytes and EOF is
		 * a terminated connection.
		 */
		if err != nil && (err != io.EOF || read == 0) {
			logger.Errorw("voltha-connection-dead",
				log.Fields{
					"device-id": ofc.DeviceID,
					"error":     err})
			break
		}

		/*
		 * We should have at least 1 message at this point so
		 * create a slice (msg) that points to the start of the
		 * buffer
		 */
		msg := buf[0:]
		for need <= have {
			logger.Debugw("process-of-message-stream",
				log.Fields{
					"device-id": ofc.DeviceID,
					"have":      have,
					"need":      need})
			/*
			 * If we get here, we have at least the 8 bytes of the
			 * header, if not enough for the complete message. So
			 * take a peek at the OF header to do simple validation
			 * and be able to get the full expected length of the
			 * packet
			 */
			peek, err := ofc.peekAtOFHeader(msg)
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

			/*
			 * If we don't have the full packet, then back around
			 * the outer loop to get more bytes
			 */
			need = int(peek.GetLength())

			logger.Debugw("processed-header-need-message",
				log.Fields{
					"device-id": ofc.DeviceID,
					"have":      have,
					"need":      need})

			if have < need {
				logger.Debugw("end-processing:continue-to-read",
					log.Fields{
						"device-id": ofc.DeviceID,
						"have":      have,
						"need":      need})
				break
			}

			// Decode and process the packet
			decoder := goloxi.NewDecoder(msg)
			header, err := ofp.DecodeHeader(decoder)
			if err != nil {
				js, _ := json.Marshal(decoder)
				logger.Errorw("failed-to-decode",
					log.Fields{
						"device-id": ofc.DeviceID,
						"decoder":   js,
						"error":     err})
				break top
			}
			if logger.V(log.DebugLevel) {
				js, _ := json.Marshal(header)
				logger.Debugw("packet-header",
					log.Fields{
						"device-id": ofc.DeviceID,
						"header":    js})
			}
			ofc.parseHeader(header)

			/*
			 * Move the msg slice to the start of the next
			 * message, which is the current message plus the
			 * used bytes (need)
			 */
			msg = msg[need:]
			have -= need

			// Finished process method, need header again
			need = 8

			logger.Debugw("message-process-complete",
				log.Fields{
					"device-id":   ofc.DeviceID,
					"have":        have,
					"need":        need,
					"read-length": len(buf[have:])})
		}
		/*
		 * If we have any left over bytes move them to the front
		 * of the byte array to be appended to bny the next read
		 */
		if have > 0 {
			copy(buf, msg)
		}
	}
	ofc.events <- ofcEventDisconnected
}

func (ofc *OFClient) sayHello() {
	hello := ofp.NewHello()
	hello.Xid = uint32(GetXid())
	elem := ofp.NewHelloElemVersionbitmap()
	elem.SetType(ofp.OFPHETVersionbitmap)
	elem.SetLength(8)
	elem.SetBitmaps([]*ofp.Uint32{&ofp.Uint32{Value: 16}})
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
	switch header.GetType() {
	case ofp.OFPTHello:
		//x := header.(*ofp.Hello)
	case ofp.OFPTError:
		go ofc.handleErrMsg(header.(*ofp.ErrorMsg))
	case ofp.OFPTEchoRequest:
		go ofc.handleEchoRequest(header.(*ofp.EchoRequest))
	case ofp.OFPTEchoReply:
	case ofp.OFPTExperimenter:
	case ofp.OFPTFeaturesRequest:
		go ofc.handleFeatureRequest(header.(*ofp.FeaturesRequest))
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
		go ofc.handleStatsRequest(header, header.(ofp.IStatsRequest).GetStatsType())
	case ofp.OFPTBarrierRequest:
		/* See note above at case ofp.OFPTFlowMod:*/
		ofc.handleBarrierRequest(header.(*ofp.BarrierRequest))
	case ofp.OFPTRoleRequest:
		go ofc.handleRoleRequest(header.(*ofp.RoleRequest))
	case ofp.OFPTMeterMod:
		go ofc.handleMeterModRequest(header.(*ofp.MeterMod))
	}
}

//Message created to allow for a single SendMessage
type Message interface {
	Serialize(encoder *goloxi.Encoder) error
}

func (ofc *OFClient) doSend(msg Message) error {
	if ofc.conn == nil {
		return errors.New("no-connection")
	}
	enc := goloxi.NewEncoder()
	msg.Serialize(enc)
	bytes := enc.Bytes()
	if _, err := ofc.conn.Write(bytes); err != nil {
		logger.Warnw("unable-to-send-message-to-controller",
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
			ofc.events <- ofcEventDisconnected
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
			if ofc.doSend(msg) != nil {
				ofc.lastUnsentMessage = msg
				ofc.events <- ofcEventDisconnected
				return
			}
			ofc.lastUnsentMessage = nil
		}
	}
}

func (ofc *OFClient) SendMessage(message Message) error {
	ofc.sendChannel <- message
	return nil
}

//SendMessage sends message to openflow server
func (ofc *OFClient) SendMessageOrig(message Message) error {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(message)
		logger.Debugw("SendMessage called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"message":   js})
	}
	enc := goloxi.NewEncoder()
	message.Serialize(enc)
	for {
		if ofc.conn == nil {
			logger.Warnln("SendMessage Connection is Nil sleeping for 10 milliseconds")
			time.Sleep(10 * time.Millisecond)
		} else {
			break
		}
	}
	bytes := enc.Bytes()
	if _, err := ofc.conn.Write(bytes); err != nil {
		jMessage, _ := json.Marshal(message)
		logger.Errorw("SendMessage failed sending message",
			log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err,
				"message":   jMessage})
		return err
	}
	return nil
}
