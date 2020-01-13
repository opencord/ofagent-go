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
	"context"
	"encoding/json"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/openflow_13"
	pb "github.com/opencord/voltha-protos/v2/go/voltha"
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
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		js, _ := json.Marshal(flowAdd)
		log.Debugw("handleFlowAdd called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"params":    js})
	}

	flowUpdate := openflow_13.FlowTableUpdate{Id: ofc.DeviceID}
	flowMod := pb.OfpFlowMod{
		Cookie:      flowAdd.Cookie,
		CookieMask:  flowAdd.CookieMask,
		TableId:     uint32(flowAdd.TableId),
		Command:     pb.OfpFlowModCommand_OFPFC_ADD,
		IdleTimeout: uint32(flowAdd.IdleTimeout),
		HardTimeout: uint32(flowAdd.HardTimeout),
		Priority:    uint32(flowAdd.Priority),
		BufferId:    flowAdd.BufferId,
		OutPort:     uint32(flowAdd.OutPort),
		OutGroup:    uint32(flowAdd.OutGroup),
		Flags:       uint32(flowAdd.Flags),
	}
	flowMatch := pb.OfpMatch{
		Type: pb.OfpMatchType(flowAdd.Match.GetType()),
	}
	var oxmList []*pb.OfpOxmField
	for _, oxmField := range flowAdd.Match.GetOxmList() {
		name := oxmMap[oxmField.GetOXMName()]
		val := oxmField.GetOXMValue()
		var ofpOxmField pb.OfpOxmField
		ofpOxmField.OxmClass = ofp.OFPXMCOpenflowBasic
		var field pb.OfpOxmOfbField
		field.Type = pb.OxmOfbFieldTypes(name)
		var x openflow_13.OfpOxmField_OfbField
		x.OfbField = &field
		ofpOxmField.Field = &x
		switch pb.OxmOfbFieldTypes(name) {
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_IN_PORT:
			port := val.(ofp.Port)
			var value pb.OfpOxmOfbField_Port
			value.Port = uint32(port)
			field.Value = &value
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_IN_PHY_PORT:
			phyPort := val.(uint32)
			var value pb.OfpOxmOfbField_PhysicalPort
			value.PhysicalPort = phyPort
			field.Value = &value
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_METADATA:
			metadata := val.(uint64)
			var value pb.OfpOxmOfbField_TableMetadata
			value.TableMetadata = metadata
			field.Value = &value
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_ETH_TYPE:
			ethType := val.(ofp.EthernetType)
			var value pb.OfpOxmOfbField_EthType
			value.EthType = uint32(ethType)
			field.Value = &value
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_IP_PROTO:
			proto := val.(ofp.IpPrototype)
			var value pb.OfpOxmOfbField_IpProto
			value.IpProto = uint32(proto)
			field.Value = &value
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_UDP_SRC:
			udpSrc := val.(uint16)
			var value pb.OfpOxmOfbField_UdpSrc
			value.UdpSrc = uint32(udpSrc)
			field.Value = &value
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_UDP_DST:
			udpDst := val.(uint16)
			var value pb.OfpOxmOfbField_UdpDst
			value.UdpDst = uint32(udpDst)
			field.Value = &value
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID:
			vid := (val.(uint16) & 0xfff) | 0x1000
			var value pb.OfpOxmOfbField_VlanVid
			value.VlanVid = uint32(vid)
			field.Value = &value
		}
		oxmList = append(oxmList, &ofpOxmField)
	}
	flowMatch.OxmFields = oxmList
	flowMod.Match = &flowMatch
	var instructions []*pb.OfpInstruction
	for _, ofpInstruction := range flowAdd.GetInstructions() {
		var instruction pb.OfpInstruction
		switch uint32(ofpInstruction.GetType()) {
		case ofp.OFPITGotoTable:
			goToTable := ofpInstruction.(ofp.IInstructionGotoTable)
			var ofpGoToTable openflow_13.OfpInstruction_GotoTable
			var oGoToTable openflow_13.OfpInstructionGotoTable
			ofpGoToTable.GotoTable = &oGoToTable
			ofpGoToTable.GotoTable.TableId = uint32(goToTable.GetTableId())
			instruction.Data = &ofpGoToTable
		case ofp.OFPITWriteMetadata:
			writeMetaData := ofpInstruction.(ofp.IInstructionWriteMetadata)
			var ofpWriteMetadata openflow_13.OfpInstruction_WriteMetadata
			var writeMetadata openflow_13.OfpInstructionWriteMetadata
			ofpWriteMetadata.WriteMetadata = &writeMetadata
			ofpWriteMetadata.WriteMetadata.Metadata = writeMetaData.GetMetadata()
			ofpWriteMetadata.WriteMetadata.MetadataMask = writeMetaData.GetMetadataMask()
			instruction.Data = &ofpWriteMetadata
		case ofp.OFPITWriteActions:
			writeAction := ofpInstruction.(ofp.IInstructionWriteActions)
			var ofpInstructionActions openflow_13.OfpInstruction_Actions
			var ofpActions []*openflow_13.OfpAction
			actions := writeAction.GetActions()
			for _, action := range actions {
				ofpActions = append(ofpActions, extractAction(action))
			}
			instruction.Data = &ofpInstructionActions
		case ofp.OFPITApplyActions:
			applyAction := ofpInstruction.(ofp.IInstructionApplyActions)
			var ofpInstructionActions openflow_13.OfpInstruction_Actions
			var ofpActions []*openflow_13.OfpAction
			actions := applyAction.GetActions()
			for _, action := range actions {
				ofpActions = append(ofpActions, extractAction(action))
			}
			var actionsHolder openflow_13.OfpInstructionActions
			actionsHolder.Actions = ofpActions
			ofpInstructionActions.Actions = &actionsHolder
			instruction.Data = &ofpInstructionActions
		case ofp.OFPITMeter:
			var instructionMeter = ofpInstruction.(ofp.IInstructionMeter)
			var meterInstruction openflow_13.OfpInstruction_Meter
			var meter openflow_13.OfpInstructionMeter

			meter.MeterId = instructionMeter.GetMeterId()
			meterInstruction.Meter = &meter
			instruction.Data = &meterInstruction
		}
		instructions = append(instructions, &instruction)
	}

	flowMod.Instructions = instructions
	flowUpdate.FlowMod = &flowMod
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		flowUpdateJs, _ := json.Marshal(flowUpdate)
		log.Debugf("FlowUpdate being sent to Voltha",
			log.Fields{
				"device-id":      ofc.DeviceID,
				"FlowModRequest": flowUpdateJs})
	}
	if _, err := ofc.volthaClient.UpdateLogicalDeviceFlowTable(context.Background(), &flowUpdate); err != nil {
		// TODO: More than log error?
		log.Errorw("Error calling FlowUpdate ",
			log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
	}
}

func (ofc *OFClient) handleFlowMod(flowMod *ofp.FlowMod) {
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		js, _ := json.Marshal(flowMod)
		log.Debugw("handleMod called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"params":    js})
	}
	// TODO: Need to return error?
	log.Error("handleFlowMod not implemented")
}

func (ofc *OFClient) handleFlowModStrict(flowModStrict *ofp.FlowModifyStrict) {
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		js, _ := json.Marshal(flowModStrict)
		log.Debugw("handleFlowModStrict called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"params":    js})
	}
	// TODO: Need to return error?
	log.Error("handleFlowModStrict not implemented")
}
func (ofc *OFClient) handleFlowDelete(flowDelete *ofp.FlowDelete) {
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		js, _ := json.Marshal(flowDelete)
		log.Debugw("handleFlowDelete called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"params":    js})
	}
	// TODO: Need to return error?
	log.Error("handleFlowDelete not implemented")

}
func (ofc *OFClient) handleFlowDeleteStrict(flowDeleteStrict *ofp.FlowDeleteStrict) {
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		js, _ := json.Marshal(flowDeleteStrict)
		log.Debugw("handleFlowAdd called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"params":    js})
	}

	flowUpdate := openflow_13.FlowTableUpdate{Id: ofc.DeviceID}
	flowMod := pb.OfpFlowMod{
		Cookie:      flowDeleteStrict.Cookie,
		CookieMask:  flowDeleteStrict.CookieMask,
		TableId:     uint32(flowDeleteStrict.TableId),
		Command:     pb.OfpFlowModCommand_OFPFC_DELETE_STRICT,
		IdleTimeout: uint32(flowDeleteStrict.IdleTimeout),
		HardTimeout: uint32(flowDeleteStrict.HardTimeout),
		Priority:    uint32(flowDeleteStrict.Priority),
		BufferId:    flowDeleteStrict.BufferId,
		OutPort:     uint32(flowDeleteStrict.OutPort),
		OutGroup:    uint32(flowDeleteStrict.OutGroup),
		Flags:       uint32(flowDeleteStrict.Flags),
	}
	var flowMatch pb.OfpMatch
	flowMatch.Type = pb.OfpMatchType(flowDeleteStrict.Match.GetType())
	var oxmList []*pb.OfpOxmField
	for _, oxmField := range flowDeleteStrict.Match.GetOxmList() {
		name := oxmMap[oxmField.GetOXMName()]
		val := oxmField.GetOXMValue()
		field := pb.OfpOxmOfbField{Type: pb.OxmOfbFieldTypes(name)}
		ofpOxmField := pb.OfpOxmField{
			OxmClass: ofp.OFPXMCOpenflowBasic,
			Field:    &openflow_13.OfpOxmField_OfbField{OfbField: &field},
		}

		switch pb.OxmOfbFieldTypes(name) {
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_IN_PORT:
			field.Value = &pb.OfpOxmOfbField_Port{
				Port: uint32(val.(ofp.Port)),
			}
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_IN_PHY_PORT:
			field.Value = &pb.OfpOxmOfbField_PhysicalPort{
				PhysicalPort: val.(uint32),
			}
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_METADATA:
			field.Value = &pb.OfpOxmOfbField_TableMetadata{
				TableMetadata: val.(uint64),
			}
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_ETH_TYPE:
			field.Value = &pb.OfpOxmOfbField_EthType{
				EthType: uint32(val.(ofp.EthernetType)),
			}
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_IP_PROTO:
			field.Value = &pb.OfpOxmOfbField_IpProto{
				IpProto: uint32(val.(ofp.IpPrototype)),
			}
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_UDP_SRC:
			field.Value = &pb.OfpOxmOfbField_UdpSrc{
				UdpSrc: uint32(val.(uint16)),
			}
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_UDP_DST:
			field.Value = &pb.OfpOxmOfbField_UdpDst{
				UdpDst: uint32(val.(uint16)),
			}
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID:
			field.Value = &pb.OfpOxmOfbField_VlanVid{
				VlanVid: uint32(val.(uint16)),
			}
		}
		oxmList = append(oxmList, &ofpOxmField)
	}
	flowMatch.OxmFields = oxmList
	flowMod.Match = &flowMatch
	flowUpdate.FlowMod = &flowMod
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		flowUpdateJs, _ := json.Marshal(flowUpdate)
		log.Debugf("FlowUpdate being sent to Voltha",
			log.Fields{
				"device-id":      ofc.DeviceID,
				"FlowModRequest": flowUpdateJs})
	}
	if _, err := ofc.volthaClient.UpdateLogicalDeviceFlowTable(context.Background(), &flowUpdate); err != nil {
		log.Errorw("Error calling FlowUpdate ",
			log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
	}
}
