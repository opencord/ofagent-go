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
	"encoding/json"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-protos/v3/go/openflow_13"
	"github.com/opencord/voltha-protos/v3/go/voltha"
)

var oxmMap = map[string]int32{
	"in_port":        0,
	"in_phy_port":    1,
	"metadata":       2,
	"eth_dst":        3,
	"eth_src":        4,
	"eth_type":       5,
	"vlan_vid":       6,
	"vlan_pcp":       7,
	"ip_dscp":        8,
	"ip_ecn":         9,
	"ip_proto":       10,
	"ipv4_src":       11,
	"ipv4_dst":       12,
	"tcp_src":        13,
	"tcp_dst":        14,
	"udp_src":        15,
	"udp_dst":        16,
	"sctp_src":       17,
	"sctp_dst":       18,
	"icmpv4_type":    19,
	"icmpv4_code":    20,
	"arp_op":         21,
	"arp_spa":        22,
	"arp_tpa":        23,
	"arp_sha":        24,
	"arp_tha":        25,
	"ipv6_src":       26,
	"ipv6_dst":       27,
	"ipv6_flabel":    28,
	"icmpv6_type":    29,
	"icmpv6_code":    30,
	"ipv6_nd_target": 31,
	"ipv6_nd_sll":    32,
	"ipv6_nd_tll":    33,
	"mpls_label":     34,
	"mpls_tc":        35,
	"mpls_bos":       36,
	"pbb_isid":       37,
	"tunnel_id":      38,
	"ipv6_exthdr":    39,
}

func (ofc *OFClient) handleFlowAdd(flowAdd *ofp.FlowAdd) {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(flowAdd)
		logger.Debugw("handleFlowAdd called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"params":    js})
	}

	// Construct the match
	var oxmList []*voltha.OfpOxmField
	for _, oxmField := range flowAdd.Match.GetOxmList() {
		name := oxmMap[oxmField.GetOXMName()]
		val := oxmField.GetOXMValue()
		field := voltha.OfpOxmOfbField{Type: voltha.OxmOfbFieldTypes(name)}
		ofpOxmField := voltha.OfpOxmField{
			OxmClass: ofp.OFPXMCOpenflowBasic,
			Field:    &openflow_13.OfpOxmField_OfbField{OfbField: &field},
		}
		switch voltha.OxmOfbFieldTypes(name) {
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_IN_PORT:
			field.Value = &voltha.OfpOxmOfbField_Port{
				Port: uint32(val.(ofp.Port)),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_IN_PHY_PORT:
			field.Value = &voltha.OfpOxmOfbField_PhysicalPort{
				PhysicalPort: val.(uint32),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_METADATA:
			field.Value = &voltha.OfpOxmOfbField_TableMetadata{
				TableMetadata: val.(uint64),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_ETH_TYPE:
			field.Value = &voltha.OfpOxmOfbField_EthType{
				EthType: uint32(val.(ofp.EthernetType)),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_IP_PROTO:
			field.Value = &voltha.OfpOxmOfbField_IpProto{
				IpProto: uint32(val.(ofp.IpPrototype)),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_UDP_SRC:
			field.Value = &voltha.OfpOxmOfbField_UdpSrc{
				UdpSrc: uint32(val.(uint16)),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_UDP_DST:
			field.Value = &voltha.OfpOxmOfbField_UdpDst{
				UdpDst: uint32(val.(uint16)),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID:
			field.Value = &voltha.OfpOxmOfbField_VlanVid{
				VlanVid: uint32((val.(uint16) & 0xfff) | 0x1000),
			}
		}
		oxmList = append(oxmList, &ofpOxmField)
	}

	// Construct the instructions
	var instructions []*voltha.OfpInstruction
	for _, ofpInstruction := range flowAdd.GetInstructions() {
		instructionType := ofpInstruction.GetType()
		instruction := voltha.OfpInstruction{Type: uint32(instructionType)}
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
		FlowMod: &voltha.OfpFlowMod{
			Cookie:      flowAdd.Cookie,
			CookieMask:  flowAdd.CookieMask,
			TableId:     uint32(flowAdd.TableId),
			Command:     voltha.OfpFlowModCommand_OFPFC_ADD,
			IdleTimeout: uint32(flowAdd.IdleTimeout),
			HardTimeout: uint32(flowAdd.HardTimeout),
			Priority:    uint32(flowAdd.Priority),
			BufferId:    flowAdd.BufferId,
			OutPort:     uint32(flowAdd.OutPort),
			OutGroup:    uint32(flowAdd.OutGroup),
			Flags:       uint32(flowAdd.Flags),
			Match: &voltha.OfpMatch{
				Type:      voltha.OfpMatchType(flowAdd.Match.GetType()),
				OxmFields: oxmList,
			},

			Instructions: instructions,
		},
	}
	if logger.V(log.DebugLevel) {
		flowUpdateJs, _ := json.Marshal(flowUpdate)
		logger.Debugf("FlowUpdate being sent to Voltha",
			log.Fields{
				"device-id":        ofc.DeviceID,
				"flow-mod-request": flowUpdateJs})
	}
	if _, err := ofc.VolthaClient.UpdateLogicalDeviceFlowTable(context.Background(), &flowUpdate); err != nil {
		logger.Errorw("Error calling FlowUpdate ",
			log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
	}
}

func (ofc *OFClient) handleFlowMod(flowMod *ofp.FlowMod) {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(flowMod)
		logger.Debugw("handleMod called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"flow-mod":  js})
	}
	logger.Errorw("handleFlowMod not implemented",
		log.Fields{"device-id": ofc.DeviceID})
}

func (ofc *OFClient) handleFlowModStrict(flowModStrict *ofp.FlowModifyStrict) {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(flowModStrict)
		logger.Debugw("handleFlowModStrict called",
			log.Fields{
				"device-id":       ofc.DeviceID,
				"flow-mod-strict": js})
	}
	logger.Error("handleFlowModStrict not implemented",
		log.Fields{"device-id": ofc.DeviceID})
}

func (ofc *OFClient) handleFlowDelete(flowDelete *ofp.FlowDelete) {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(flowDelete)
		logger.Debugw("handleFlowDelete called",
			log.Fields{
				"device-id":   ofc.DeviceID,
				"flow-delete": js})
	}
	logger.Error("handleFlowDelete not implemented",
		log.Fields{"device-id": ofc.DeviceID})

}

func (ofc *OFClient) handleFlowDeleteStrict(flowDeleteStrict *ofp.FlowDeleteStrict) {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(flowDeleteStrict)
		logger.Debugw("handleFlowAdd called",
			log.Fields{
				"device-id":          ofc.DeviceID,
				"flow-delete-strict": js})
	}

	// Construct match
	var oxmList []*voltha.OfpOxmField
	for _, oxmField := range flowDeleteStrict.Match.GetOxmList() {
		name := oxmMap[oxmField.GetOXMName()]
		val := oxmField.GetOXMValue()
		var ofpOxmField voltha.OfpOxmField
		ofpOxmField.OxmClass = ofp.OFPXMCOpenflowBasic
		var field voltha.OfpOxmOfbField
		field.Type = voltha.OxmOfbFieldTypes(name)

		var x openflow_13.OfpOxmField_OfbField
		x.OfbField = &field
		ofpOxmField.Field = &x

		switch voltha.OxmOfbFieldTypes(name) {
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_IN_PORT:
			field.Value = &voltha.OfpOxmOfbField_Port{
				Port: uint32(val.(ofp.Port)),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_IN_PHY_PORT:
			field.Value = &voltha.OfpOxmOfbField_PhysicalPort{
				PhysicalPort: val.(uint32),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_METADATA:
			field.Value = &voltha.OfpOxmOfbField_TableMetadata{
				TableMetadata: val.(uint64),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_ETH_TYPE:
			field.Value = &voltha.OfpOxmOfbField_EthType{
				EthType: uint32(val.(ofp.EthernetType)),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_IP_PROTO:
			field.Value = &voltha.OfpOxmOfbField_IpProto{
				IpProto: uint32(val.(ofp.IpPrototype)),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_UDP_SRC:
			field.Value = &voltha.OfpOxmOfbField_UdpSrc{
				UdpSrc: uint32(val.(uint16)),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_UDP_DST:
			field.Value = &voltha.OfpOxmOfbField_UdpDst{
				UdpDst: uint32(val.(uint16)),
			}
		case voltha.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID:
			field.Value = &voltha.OfpOxmOfbField_VlanVid{
				VlanVid: uint32(val.(uint16)),
			}
		}
		oxmList = append(oxmList, &ofpOxmField)
	}

	// Construct request
	flowUpdate := openflow_13.FlowTableUpdate{
		Id: ofc.DeviceID,
		FlowMod: &voltha.OfpFlowMod{
			Cookie:      flowDeleteStrict.Cookie,
			CookieMask:  flowDeleteStrict.CookieMask,
			TableId:     uint32(flowDeleteStrict.TableId),
			Command:     voltha.OfpFlowModCommand_OFPFC_DELETE_STRICT,
			IdleTimeout: uint32(flowDeleteStrict.IdleTimeout),
			HardTimeout: uint32(flowDeleteStrict.HardTimeout),
			Priority:    uint32(flowDeleteStrict.Priority),
			BufferId:    flowDeleteStrict.BufferId,
			OutPort:     uint32(flowDeleteStrict.OutPort),
			OutGroup:    uint32(flowDeleteStrict.OutGroup),
			Flags:       uint32(flowDeleteStrict.Flags),
			Match: &voltha.OfpMatch{
				Type:      voltha.OfpMatchType(flowDeleteStrict.Match.GetType()),
				OxmFields: oxmList,
			},
		},
	}

	if logger.V(log.DebugLevel) {
		flowUpdateJs, _ := json.Marshal(flowUpdate)
		logger.Debugf("FlowUpdate being sent to Voltha",
			log.Fields{
				"device-id":   ofc.DeviceID,
				"flow-update": flowUpdateJs})
	}
	if _, err := ofc.VolthaClient.UpdateLogicalDeviceFlowTable(context.Background(), &flowUpdate); err != nil {
		logger.Errorw("Error calling FlowUpdate ",
			log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
	}
}
