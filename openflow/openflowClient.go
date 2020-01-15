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
	"time"

	"github.com/donNewtonAlpha/goloxi"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/ofagent-go/settings"
	l "github.com/opencord/voltha-lib-go/v2/pkg/log"
	pb "github.com/opencord/voltha-protos/v2/go/voltha"

	"net"
)

var volthaClient *pb.VolthaServiceClient
var packetOutChannel chan pb.PacketOut
var logger, _ = l.AddPackage(l.JSON, l.DebugLevel, nil)

//Client structure to hold fields of Openflow Client
type Client struct {
	OfServerAddress string
	Port            uint16
	DeviceID        string
	KeepRunning     bool
	conn            net.Conn
}

//NewClient  contstructs a new Openflow Client and then starts up
func NewClient(ofServerAddress string, port uint16, deviceID string, keepRunning bool) *Client {

	client := Client{OfServerAddress: ofServerAddress, Port: port, DeviceID: deviceID, KeepRunning: keepRunning}
	addressLine := fmt.Sprintf("%s:%d", client.OfServerAddress, client.Port)
	raddr, e := net.ResolveTCPAddr("tcp", addressLine)
	if e != nil {
		logger.Fatalw("openflowClient Start unable to resolve address", l.Fields{"Address": addressLine})
	}
	connection, e := net.DialTCP("tcp", nil, raddr)
	client.conn = connection
	if e != nil {
		logger.Fatalf("openflowClient Unable to connect to voltha @  %v exiting", raddr)
	}
	client.sayHello()
	return &client
}

//End - set keepRunning to false so start loop exits
func (client *Client) End() {
	client.KeepRunning = false
}

//Start run loop for the openflow client
func (client *Client) Start() {

	for {
		if client.KeepRunning {
			defer client.conn.Close()
			for {
				defer func() {

					if r := recover(); r != nil {
						err := r.(error)
						fmt.Printf("Caught error in client.Start() %v \n ", err)
					}
				}()

				if !client.KeepRunning {
					return
				}
				buf := make([]byte, 1500)
				read, e := client.conn.Read(buf)
				if settings.GetDebug(client.DeviceID) {
					fmt.Printf("conn.Read read %d bytes\n", read)
				}
				if read < 8 {
					continue
				}

				if e != nil {
					logger.Errorw("Voltha connection has died", l.Fields{"DeviceID": client.DeviceID, "Error": e})
					break
				}
				decoder := goloxi.NewDecoder(buf)
				header, e := ofp.DecodeHeader(decoder)

				if e != nil {
					js, _ := json.Marshal(decoder)
					logger.Errorw("DecodeHeader threw error", l.Fields{"DeviceID": client.DeviceID, "Decoder": js, "Error": e})
				}
				if settings.GetDebug(client.DeviceID) {
					js, _ := json.Marshal(header)
					logger.Debugw("Header Decode", l.Fields{"DeviceID": client.DeviceID, "Header": js})
				}
				client.parseHeader(header, buf) //first one is ready
				len := header.GetLength()
				if len < uint16(read) {
					for {
						read = read - int(len)
						newBuf := buf[len:]
						decoder = goloxi.NewDecoder(newBuf)
						newHeader, e := ofp.DecodeHeader(decoder)
						if e != nil {
							js, _ := json.Marshal(decoder)
							logger.Errorw("DecodeHeader threw error", l.Fields{"DeviceID": client.DeviceID, "Decoder": js, "Error": e})
						}
						if e != nil {
							js, _ := json.Marshal(decoder)
							logger.Errorw("DecodeHeader threw error", l.Fields{"DeviceID": client.DeviceID, "Decoder": js, "Error": e})
						}
						if settings.GetDebug(client.DeviceID) {
							js, _ := json.Marshal(header)
							logger.Debugw("Header Decode", l.Fields{"DeviceID": client.DeviceID, "Header": js})
						}
						len = newHeader.GetLength()
						client.parseHeader(newHeader, newBuf)
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
func (client *Client) sayHello() {
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
	if settings.GetDebug(client.DeviceID) {
		js, _ := json.Marshal(hello)
		logger.Debugw("sayHello Called", l.Fields{"DeviceID": client.DeviceID, "HelloMessage": js})
	}
	e := client.SendMessage(hello)
	if e != nil {
		logger.Fatalw("Failed saying hello to Openflow Server, unable to proceed",
			l.Fields{"DeviceID": client.DeviceID, "Error": e})
	}
}
func (client *Client) parseHeader(header ofp.IHeader, buf []byte) {
	switch header.GetType() {
	case ofp.OFPTHello:
		//x := header.(*ofp.Hello)
	case ofp.OFPTError:
		errMsg := header.(*ofp.ErrorMsg)
		go handleErrMsg(errMsg, client.DeviceID)
	case ofp.OFPTEchoRequest:
		echoReq := header.(*ofp.EchoRequest)
		go handleEchoRequest(echoReq, client.DeviceID, client)
	case ofp.OFPTEchoReply:
	case ofp.OFPTExperimenter:
	case ofp.OFPTFeaturesRequest:
		featReq := header.(*ofp.FeaturesRequest)
		go handleFeatureRequest(featReq, client.DeviceID, client)
	case ofp.OFPTFeaturesReply:
	case ofp.OFPTGetConfigRequest:
		configReq := header.(*ofp.GetConfigRequest)
		go handleGetConfigRequest(configReq, client.DeviceID, client)
	case ofp.OFPTGetConfigReply:
	case ofp.OFPTSetConfig:
		setConf := header.(*ofp.SetConfig)
		go handleSetConfig(setConf, client.DeviceID)
	case ofp.OFPTPacketIn:
	case ofp.OFPTFlowRemoved:
	case ofp.OFPTPortStatus:
	case ofp.OFPTPacketOut:
		packetOut := header.(*ofp.PacketOut)
		go handlePacketOut(packetOut, client.DeviceID)
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
			handleFlowAdd(flowAdd, client.DeviceID)
		case ofp.OFPFCModify:
			flowMod := header.(*ofp.FlowMod)
			handleFlowMod(flowMod, client.DeviceID)
		case ofp.OFPFCModifyStrict:
			flowModStrict := header.(*ofp.FlowModifyStrict)
			handleFlowModStrict(flowModStrict, client.DeviceID)
		case ofp.OFPFCDelete:
			flowDelete := header.(*ofp.FlowDelete)
			handleFlowDelete(flowDelete, client.DeviceID)
		case ofp.OFPFCDeleteStrict:
			flowDeleteStrict := header.(*ofp.FlowDeleteStrict)
			handleFlowDeleteStrict(flowDeleteStrict, client.DeviceID)
		}
	case ofp.OFPTStatsRequest:
		var statType = uint16(buf[8])<<8 + uint16(buf[9])
		go handleStatsRequest(header, statType, client.DeviceID, client)
	case ofp.OFPTBarrierRequest:
		/* See note above at case ofp.OFPTFlowMod:*/
		barRequest := header.(*ofp.BarrierRequest)
		handleBarrierRequest(barRequest, client.DeviceID, client)
	case ofp.OFPTRoleRequest:
		roleReq := header.(*ofp.RoleRequest)
		go handleRoleRequest(roleReq, client.DeviceID, client)
	case ofp.OFPTMeterMod:
		meterMod := header.(*ofp.MeterMod)
		go handleMeterModRequest(meterMod, client.DeviceID, client)

	}
}

//Message created to allow for a single SendMessage
type Message interface {
	Serialize(encoder *goloxi.Encoder) error
}

//SendMessage sends message to openflow server
func (client *Client) SendMessage(message Message) error {
	if settings.GetDebug(client.DeviceID) {
		js, _ := json.Marshal(message)
		logger.Debugw("SendMessage called", l.Fields{"DeviceID": client.DeviceID, "Message": js})
	}
	enc := goloxi.NewEncoder()
	message.Serialize(enc)
	for {
		if client.conn == nil {
			logger.Warnln("SendMessage Connection is Nil sleeping for 10 milliseconds")
			time.Sleep(10 * time.Millisecond)
		} else {
			break
		}
	}
	bytes := enc.Bytes()
	_, err := client.conn.Write(bytes)
	if err != nil {
		jMessage, _ := json.Marshal(message)
		logger.Errorw("SendMessage failed sending message", l.Fields{"DeviceID": client.DeviceID, "Error": err, "Message": jMessage})
		return err
	}
	return nil
}

//SetGrpcClient store a reference to the grpc client connection
func SetGrpcClient(client *pb.VolthaServiceClient) {
	volthaClient = client
}
func getGrpcClient() *pb.VolthaServiceClient {
	return volthaClient
}

//SetPacketOutChannel - store the channel to send packet outs to grpc connection
func SetPacketOutChannel(pktOutChan chan pb.PacketOut) {
	packetOutChannel = pktOutChan
}
