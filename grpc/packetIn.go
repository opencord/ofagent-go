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

package grpc

import (
	"context"
	"github.com/donNewtonAlpha/goloxi"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/ofagent-go/openflow"
	"github.com/opencord/voltha-protos/go/openflow_13"
	pb "github.com/opencord/voltha-protos/go/voltha"
	"google.golang.org/grpc"
	"log"
)

func receivePacketIn(client pb.VolthaServiceClient) {
	opt := grpc.EmptyCallOption{}
	stream, err := client.ReceivePacketsIn(context.Background(), &empty.Empty{}, opt)
	if err != nil {
		log.Fatalln("Unable to establish stream")
	}
	for {
		packet, err := stream.Recv()
		packetIn := packet.GetPacketIn()

		if err != nil {
			log.Fatalf("error on stream.Rec %v", err)
		}
		deviceID := packet.GetId()
		ofPacketIn := ofp.NewPacketIn()
		ofPacketIn.SetVersion(uint8(4))
		ofPacketIn.SetXid(openflow.GetXid())
		ofPacketIn.SetBufferId(packetIn.GetBufferId())
		ofPacketIn.SetCookie(packetIn.GetCookie())
		ofPacketIn.SetData(packetIn.GetData())
		match := ofp.NewMatchV3()
		inMatch := packetIn.GetMatch()
		match.SetType(uint16(inMatch.GetType()))
		oxFields := inMatch.GetOxmFields()
		var fields []goloxi.IOxm
		var size uint16
		size = 4
		for i := 0; i < len(oxFields); i++ {
			oxmField := oxFields[i]
			field := oxmField.GetField()
			ofbField := field.(*openflow_13.OfpOxmField_OfbField).OfbField
			size += 4 //header for oxm
			switch ofbField.Type {
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_IN_PORT:
				ofpInPort := ofp.NewOxmInPort()
				val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_Port)
				ofpInPort.Value = ofp.Port(val.Port)
				size += 4
				fields = append(fields, ofpInPort)
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_ETH_TYPE:
				ofpEthType := ofp.NewOxmEthType()
				val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_EthType)
				ofpEthType.Value = ofp.EthernetType(val.EthType)
				size += 2
				fields = append(fields, ofpEthType)
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_IN_PHY_PORT:
				ofpInPhyPort := ofp.NewOxmInPhyPort()
				val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_PhysicalPort)
				ofpInPhyPort.Value = ofp.Port(val.PhysicalPort)
				size += 4
				fields = append(fields, ofpInPhyPort)
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_IP_PROTO:
				ofpIpProto := ofp.NewOxmIpProto()
				val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_IpProto)
				ofpIpProto.Value = ofp.IpPrototype(val.IpProto)
				size += 1
				fields = append(fields, ofpIpProto)
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_UDP_SRC:
				ofpUdpSrc := ofp.NewOxmUdpSrc()
				val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_UdpSrc)
				ofpUdpSrc.Value = uint16(val.UdpSrc)
				size += 2
				fields = append(fields, ofpUdpSrc)
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_UDP_DST:
				ofpUdpDst := ofp.NewOxmUdpDst()
				val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_UdpDst)
				ofpUdpDst.Value = uint16(val.UdpDst)
				size += 2
				fields = append(fields, ofpUdpDst)
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID:
				ofpVlanVid := ofp.NewOxmVlanVid()
				val := ofbField.GetValue()
				if val != nil {
					vlanId := val.(*openflow_13.OfpOxmOfbField_VlanVid)
					ofpVlanVid.Value = uint16(vlanId.VlanVid) + 0x1000
					size += 2
				} else {
					ofpVlanVid.Value = uint16(0)
				}

				fields = append(fields, ofpVlanVid)
			default:
				log.Printf("handleFlowStatsRequest   Unhandled OxmField %v", ofbField.Type)
			}
		}
		match.SetLength(size)

		match.SetOxmList(fields)

		ofPacketIn.SetMatch(*match)
		ofPacketIn.SetReason(uint8(packetIn.GetReason()))
		ofPacketIn.SetTableId(uint8(packetIn.GetTableId()))
		ofPacketIn.SetTotalLen(uint16(len(ofPacketIn.GetData())))
		openFlowClient := GetClient(deviceID)
		openFlowClient.SendMessage(ofPacketIn)

	}

}