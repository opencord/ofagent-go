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

package openflow

import (
	"encoding/binary"
	"encoding/json"
	"github.com/donNewtonAlpha/goloxi"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/voltha"
	"io"
	"net"
	"time"
)

//Client structure to hold fields of Openflow Client
type OFClient struct {
	OFControllerEndPoint string
	DeviceID             string
	KeepRunning          bool
	conn                 net.Conn
	volthaClient         voltha.VolthaServiceClient
	packetOutChannel     chan voltha.PacketOut
}

//NewClient  contstructs a new Openflow Client and then starts up
func NewClient(volthaClient voltha.VolthaServiceClient, poc chan voltha.PacketOut,
	ofControllerEndPoint string, deviceID string, keepRunning bool) *OFClient {

	ofc := OFClient{
		volthaClient:         volthaClient,
		OFControllerEndPoint: ofControllerEndPoint,
		packetOutChannel:     poc,
		DeviceID:             deviceID,
		KeepRunning:          keepRunning}
	raddr, e := net.ResolveTCPAddr("tcp", ofControllerEndPoint)
	if e != nil {
		log.Fatalw("openflowClient Start unable to resolve address", log.Fields{"end-point": ofControllerEndPoint})
	}
	if connection, err := net.DialTCP("tcp", nil, raddr); err != nil {
		log.Fatalf("openflowClient Unable to connect to voltha @  %v exiting", raddr)
	} else {
		ofc.conn = connection
		ofc.sayHello()
	}
	return &ofc
}

//End - set keepRunning to false so start loop exits
func (ofc *OFClient) End() {
	ofc.KeepRunning = false
}

func (ofc *OFClient) peekAtOFHeader(buf []byte) (ofp.IHeader, error) {
	header := ofp.Header{}
	header.Version = uint8(buf[0])
	header.Type = uint8(buf[1])
	log.Debugw("length-is",
		log.Fields{"length": binary.BigEndian.Uint16(buf[2:4])})
	header.Length = binary.BigEndian.Uint16(buf[2:4])
	header.Xid = binary.BigEndian.Uint32(buf[4:8])

	return &header, nil
}

//Start run loop for the openflow client
func (ofc *OFClient) Run() {

	// TODO: should buffer size be configurable
	buf := make([]byte, 1500)
	var need, have int

	/*
	 * First thing we are looking for is an open flow header, so we need
	 * at least 8 bytes
	 */
	need = 8

	for ofc.KeepRunning {
		log.Debugw("top-of-loop",
			log.Fields{
				"have":       have,
				"need":       need,
				"buf-length": len(buf[have:])})
		read, err := ofc.conn.Read(buf[have:])
		have += read
		log.Debugw("read-of-connection",
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
			log.Debugw("1-need-more",
				log.Fields{
					"have":  have,
					"need":  need,
					"error": err})
			continue
		}

		log.Debugw("read-of-connection",
			log.Fields{
				"device-id":  ofc.DeviceID,
				"byte-count": read})

		/*
		 * Single out EOF here, because if we have bytes
		 * but have an EOF we still want to process the
		 * the last meesage. A read of 0 bytes and EOF is
		 * a terminated connection.
		 */
		if err != nil && (err != io.EOF || read == 0) {
			log.Errorw("Voltha connection has died",
				log.Fields{
					"device-id": ofc.DeviceID,
					"error":     err})
			break
		}

		/*
		 * If we get here, we have at least the 8 bytes of the
		 * header, if not enough for the complete message. So
		 * take a peek at the OF header to do simple validation
		 * and be able to get the full expected length of the
		 * packet
		 */
		log.Debugw("2-need-more",
			log.Fields{
				"have": have,
				"need": need})
		for need <= have {
			log.Debug("top-of-inner-loop")
			peek, err := ofc.peekAtOFHeader(buf)
			if err != nil {
				log.Fatal("bad-of-packet",
					log.Fields{
						"device-id": ofc.DeviceID,
						"error":     err})
			}

			/*
			 * If we don't have the full packet, then back around
			 * the outer loop to get more bytes
			 */
			need = int(peek.GetLength())
			log.Debugw("3-need-more",
				log.Fields{
					"have": have,
					"need": need})

			if have < need {
				log.Debug("need-more")
				break
			}

			// Decode and process the packet
			decoder := goloxi.NewDecoder(buf)
			header, err := ofp.DecodeHeader(decoder)
			if err != nil {
				js, _ := json.Marshal(decoder)
				log.Fatalw("DecodeHeader threw error",
					log.Fields{
						"device-id": ofc.DeviceID,
						"decoder":   js,
						"error":     err})
				// TODO: bad packet? continue?
			}
			if log.GetDefaultLogLevel() <= log.DebugLevel {
				js, _ := json.Marshal(header)
				log.Debugw("Header Decode",
					log.Fields{
						"device-id": ofc.DeviceID,
						"reader":    js})
			}
			ofc.parseHeader(header)
			copy(buf, buf[need:])
			have -= need
			need = 8
			log.Debugw("4-need-more",
				log.Fields{
					"have":        have,
					"need":        need,
					"buf-length":  len(buf),
					"read-length": len(buf[have:])})
		}
	}
}

func (ofc *OFClient) sayHello() {
	hello := ofp.NewHello()
	hello.Xid = uint32(GetXid())
	var elements []ofp.IHelloElem
	elem := ofp.NewHelloElemVersionbitmap()
	elem.SetType(ofp.OFPHETVersionbitmap)
	elem.SetLength(8)
	var bitmap = ofp.Uint32{Value: 16}
	bitmaps := []*ofp.Uint32{&bitmap}
	elem.SetBitmaps(bitmaps)
	elements = append(elements, elem)
	hello.SetElements(elements)
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		js, _ := json.Marshal(hello)
		log.Debugw("sayHello Called", log.Fields{"DeviceID": ofc.DeviceID, "HelloMessage": js})
	}
	e := ofc.SendMessage(hello)
	if e != nil {
		log.Fatalw("Failed saying hello to Openflow Server, unable to proceed",
			log.Fields{"DeviceID": ofc.DeviceID, "Error": e})
	}
}
func (ofc *OFClient) parseHeader(header ofp.IHeader) {
	switch header.GetType() {
	case ofp.OFPTHello:
		//x := header.(*ofp.Hello)
	case ofp.OFPTError:
		go ofc.handleErrMsg(header.(ofp.IErrorMsg))
	case ofp.OFPTEchoRequest:
		go ofc.handleEchoRequest(header.(ofp.IEchoRequest))
	case ofp.OFPTEchoReply:
	case ofp.OFPTExperimenter:
	case ofp.OFPTFeaturesRequest:
		featReq := header.(ofp.IFeaturesRequest)
		go ofc.handleFeatureRequest(featReq)
	case ofp.OFPTFeaturesReply:
	case ofp.OFPTGetConfigRequest:
		go ofc.handleGetConfigRequest(header.(ofp.IGetConfigRequest))
	case ofp.OFPTGetConfigReply:
	case ofp.OFPTSetConfig:
		go ofc.handleSetConfig(header.(ofp.ISetConfig))
	case ofp.OFPTPacketIn:
	case ofp.OFPTFlowRemoved:
	case ofp.OFPTPortStatus:
	case ofp.OFPTPacketOut:
		go ofc.handlePacketOut(header.(ofp.IPacketOut))
	case ofp.OFPTFlowMod:
		switch header.(ofp.IFlowMod).GetCommand() {
		case ofp.OFPFCAdd:
			go ofc.handleFlowAdd(header.(ofp.IFlowAdd))
		case ofp.OFPFCModify:
			go ofc.handleFlowMod(header.(ofp.IFlowMod))
		case ofp.OFPFCModifyStrict:
			go ofc.handleFlowModStrict(header.(ofp.IFlowModifyStrict))
		case ofp.OFPFCDelete:
			go ofc.handleFlowDelete(header.(ofp.IFlowDelete))
		case ofp.OFPFCDeleteStrict:
			go ofc.handleFlowDeleteStrict(header.(ofp.IFlowDeleteStrict))
		default:
			log.Fatalw("Unknown type",
				log.Fields{"type": header.(ofp.IFlowMod).GetType()})
		}
	case ofp.OFPTStatsRequest:
		go ofc.handleStatsRequest(header, header.(ofp.IStatsRequest).GetStatsType())
	case ofp.OFPTBarrierRequest:
		go ofc.handleBarrierRequest(header.(ofp.IBarrierRequest))
	case ofp.OFPTRoleRequest:
		go ofc.handleRoleRequest(header.(ofp.IRoleRequest))
	case ofp.OFPTMeterMod:
		go ofc.handleMeterModRequest(header.(ofp.IMeterMod))
	}
}

//Message created to allow for a single SendMessage
type Message interface {
	Serialize(encoder *goloxi.Encoder) error
}

//SendMessage sends message to openflow server
func (ofc *OFClient) SendMessage(message Message) error {
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		js, _ := json.Marshal(message)
		log.Debugw("SendMessage called", log.Fields{"DeviceID": ofc.DeviceID, "Message": js})
	}
	enc := goloxi.NewEncoder()
	message.Serialize(enc)
	for {
		if ofc.conn == nil {
			log.Warnln("SendMessage Connection is Nil sleeping for 10 milliseconds")
			time.Sleep(10 * time.Millisecond)
		} else {
			break
		}
	}
	bytes := enc.Bytes()
	_, err := ofc.conn.Write(bytes)
	if err != nil {
		jMessage, _ := json.Marshal(message)
		log.Errorw("SendMessage failed sending message", log.Fields{"DeviceID": ofc.DeviceID, "Error": err, "Message": jMessage})
		return err
	}
	return nil
}

//SetGrpcClient store a reference to the grpc client connection
func (ofc *OFClient) SetVolthaServiceClient(volthaClient voltha.VolthaServiceClient) {
	ofc.volthaClient = volthaClient
}

//SetPacketOutChannel - store the channel to send packet outs to grpc connection
func (ofc *OFClient) SetPacketOutChannel(pktOutChan chan voltha.PacketOut) {
	ofc.packetOutChannel = pktOutChan
}
