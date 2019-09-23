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

	pb "github.com/opencord/voltha-protos/go/voltha"
	"github.com/skydive-project/goloxi"
	ofp "github.com/skydive-project/goloxi/of13"

	"log"
	"net"
)

var conn net.Conn
var volthaClient *pb.VolthaServiceClient
var packetOutChannel chan pb.PacketOut

type Client struct {
	OfServerAddress string
	Port            uint16
	DeviceId        string
	KeepRunning     bool
}

func NewClient(ofServerAddress string, port uint16, deviceId string, keepRunning bool) *Client {
	client := Client{OfServerAddress: ofServerAddress, Port: port, DeviceId: deviceId, KeepRunning: keepRunning}
	go client.Start()
	return &client
}

func (client *Client) End() {
	client.KeepRunning = false
}
func (client *Client) Start() {
	addressLine := fmt.Sprintf("%s:%d", client.OfServerAddress, client.Port)
	raddr, e := net.ResolveTCPAddr("tcp", addressLine)
	if e != nil {
		errMessage := fmt.Errorf("Unable to resolve %s", addressLine)
		fmt.Println(errMessage)
	}
	for {
		if client.KeepRunning {
			connection, e := net.DialTCP("tcp", nil, raddr)
			conn = connection
			if e != nil {
				log.Fatalf("unable to connect to %v", raddr)
			}
			defer conn.Close()
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
			e = client.SendMessage(hello)
			if e != nil {
				log.Fatal(e)
			}
			for {
				buf := make([]byte, 1500)
				read, e := conn.Read(buf)
				if e != nil {
					log.Printf("Connection has died %v", e)
					break
				}
				decoder := goloxi.NewDecoder(buf)
				log.Printf("decoder offset %d baseOffset %d", decoder.Offset(), decoder.BaseOffset())
				header, e := ofp.DecodeHeader(decoder)
				log.Printf("received packet read: %d", read)

				if e != nil {
					log.Printf("decodeheader threw error %v", e)
				}
				message, _ := json.Marshal(header)
				log.Printf("message %s\n", message)
				client.parseHeader(header, buf) //first one is ready
				len := header.GetLength()
				if len < uint16(read) {
					log.Printf("Len %d was less than read %d", len, read)
					for {
						read = read - int(len)
						newBuf := buf[len:]
						decoder = goloxi.NewDecoder(newBuf)
						newHeader, e := ofp.DecodeHeader(decoder)
						message, _ := json.Marshal(newHeader)
						log.Printf("message %s\n", message)
						if e != nil {

						}
						len = newHeader.GetLength()
						client.parseHeader(newHeader, newBuf)
						if read == int(len) {
							log.Printf(" not continuing read %d len %d", read, len)
							break
						} else {
							log.Printf("continuing read %d len %d", read, len)
						}
						buf = newBuf
					}
				}
				if e != nil {
					log.Println("Decode Header error ", e)
				}
			}
		}
		break
	}
}

func (client *Client) parseHeader(header ofp.IHeader, buf []byte) {

	log.Printf("parseHeader called with type %d", header.GetType())
	switch header.GetType() {
	case 0:
		x := header.(*ofp.Hello)
		log.Printf("helloMessage : %+v", x)
		//nothing real to do
	case 1:
		errMsg := header.(*ofp.ErrorMsg)
		go handleErrMsg(errMsg)
	case 2:
		echoReq := header.(*ofp.EchoRequest)
		go handleEchoRequest(echoReq, client)
	case 3:
		//EchoReply
	case 4:
		//Expirementer
	case 5:
		featReq := header.(*ofp.FeaturesRequest)
		go handleFeatureRequest(featReq, client.DeviceId, client)
	case 6:
		//feature reply
	case 7:
		configReq := header.(*ofp.GetConfigRequest)
		go handleGetConfigRequest(configReq, client)
	case 8:
		//GetConfigReply
	case 9:
		setConf := header.(*ofp.SetConfig)
		go handleSetConfig(setConf)
	case 10:
		//packetIn := header.(*ofp.PacketIn)
		//go handlePacketIn(packetIn)
	case 11:
		//FlowRemoved
	case 12:
		//portStatus
	case 13:
		packetOut := header.(*ofp.PacketOut)
		go handlePacketOut(packetOut, client.DeviceId)
	case 14:
		flowModType := uint8(buf[25])
		switch flowModType {
		case 0:
			flowAdd := header.(*ofp.FlowAdd)
			go handleFlowAdd(flowAdd, client.DeviceId)
			//return DecodeFlowAdd(_flowmod, decoder)
		case 1:
			flowMod := header.(*ofp.FlowMod)
			go handleFlowMod(flowMod, client.DeviceId)
			//return DecodeFlowModify(_flowmod, decoder)
		case 2:
			flowModStrict := header.(*ofp.FlowModifyStrict)
			go handleFlowModStrict(flowModStrict, client.DeviceId)
			//return DecodeFlowModifyStrict(_flowmod, decoder)
		case 3:
			flowDelete := header.(*ofp.FlowDelete)
			go handleFlowDelete(flowDelete, client.DeviceId)
			//return DecodeFlowDelete(_flowmod, decoder)
		case 4:
			flowDeleteStrict := header.(*ofp.FlowDeleteStrict)
			go handleFlowDeleteStrict(flowDeleteStrict, client.DeviceId)
			//return DecodeFlowDeleteStrict(_flowmod, decoder)
		default:
			//return nil, fmt.Errorf("Invalid type '%d' for 'FlowMod'", _flowmod.Command)

		}
	case 18:
		var statType = uint16(buf[8])<<8 + uint16(buf[9])
		log.Println("statsType", statType)
		go handleStatsRequest(header, statType, client.DeviceId, client)
	case 20:
		barRequest := header.(*ofp.BarrierRequest)
		go handleBarrierRequest(barRequest, client)
	case 24:
		roleReq := header.(*ofp.RoleRequest)
		go handleRoleRequest(roleReq, client)

	}
}

//openFlowMessage created to allow for a single SendMessage
type OpenFlowMessage interface {
	Serialize(encoder *goloxi.Encoder) error
}

func (client *Client) SendMessage(message OpenFlowMessage) error {
	jsonMessage, _ := json.Marshal(message)
	log.Printf("send message called %s\n", jsonMessage)
	enc := goloxi.NewEncoder()
	message.Serialize(enc)
	jMessage, _ := json.Marshal(message)
	log.Printf("message after serialize %s", jMessage)

	for {
		if conn == nil {
			log.Println("CONN IS NIL")
			time.Sleep(10 * time.Millisecond)
		} else {
			break
		}
	}
	_, err := conn.Write(enc.Bytes())
	if err != nil {
		log.Printf("SendMessage had error %s", err)
		return err
	}
	return nil
}
func SetGrpcClient(client *pb.VolthaServiceClient) {
	volthaClient = client
}
func getGrpcClient() *pb.VolthaServiceClient {
	return volthaClient
}
func test(hello ofp.IHello) {
	fmt.Println("works")
}
func SetPacketOutChannel(pktOutChan chan pb.PacketOut) {
	packetOutChannel = pktOutChan
}
