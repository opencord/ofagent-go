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
	"net"

	ofp "github.com/opencord/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-protos/v5/go/openflow_13"
)

var oxmMap = map[string]int32{
	"in_port":         0,
	"in_phy_port":     1,
	"metadata":        2,
	"eth_dst":         3,
	"eth_src":         4,
	"eth_type":        5,
	"vlan_vid":        6,
	"vlan_pcp":        7,
	"ip_dscp":         8,
	"ip_ecn":          9,
	"ip_proto":        10,
	"ipv4_src":        11,
	"ipv4_dst":        12,
	"tcp_src":         13,
	"tcp_dst":         14,
	"udp_src":         15,
	"udp_dst":         16,
	"sctp_src":        17,
	"sctp_dst":        18,
	"icmpv4_type":     19,
	"icmpv4_code":     20,
	"arp_op":          21,
	"arp_spa":         22,
	"arp_tpa":         23,
	"arp_sha":         24,
	"arp_tha":         25,
	"ipv6_src":        26,
	"ipv6_dst":        27,
	"ipv6_flabel":     28,
	"icmpv6_type":     29,
	"icmpv6_code":     30,
	"ipv6_nd_target":  31,
	"ipv6_nd_sll":     32,
	"ipv6_nd_tll":     33,
	"mpls_label":      34,
	"mpls_tc":         35,
	"mpls_bos":        36,
	"pbb_isid":        37,
	"tunnel_id":       38,
	"ipv6_exthdr":     39,
	"vlan_vid_masked": 200, //made up
}

func (ofc *OFConnection) handleFlowAdd(ctx context.Context, flowAdd *ofp.FlowAdd) {
	span, ctx := log.CreateChildSpan(ctx, "openflow-flow-add")
	defer span.Finish()

	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "handleFlowAdd called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"flow":      flowAdd})
	}

	volthaClient := ofc.VolthaClient.Get()
	if volthaClient == nil {
		logger.Errorw(ctx, "no-voltha-connection",
			log.Fields{"device-id": ofc.DeviceID})
		return
	}

	// Construct the match
	var oxmList []*openflow_13.OfpOxmField
	for _, oxmField := range flowAdd.Match.GetOxmList() {
		name := oxmMap[oxmField.GetOXMName()]
		val := oxmField.GetOXMValue()
		field := openflow_13.OfpOxmOfbField{Type: openflow_13.OxmOfbFieldTypes(name)}
		ofpOxmField := openflow_13.OfpOxmField{
			OxmClass: ofp.OFPXMCOpenflowBasic,
			Field:    &openflow_13.OfpOxmField_OfbField{OfbField: &field},
		}
		switch openflow_13.OxmOfbFieldTypes(name) {
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IN_PORT:
			field.Value = &openflow_13.OfpOxmOfbField_Port{
				Port: uint32(val.(ofp.Port)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IN_PHY_PORT:
			field.Value = &openflow_13.OfpOxmOfbField_PhysicalPort{
				PhysicalPort: val.(uint32),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_METADATA:
			field.Value = &openflow_13.OfpOxmOfbField_TableMetadata{
				TableMetadata: val.(uint64),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_ETH_TYPE:
			field.Value = &openflow_13.OfpOxmOfbField_EthType{
				EthType: uint32(val.(ofp.EthernetType)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IP_PROTO:
			field.Value = &openflow_13.OfpOxmOfbField_IpProto{
				IpProto: uint32(val.(ofp.IpPrototype)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IPV4_DST:
			field.Value = &openflow_13.OfpOxmOfbField_Ipv4Dst{
				Ipv4Dst: binary.BigEndian.Uint32(val.(net.IP)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_ETH_DST:
			field.Value = &openflow_13.OfpOxmOfbField_EthDst{
				EthDst: val.(net.HardwareAddr),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_ETH_SRC:
			field.Value = &openflow_13.OfpOxmOfbField_EthSrc{
				EthSrc: val.(net.HardwareAddr),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_UDP_SRC:
			field.Value = &openflow_13.OfpOxmOfbField_UdpSrc{
				UdpSrc: uint32(val.(uint16)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_UDP_DST:
			field.Value = &openflow_13.OfpOxmOfbField_UdpDst{
				UdpDst: uint32(val.(uint16)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID:
			field.Value = &openflow_13.OfpOxmOfbField_VlanVid{
				VlanVid: uint32((val.(uint16) & 0xfff) | 0x1000),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_PCP:
			field.Value = &openflow_13.OfpOxmOfbField_VlanPcp{
				VlanPcp: uint32(val.(uint8)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_MPLS_LABEL:
			field.Value = &openflow_13.OfpOxmOfbField_MplsLabel{
				MplsLabel: val.(uint32),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_MPLS_BOS:
			field.Value = &openflow_13.OfpOxmOfbField_MplsBos{
				MplsBos: uint32(val.(uint8)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_MPLS_TC:
			field.Value = &openflow_13.OfpOxmOfbField_MplsTc{
				MplsTc: val.(uint32),
			}
		case 200: // voltha-protos doesn't actually have a type for vlan_mask
			field = openflow_13.OfpOxmOfbField{Type: openflow_13.OxmOfbFieldTypes(oxmMap["vlan_vid"])}
			field.HasMask = true
			ofpOxmField = openflow_13.OfpOxmField{
				OxmClass: ofp.OFPXMCOpenflowBasic,
				Field:    &openflow_13.OfpOxmField_OfbField{OfbField: &field},
			}
			field.Value = &openflow_13.OfpOxmOfbField_VlanVid{
				VlanVid: uint32(val.(uint16)),
			}
			vidMask := val.(uint16)
			field.Mask = &openflow_13.OfpOxmOfbField_VlanVidMask{
				VlanVidMask: uint32(vidMask),
			}
		}
		oxmList = append(oxmList, &ofpOxmField)
	}

	// Construct the instructions
	var instructions []*openflow_13.OfpInstruction
	for _, ofpInstruction := range flowAdd.GetInstructions() {
		instructionType := ofpInstruction.GetType()
		instruction := openflow_13.OfpInstruction{Type: uint32(instructionType)}
		switch instructionType {
		case ofp.OFPITGotoTable:
			instruction.Data = &openflow_13.OfpInstruction_GotoTable{
				GotoTable: &openflow_13.OfpInstructionGotoTable{
					TableId: uint32(ofpInstruction.(ofp.IInstructionGotoTable).GetTableId()),
				},
			}
		case ofp.OFPITWriteMetadata:
			instruction.Data = &openflow_13.OfpInstruction_WriteMetadata{
				WriteMetadata: &openflow_13.OfpInstructionWriteMetadata{
					Metadata:     ofpInstruction.(ofp.IInstructionWriteMetadata).GetMetadata(),
					MetadataMask: ofpInstruction.(ofp.IInstructionWriteMetadata).GetMetadataMask(),
				},
			}
		case ofp.OFPITWriteActions:
			var ofpActions []*openflow_13.OfpAction
			for _, action := range ofpInstruction.(ofp.IInstructionWriteActions).GetActions() {
				ofpActions = append(ofpActions, extractAction(action))
			}
			instruction.Data = &openflow_13.OfpInstruction_Actions{
				Actions: &openflow_13.OfpInstructionActions{
					Actions: ofpActions,
				},
			}
		case ofp.OFPITApplyActions:
			var ofpActions []*openflow_13.OfpAction
			for _, action := range ofpInstruction.(ofp.IInstructionApplyActions).GetActions() {
				ofpActions = append(ofpActions, extractAction(action))
			}
			instruction.Data = &openflow_13.OfpInstruction_Actions{
				Actions: &openflow_13.OfpInstructionActions{
					Actions: ofpActions,
				},
			}
		case ofp.OFPITMeter:
			instruction.Data = &openflow_13.OfpInstruction_Meter{
				Meter: &openflow_13.OfpInstructionMeter{
					MeterId: ofpInstruction.(ofp.IInstructionMeter).GetMeterId(),
				},
			}
		}
		instructions = append(instructions, &instruction)
	}

	// Construct the request
	flowUpdate := openflow_13.FlowTableUpdate{
		Id: ofc.DeviceID,
		FlowMod: &openflow_13.OfpFlowMod{
			Cookie:      flowAdd.Cookie,
			CookieMask:  flowAdd.CookieMask,
			TableId:     uint32(flowAdd.TableId),
			Command:     openflow_13.OfpFlowModCommand_OFPFC_ADD,
			IdleTimeout: uint32(flowAdd.IdleTimeout),
			HardTimeout: uint32(flowAdd.HardTimeout),
			Priority:    uint32(flowAdd.Priority),
			BufferId:    flowAdd.BufferId,
			OutPort:     uint32(flowAdd.OutPort),
			OutGroup:    uint32(flowAdd.OutGroup),
			Flags:       uint32(flowAdd.Flags),
			Match: &openflow_13.OfpMatch{
				Type:      openflow_13.OfpMatchType(flowAdd.Match.GetType()),
				OxmFields: oxmList,
			},

			Instructions: instructions,
		},
		Xid: flowAdd.Xid,
	}
	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "FlowAdd being sent to Voltha",
			log.Fields{
				"device-id":       ofc.DeviceID,
				"flow-mod-object": flowUpdate,
			})
	}
	if _, err := volthaClient.UpdateLogicalDeviceFlowTable(log.WithSpanFromContext(context.Background(), ctx), &flowUpdate); err != nil {
		logger.Errorw(ctx, "Error calling FlowAdd ",
			log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
		// Report failure to controller
		message := ofp.NewFlowModFailedErrorMsg()
		message.SetXid(flowAdd.Xid)
		message.SetCode(ofp.OFPFMFCBadCommand)
		//OF 1.3
		message.SetVersion(4)
		bs := make([]byte, 2)
		//OF 1.3
		bs[0] = byte(4)
		//Flow Mod
		bs[1] = byte(14)
		//Length of the message
		length := make([]byte, 2)
		binary.BigEndian.PutUint16(length, 56)
		bs = append(bs, length...)
		empty := []byte{0, 0, 0, 0}
		bs = append(bs, empty...)
		//Cookie of the Flow
		cookie := make([]byte, 52)
		binary.BigEndian.PutUint64(cookie, flowAdd.Cookie)
		bs = append(bs, cookie...)
		message.SetData(bs)
		err := ofc.SendMessage(ctx, message)
		if err != nil {
			logger.Errorw(ctx, "Error reporting failure of FlowUpdate to controller",
				log.Fields{
					"device-id": ofc.DeviceID,
					"error":     err})
		}
	}
}

func (ofc *OFConnection) handleFlowMod(ctx context.Context, flowMod *ofp.FlowMod) {
	span, ctx := log.CreateChildSpan(ctx, "openflow-flow-modification")
	defer span.Finish()

	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "handleFlowMod called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"flow-mod":  flowMod})
	}
	logger.Errorw(ctx, "handleFlowMod not implemented",
		log.Fields{"device-id": ofc.DeviceID})
}

func (ofc *OFConnection) handleFlowModStrict(ctx context.Context, flowModStrict *ofp.FlowModifyStrict) {
	span, ctx := log.CreateChildSpan(ctx, "openflow-flow-modification-strict")
	defer span.Finish()

	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "handleFlowModStrict called",
			log.Fields{
				"device-id":       ofc.DeviceID,
				"flow-mod-strict": flowModStrict})
	}
	logger.Error(ctx, "handleFlowModStrict not implemented",
		log.Fields{"device-id": ofc.DeviceID})
}

func (ofc *OFConnection) handleFlowDelete(ctx context.Context, flowDelete *ofp.FlowDelete) {
	span, ctx := log.CreateChildSpan(ctx, "openflow-flow-delete")
	defer span.Finish()

	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "handleFlowDelete called",
			log.Fields{
				"device-id":   ofc.DeviceID,
				"flow-delete": flowDelete})
	}
	logger.Error(ctx, "handleFlowDelete not implemented",
		log.Fields{"device-id": ofc.DeviceID})

}

func (ofc *OFConnection) handleFlowDeleteStrict(ctx context.Context, flowDeleteStrict *ofp.FlowDeleteStrict) {
	span, ctx := log.CreateChildSpan(ctx, "openflow-flow-delete-strict")
	defer span.Finish()

	if logger.V(log.DebugLevel) {
		logger.Debugw(ctx, "handleFlowDeleteStrict called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"flow":      flowDeleteStrict})
	}

	volthaClient := ofc.VolthaClient.Get()
	if volthaClient == nil {
		logger.Errorw(ctx, "no-voltha-connection",
			log.Fields{"device-id": ofc.DeviceID})
		return
	}

	// Construct match
	var oxmList []*openflow_13.OfpOxmField
	for _, oxmField := range flowDeleteStrict.Match.GetOxmList() {
		name := oxmMap[oxmField.GetOXMName()]
		val := oxmField.GetOXMValue()
		var ofpOxmField openflow_13.OfpOxmField
		ofpOxmField.OxmClass = ofp.OFPXMCOpenflowBasic
		var field openflow_13.OfpOxmOfbField
		field.Type = openflow_13.OxmOfbFieldTypes(name)

		var x openflow_13.OfpOxmField_OfbField
		x.OfbField = &field
		ofpOxmField.Field = &x

		switch openflow_13.OxmOfbFieldTypes(name) {
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IN_PORT:
			field.Value = &openflow_13.OfpOxmOfbField_Port{
				Port: uint32(val.(ofp.Port)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IN_PHY_PORT:
			field.Value = &openflow_13.OfpOxmOfbField_PhysicalPort{
				PhysicalPort: val.(uint32),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_METADATA:
			field.Value = &openflow_13.OfpOxmOfbField_TableMetadata{
				TableMetadata: val.(uint64),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_ETH_TYPE:
			field.Value = &openflow_13.OfpOxmOfbField_EthType{
				EthType: uint32(val.(ofp.EthernetType)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IP_PROTO:
			field.Value = &openflow_13.OfpOxmOfbField_IpProto{
				IpProto: uint32(val.(ofp.IpPrototype)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IPV4_DST:
			field.Value = &openflow_13.OfpOxmOfbField_Ipv4Dst{
				Ipv4Dst: binary.BigEndian.Uint32(val.(net.IP)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_ETH_DST:
			field.Value = &openflow_13.OfpOxmOfbField_EthDst{
				EthDst: val.(net.HardwareAddr),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_ETH_SRC:
			field.Value = &openflow_13.OfpOxmOfbField_EthSrc{
				EthSrc: val.(net.HardwareAddr),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_UDP_SRC:
			field.Value = &openflow_13.OfpOxmOfbField_UdpSrc{
				UdpSrc: uint32(val.(uint16)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_UDP_DST:
			field.Value = &openflow_13.OfpOxmOfbField_UdpDst{
				UdpDst: uint32(val.(uint16)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID:
			field.Value = &openflow_13.OfpOxmOfbField_VlanVid{
				VlanVid: uint32(val.(uint16)),
			}
		case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_PCP:
			field.Value = &openflow_13.OfpOxmOfbField_VlanPcp{
				VlanPcp: uint32(val.(uint8)),
			}
		case 200: // voltha-protos doesn't actually have a type for vlan_mask
			field = openflow_13.OfpOxmOfbField{Type: openflow_13.OxmOfbFieldTypes(oxmMap["vlan_vid"])}
			field.HasMask = true
			ofpOxmField = openflow_13.OfpOxmField{
				OxmClass: ofp.OFPXMCOpenflowBasic,
				Field:    &openflow_13.OfpOxmField_OfbField{OfbField: &field},
			}
			field.Value = &openflow_13.OfpOxmOfbField_VlanVid{
				VlanVid: uint32(val.(uint16)),
			}
			vidMask := val.(uint16)
			field.Mask = &openflow_13.OfpOxmOfbField_VlanVidMask{
				VlanVidMask: uint32(vidMask),
			}
		}

		oxmList = append(oxmList, &ofpOxmField)
	}

	responseRequired := false

	if flowDeleteStrict.GetFlags() == ofp.OFPFFSendFlowRem {
		responseRequired = true
	}

	// Construct request
	flowUpdate := openflow_13.FlowTableUpdate{
		Id: ofc.DeviceID,
		FlowMod: &openflow_13.OfpFlowMod{
			Cookie:      flowDeleteStrict.Cookie,
			CookieMask:  flowDeleteStrict.CookieMask,
			TableId:     uint32(flowDeleteStrict.TableId),
			Command:     openflow_13.OfpFlowModCommand_OFPFC_DELETE_STRICT,
			IdleTimeout: uint32(flowDeleteStrict.IdleTimeout),
			HardTimeout: uint32(flowDeleteStrict.HardTimeout),
			Priority:    uint32(flowDeleteStrict.Priority),
			BufferId:    flowDeleteStrict.BufferId,
			OutPort:     uint32(flowDeleteStrict.OutPort),
			OutGroup:    uint32(flowDeleteStrict.OutGroup),
			Flags:       uint32(flowDeleteStrict.Flags),
			Match: &openflow_13.OfpMatch{
				Type:      openflow_13.OfpMatchType(flowDeleteStrict.Match.GetType()),
				OxmFields: oxmList,
			},
		},
		Xid: flowDeleteStrict.Xid,
	}

	if logger.V(log.DebugLevel) {
		logger.Debugf(ctx, "FlowDeleteStrict being sent to Voltha",
			log.Fields{
				"device-id":   ofc.DeviceID,
				"flow-update": flowUpdate})
	}

	if _, err := volthaClient.UpdateLogicalDeviceFlowTable(log.WithSpanFromContext(context.Background(), ctx), &flowUpdate); err != nil {
		logger.Errorw(ctx, "Error calling FlowDelete ",
			log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
		return
	}

	if responseRequired {
		response := ofp.NewFlowRemoved()

		response.Cookie = flowDeleteStrict.Cookie
		response.Priority = flowDeleteStrict.Priority
		response.Reason = ofp.OFPRRDelete
		response.Match = flowDeleteStrict.Match
		response.IdleTimeout = flowDeleteStrict.IdleTimeout
		response.HardTimeout = flowDeleteStrict.HardTimeout
		response.Xid = flowDeleteStrict.Xid

		err := ofc.SendMessage(ctx, response)
		if err != nil {
			logger.Errorw(ctx, "Error sending FlowRemoved to ONOS",
				log.Fields{
					"device-id": ofc.DeviceID,
					"error":     err})
		}
	}

}
