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
	"encoding/json"
	"fmt"
	"github.com/donNewtonAlpha/goloxi"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	voltha "github.com/opencord/voltha-protos/v2/go/voltha"
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
func NewClient(ofControllerEndPoint string, deviceID string, keepRunning bool) *OFClient {

	ofc := OFClient{
		OFControllerEndPoint: ofControllerEndPoint,
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
				if log.GetDefaultLogLevel() <= log.DebugLevel {
					fmt.Printf("conn.Read read %d bytes\n", read)
				}
				if read < 8 {
					continue
				}

				if e != nil {
					log.Errorw("Voltha connection has died", log.Fields{"DeviceID": ofc.DeviceID, "Error": e})
					break
				}
				decoder := goloxi.NewDecoder(buf)
				header, e := ofp.DecodeHeader(decoder)

				if e != nil {
					js, _ := json.Marshal(decoder)
					log.Errorw("DecodeHeader threw error", log.Fields{"DeviceID": ofc.DeviceID, "Decoder": js, "Error": e})
				}
				if log.GetDefaultLogLevel() <= log.DebugLevel {
					js, _ := json.Marshal(header)
					log.Debugw("Header Decode", log.Fields{"DeviceID": ofc.DeviceID, "Header": js})
				}
				ofc.parseHeader(header) //first one is ready
				len := header.GetLength()
				if len < uint16(read) {
					for {
						read = read - int(len)
						newBuf := buf[len:]
						decoder = goloxi.NewDecoder(newBuf)
						newHeader, e := ofp.DecodeHeader(decoder)
						if e != nil {
							js, _ := json.Marshal(decoder)
							log.Errorw("DecodeHeader threw error", log.Fields{"DeviceID": ofc.DeviceID, "Decoder": js, "Error": e})
						}
						if e != nil {
							js, _ := json.Marshal(decoder)
							log.Errorw("DecodeHeader threw error", log.Fields{"DeviceID": ofc.DeviceID, "Decoder": js, "Error": e})
						}
						if log.GetDefaultLogLevel() <= log.DebugLevel {
							js, _ := json.Marshal(header)
							log.Debugw("Header Decode", log.Fields{"DeviceID": ofc.DeviceID, "Header": js})
						}
						len = newHeader.GetLength()
						ofc.parseHeader(newHeader)
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
		go ofc.handleErrMsg(header.(*ofp.ErrorMsg))
	case ofp.OFPTEchoRequest:
		go ofc.handleEchoRequest(header.(*ofp.EchoRequest))
	case ofp.OFPTEchoReply:
	case ofp.OFPTExperimenter:
	case ofp.OFPTFeaturesRequest:
		featReq := header.(*ofp.FeaturesRequest)
		go ofc.handleFeatureRequest(featReq)
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
		switch header.(*ofp.FlowMod).GetType() {
		case ofp.OFPFCAdd:
			go ofc.handleFlowAdd(header.(*ofp.FlowAdd))
		case ofp.OFPFCModify:
			go ofc.handleFlowMod(header.(*ofp.FlowMod))
		case ofp.OFPFCModifyStrict:
			go ofc.handleFlowModStrict(header.(*ofp.FlowModifyStrict))
		case ofp.OFPFCDelete:
			go ofc.handleFlowDelete(header.(*ofp.FlowDelete))
		case ofp.OFPFCDeleteStrict:
			go ofc.handleFlowDeleteStrict(header.(*ofp.FlowDeleteStrict))
		}
	case ofp.OFPTStatsRequest:
		go ofc.handleStatsRequest(header, header.(*ofp.StatsRequest).GetStatsType())
	case ofp.OFPTBarrierRequest:
		go ofc.handleBarrierRequest(header.(*ofp.BarrierRequest))
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
