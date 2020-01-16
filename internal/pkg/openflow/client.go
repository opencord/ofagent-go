/*
   Copyright 2019 the original author or authors.

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
	"encoding/json"
	"fmt"
	"github.com/donNewtonAlpha/goloxi"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/voltha"
	"net"
	"time"
)

var logger, _ = log.AddPackage(log.JSON, log.DebugLevel, nil)

//Client structure to hold fields of Openflow Client
type OFClient struct {
	OFControllerEndPoint string
	Port                 uint16
	DeviceID             string
	KeepRunning          bool
	conn                 net.Conn
	volthaClient         voltha.VolthaServiceClient
	packetOutChannel     chan voltha.PacketOut
}

//NewClient  contstructs a new Openflow Client and then starts up
func NewOFClient(deviceID string, ofControllerEndPoint string,
	volthaClient voltha.VolthaServiceClient, packetOutChannel chan voltha.PacketOut,
	keepRunning bool) *OFClient {

	ofc := OFClient{
		DeviceID:             deviceID,
		OFControllerEndPoint: ofControllerEndPoint,
		volthaClient:         volthaClient,
		packetOutChannel:     packetOutChannel,
		KeepRunning:          keepRunning,
	}
	raddr, e := net.ResolveTCPAddr("tcp", ofc.OFControllerEndPoint)
	if e != nil {
		logger.Fatalw("openflowClient Start unable to resolve address",
			log.Fields{
				"endpoint": ofc.OFControllerEndPoint})
	}
	connection, e := net.DialTCP("tcp", nil, raddr)
	ofc.conn = connection
	if e != nil {
		logger.Fatalf("openflowClient Unable to connect to controller @ %v exiting", raddr)
	}
	ofc.sayHello()
	return &ofc
}

//End - set keepRunning to false so start loop exits
func (ofc *OFClient) End() {
	ofc.KeepRunning = false
}

//Start run loop for the openflow client
func (ofc *OFClient) Start() {

	for {
		if ofc.KeepRunning {
			defer ofc.conn.Close()
			for {
				defer func() {

					if r := recover(); r != nil {
						err := r.(error)
						fmt.Printf("Caught error in ofc.Start() %v \n ", err)
					}
				}()

				if !ofc.KeepRunning {
					return
				}
				buf := make([]byte, 1500)
				read, e := ofc.conn.Read(buf)
				if logger.V(log.DebugLevel) {
					fmt.Printf("conn.Read read %d bytes\n", read)
				}
				if read < 8 {
					continue
				}

				if e != nil {
					logger.Errorw("Voltha connection has died", log.Fields{"device-id": ofc.DeviceID, "Error": e})
					break
				}
				decoder := goloxi.NewDecoder(buf)
				header, e := ofp.DecodeHeader(decoder)

				if e != nil {
					js, _ := json.Marshal(decoder)
					logger.Errorw("DecodeHeader threw error", log.Fields{"device-id": ofc.DeviceID, "Decoder": js, "Error": e})
				}
				if logger.V(log.DebugLevel) {
					js, _ := json.Marshal(header)
					logger.Debugw("Header Decode", log.Fields{"device-id": ofc.DeviceID, "Header": js})
				}
				ofc.parseHeader(header, buf) //first one is ready
				len := header.GetLength()
				if len < uint16(read) {
					for {
						read = read - int(len)
						newBuf := buf[len:]
						decoder = goloxi.NewDecoder(newBuf)
						newHeader, e := ofp.DecodeHeader(decoder)
						if e != nil {
							js, _ := json.Marshal(decoder)
							logger.Errorw("DecodeHeader threw error", log.Fields{"device-id": ofc.DeviceID, "Decoder": js, "Error": e})
						}
						if e != nil {
							js, _ := json.Marshal(decoder)
							logger.Errorw("DecodeHeader threw error", log.Fields{"device-id": ofc.DeviceID, "Decoder": js, "Error": e})
						}
						if logger.V(log.DebugLevel) {
							js, _ := json.Marshal(header)
							logger.Debugw("Header Decode", log.Fields{"device-id": ofc.DeviceID, "Header": js})
						}
						len = newHeader.GetLength()
						ofc.parseHeader(newHeader, newBuf)
						if read == int(len) {
							break
						}
						buf = newBuf
					}
				}
			}
		}
		break
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
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(hello)
		logger.Debugw("sayHello Called", log.Fields{"device-id": ofc.DeviceID, "HelloMessage": js})
	}
	e := ofc.SendMessage(hello)
	if e != nil {
		logger.Fatalw("Failed saying hello to Openflow Server, unable to proceed",
			log.Fields{"device-id": ofc.DeviceID, "Error": e})
	}
}
func (ofc *OFClient) parseHeader(header ofp.IHeader, buf []byte) {
	switch header.GetType() {
	case ofp.OFPTHello:
		//x := header.(*ofp.Hello)
	case ofp.OFPTError:
		errMsg := header.(*ofp.ErrorMsg)
		go ofc.handleErrMsg(errMsg)
	case ofp.OFPTEchoRequest:
		echoReq := header.(*ofp.EchoRequest)
		go ofc.handleEchoRequest(echoReq)
	case ofp.OFPTEchoReply:
	case ofp.OFPTExperimenter:
	case ofp.OFPTFeaturesRequest:
		featReq := header.(*ofp.FeaturesRequest)
		go ofc.handleFeatureRequest(featReq)
	case ofp.OFPTFeaturesReply:
	case ofp.OFPTGetConfigRequest:
		configReq := header.(*ofp.GetConfigRequest)
		go ofc.handleGetConfigRequest(configReq)
	case ofp.OFPTGetConfigReply:
	case ofp.OFPTSetConfig:
		setConf := header.(*ofp.SetConfig)
		go ofc.handleSetConfig(setConf)
	case ofp.OFPTPacketIn:
	case ofp.OFPTFlowRemoved:
	case ofp.OFPTPortStatus:
	case ofp.OFPTPacketOut:
		packetOut := header.(*ofp.PacketOut)
		go ofc.handlePacketOut(packetOut)
	case ofp.OFPTFlowMod:
		/* Not using go routine to handle flow* messages or barrier requests
		onos typically issues barrier requests just before a flow* message.
		by handling in this thread I ensure all flow* are handled when barrier
		request is issued.
		*/
		flowModType := uint8(buf[25])
		switch flowModType {
		case ofp.OFPFCAdd:
			flowAdd := header.(*ofp.FlowAdd)
			ofc.handleFlowAdd(flowAdd)
		case ofp.OFPFCModify:
			flowMod := header.(*ofp.FlowMod)
			ofc.handleFlowMod(flowMod)
		case ofp.OFPFCModifyStrict:
			flowModStrict := header.(*ofp.FlowModifyStrict)
			ofc.handleFlowModStrict(flowModStrict)
		case ofp.OFPFCDelete:
			flowDelete := header.(*ofp.FlowDelete)
			ofc.handleFlowDelete(flowDelete)
		case ofp.OFPFCDeleteStrict:
			flowDeleteStrict := header.(*ofp.FlowDeleteStrict)
			ofc.handleFlowDeleteStrict(flowDeleteStrict)
		}
	case ofp.OFPTStatsRequest:
		var statType = uint16(buf[8])<<8 + uint16(buf[9])
		go ofc.handleStatsRequest(header, statType)
	case ofp.OFPTBarrierRequest:
		/* See note above at case ofp.OFPTFlowMod:*/
		barRequest := header.(*ofp.BarrierRequest)
		ofc.handleBarrierRequest(barRequest)
	case ofp.OFPTRoleRequest:
		roleReq := header.(*ofp.RoleRequest)
		go ofc.handleRoleRequest(roleReq)
	case ofp.OFPTMeterMod:
		meterMod := header.(*ofp.MeterMod)
		go ofc.handleMeterModRequest(meterMod)

	}
}

//Message created to allow for a single SendMessage
type Message interface {
	Serialize(encoder *goloxi.Encoder) error
}

//SendMessage sends message to openflow server
func (ofc *OFClient) SendMessage(message Message) error {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(message)
		logger.Debugw("SendMessage called", log.Fields{"device-id": ofc.DeviceID, "Message": js})
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
	_, err := ofc.conn.Write(bytes)
	if err != nil {
		jMessage, _ := json.Marshal(message)
		logger.Errorw("SendMessage failed sending message", log.Fields{"device-id": ofc.DeviceID, "Error": err, "Message": jMessage})
		return err
	}
	return nil
}
