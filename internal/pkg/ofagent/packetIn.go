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

package ofagent

import (
	"context"
	"encoding/json"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/goloxi"
	ofp "github.com/opencord/goloxi/of13"
	"github.com/opencord/ofagent-go/internal/pkg/openflow"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-protos/v3/go/openflow_13"
	"github.com/opencord/voltha-protos/v3/go/voltha"
	"google.golang.org/grpc"
)

func (ofa *OFAgent) receivePacketsIn(ctx context.Context) {
	logger.Debug(ctx, "receive-packets-in-started")
	// If we exit, assume disconnected
	defer func() {
		ofa.events <- ofaEventVolthaDisconnected
		logger.Debug(ctx, "receive-packets-in-finished")
	}()
	if ofa.volthaClient == nil {
		logger.Error(ctx, "no-voltha-connection")
		return
	}
	opt := grpc.EmptyCallOption{}
	streamCtx, streamDone := context.WithCancel(context.Background())
	defer streamDone()
	stream, err := ofa.volthaClient.Get().ReceivePacketsIn(streamCtx, &empty.Empty{}, opt)
	if err != nil {
		logger.Errorw(ctx, "Unable to establish Receive PacketIn Stream",
			log.Fields{"error": err})
		return
	}

top:

	for {
		select {
		case <-ctx.Done():
			break top
		default:
			pkt, err := stream.Recv()
			if err != nil {
				logger.Errorw(ctx, "error receiving packet",
					log.Fields{"error": err})
				break top
			}
			ofa.packetInChannel <- pkt
		}
	}
}

func (ofa *OFAgent) handlePacketsIn(ctx context.Context) {
	logger.Debug(ctx, "handle-packets-in-started")
top:
	for {
		select {
		case <-ctx.Done():
			break top
		case packet := <-ofa.packetInChannel:
			packetIn := packet.GetPacketIn()

			if logger.V(log.DebugLevel) {
				js, _ := json.Marshal(packetIn)
				logger.Debugw(ctx, "packet-in received", log.Fields{"packet-in": js})
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
			var fields []goloxi.IOxm
			for _, oxmField := range inMatch.GetOxmFields() {
				field := oxmField.GetField()
				ofbField := field.(*openflow_13.OfpOxmField_OfbField).OfbField
				switch ofbField.Type {
				case voltha.OxmOfbFieldTypes_OFPXMT_OFB_IN_PORT:
					ofpInPort := ofp.NewOxmInPort()
					val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_Port)
					ofpInPort.Value = ofp.Port(val.Port)
					fields = append(fields, ofpInPort)
				case voltha.OxmOfbFieldTypes_OFPXMT_OFB_ETH_TYPE:
					ofpEthType := ofp.NewOxmEthType()
					val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_EthType)
					ofpEthType.Value = ofp.EthernetType(val.EthType)
					fields = append(fields, ofpEthType)
				case voltha.OxmOfbFieldTypes_OFPXMT_OFB_IN_PHY_PORT:
					ofpInPhyPort := ofp.NewOxmInPhyPort()
					val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_PhysicalPort)
					ofpInPhyPort.Value = ofp.Port(val.PhysicalPort)
					fields = append(fields, ofpInPhyPort)
				case voltha.OxmOfbFieldTypes_OFPXMT_OFB_IP_PROTO:
					ofpIpProto := ofp.NewOxmIpProto()
					val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_IpProto)
					ofpIpProto.Value = ofp.IpPrototype(val.IpProto)
					fields = append(fields, ofpIpProto)
				case voltha.OxmOfbFieldTypes_OFPXMT_OFB_UDP_SRC:
					ofpUdpSrc := ofp.NewOxmUdpSrc()
					val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_UdpSrc)
					ofpUdpSrc.Value = uint16(val.UdpSrc)
					fields = append(fields, ofpUdpSrc)
				case voltha.OxmOfbFieldTypes_OFPXMT_OFB_UDP_DST:
					ofpUdpDst := ofp.NewOxmUdpDst()
					val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_UdpDst)
					ofpUdpDst.Value = uint16(val.UdpDst)
					fields = append(fields, ofpUdpDst)
				case voltha.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID:
					ofpVlanVid := ofp.NewOxmVlanVid()
					val := ofbField.GetValue()
					if val != nil {
						vlanId := val.(*openflow_13.OfpOxmOfbField_VlanVid)
						ofpVlanVid.Value = uint16(vlanId.VlanVid) + 0x1000
					} else {
						ofpVlanVid.Value = uint16(0)
					}

					fields = append(fields, ofpVlanVid)
				default:
					logger.Warnw(ctx, "receive-packet-in:unhandled-oxm-field",
						log.Fields{"field": ofbField.Type})
				}
			}

			match.SetOxmList(fields)

			ofPacketIn.SetMatch(*match)
			ofPacketIn.SetReason(uint8(packetIn.GetReason()))
			ofPacketIn.SetTableId(uint8(packetIn.GetTableId()))
			ofc := ofa.getOFClient(ctx, deviceID)
			if err := ofc.SendMessage(ctx, ofPacketIn); err != nil {
				logger.Errorw(ctx, "send-message-failed", log.Fields{
					"device-id": deviceID,
					"error":     err})
			}

		}
	}
	logger.Debug(ctx, "handle-packets-in-finished")
}
