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
	"log"

	ofp "github.com/donNewtonAlpha/goloxi/of13"
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

func handleFlowAdd(flowAdd *ofp.FlowAdd, deviceId string) {
	js, _ := json.Marshal(flowAdd)
	log.Printf("handleFlowAdd called with %s", js)

	var flowUpdate openflow_13.FlowTableUpdate
	flowUpdate.Id = deviceId
	var flowMod pb.OfpFlowMod
	flowMod.Cookie = flowAdd.Cookie
	flowMod.CookieMask = flowAdd.CookieMask
	flowMod.TableId = uint32(flowAdd.TableId)
	//flowMod.Command = pb.OfpFlowModCommand_OFPFC_ADD
	flowMod.Command = 0
	flowMod.IdleTimeout = uint32(flowAdd.IdleTimeout)
	flowMod.HardTimeout = uint32(flowAdd.HardTimeout)
	flowMod.Priority = uint32(flowAdd.Priority)
	flowMod.BufferId = flowAdd.BufferId
	flowMod.OutPort = uint32(flowAdd.OutPort)
	flowMod.OutGroup = uint32(flowAdd.OutGroup)
	flowMod.Flags = uint32(flowAdd.Flags)
	js, _ = json.Marshal(flowMod)
	log.Printf("HANDLE FLOW ADD FLOW_MOD %s", js)
	inMatch := flowAdd.Match
	var flowMatch pb.OfpMatch
	flowMatch.Type = pb.OfpMatchType(inMatch.GetType())
	var oxmList []*pb.OfpOxmField
	inOxmList := inMatch.GetOxmList()
	for i := 0; i < len(inOxmList); i++ {
		oxmField := inOxmList[i]
		j, _ := json.Marshal(oxmField)

		name := oxmMap[oxmField.GetOXMName()]
		log.Printf("\n\n\n %s  %d  %s\n\n\n", j, name, oxmField.GetOXMName())

		val := oxmField.GetOXMValue()
		var ofpOxmField pb.OfpOxmField
		ofpOxmField.OxmClass = ofp.OFPXMCOpenflowBasic
		var field pb.OfpOxmOfbField

		field.Type = pb.OxmOfbFieldTypes(name)
		log.Println("****\nFieldType: " + openflow_13.OxmOfbFieldTypes_name[name] + "\n\n\n\n")

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
			log.Printf("udpSrc %v", udpSrc)
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
	ofpInstructions := flowAdd.GetInstructions()
	for i := 0; i < len(ofpInstructions); i++ {
		var instruction pb.OfpInstruction
		ofpInstruction := ofpInstructions[i]
		instructionType := ofpInstruction.GetType()
		instruction.Type = uint32(instructionType)
		switch instructionType {
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
			for i := 0; i < len(actions); i++ {
				action := actions[i]
				ofpAction := extractAction(action)
				ofpActions = append(ofpActions, ofpAction)
			}
			instruction.Data = &ofpInstructionActions
		case ofp.OFPITApplyActions:
			applyAction := ofpInstruction.(ofp.IInstructionApplyActions)
			var ofpInstructionActions openflow_13.OfpInstruction_Actions
			var ofpActions []*openflow_13.OfpAction
			actions := applyAction.GetActions()
			for i := 0; i < len(actions); i++ {
				action := actions[i]
				ofpAction := extractAction(action)
				ofpActions = append(ofpActions, ofpAction)
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
	grpcClient := *getGrpcClient()
	flowUpdateJs, _ := json.Marshal(flowUpdate)
	log.Printf("FLOW UPDATE %s", flowUpdateJs)
	empty, err := grpcClient.UpdateLogicalDeviceFlowTable(context.Background(), &flowUpdate)
	if err != nil {
		log.Printf("ERROR DOING FLOW MOD ADD %v", err)
	}
	emptyJs, _ := json.Marshal(empty)
	log.Printf("FLOW MOD RESPONSE %s", emptyJs)

}

func handleFlowMod(flowMod *ofp.FlowMod, deviceId string) {
	js, _ := json.Marshal(flowMod)
	log.Printf("handleFlowMod called with %s", js)
}

func handleFlowModStrict(flowModStrict *ofp.FlowModifyStrict, deviceId string) {
	js, _ := json.Marshal(flowModStrict)
	log.Printf("handleFlowModStrict called with %s", js)
}
func handleFlowDelete(flowDelete *ofp.FlowDelete, deviceId string) {
	js, _ := json.Marshal(flowDelete)
	log.Printf("handleFlowDelete called with %s", js)

}
func handleFlowDeleteStrict(flowDeleteStrict *ofp.FlowDeleteStrict, deviceId string) {
	js, _ := json.Marshal(flowDeleteStrict)
	log.Printf("handleFlowDeleteStrict called with %s", js)

	var flowUpdate openflow_13.FlowTableUpdate
	flowUpdate.Id = deviceId
	var flowMod pb.OfpFlowMod
	flowMod.Cookie = flowDeleteStrict.Cookie
	flowMod.CookieMask = flowDeleteStrict.CookieMask
	flowMod.TableId = uint32(flowDeleteStrict.TableId)
	flowMod.Command = pb.OfpFlowModCommand_OFPFC_DELETE_STRICT
	flowMod.IdleTimeout = uint32(flowDeleteStrict.IdleTimeout)
	flowMod.HardTimeout = uint32(flowDeleteStrict.HardTimeout)
	flowMod.Priority = uint32(flowDeleteStrict.Priority)
	flowMod.BufferId = flowDeleteStrict.BufferId
	flowMod.OutPort = uint32(flowDeleteStrict.OutPort)
	flowMod.OutGroup = uint32(flowDeleteStrict.OutGroup)
	flowMod.Flags = uint32(flowDeleteStrict.Flags)
	inMatch := flowDeleteStrict.Match
	var flowMatch pb.OfpMatch
	flowMatch.Type = pb.OfpMatchType(inMatch.GetType())
	var oxmList []*pb.OfpOxmField
	inOxmList := inMatch.GetOxmList()
	for i := 0; i < len(inOxmList); i++ {
		oxmField := inOxmList[i]
		j, _ := json.Marshal(oxmField)

		name := oxmMap[oxmField.GetOXMName()]
		log.Printf("\n\n\n %s  %d  %s\n\n\n", j, name, oxmField.GetOXMName())

		val := oxmField.GetOXMValue()
		var ofpOxmField pb.OfpOxmField
		ofpOxmField.OxmClass = ofp.OFPXMCOpenflowBasic
		var field pb.OfpOxmOfbField

		field.Type = pb.OxmOfbFieldTypes(name)
		log.Println("****\nFieldType: " + openflow_13.OxmOfbFieldTypes_name[name] + "\n\n\n\n")

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
			log.Printf("udpSrc %v", udpSrc)
			field.Value = &value
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_UDP_DST:
			udpDst := val.(uint16)
			var value pb.OfpOxmOfbField_UdpDst
			value.UdpDst = uint32(udpDst)
			field.Value = &value
		case pb.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID:
			//vid := (val.(uint16) & 0xfff)|0x1000
			vid := val.(uint16)
			//vid := val.(uint16) & 0xfff
			var value pb.OfpOxmOfbField_VlanVid
			value.VlanVid = uint32(vid)
			field.Value = &value
		}
		oxmList = append(oxmList, &ofpOxmField)
	}
	flowMatch.OxmFields = oxmList
	flowMod.Match = &flowMatch
	/*var instructions []*pb.OfpInstruction
	ofpInstructions := flowDeleteStrict.GetInstructions()
	for i := 0; i < len(ofpInstructions); i++ {
		var instruction pb.OfpInstruction
		ofpInstruction := ofpInstructions[i]
		instructionType := ofpInstruction.GetType()
		instruction.Type = uint32(instructionType)
		switch instructionType {
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
			for i := 0; i < len(actions); i++ {
				action := actions[i]
				ofpAction := extractAction(action)
				ofpActions = append(ofpActions, ofpAction)
			}
			instruction.Data = &ofpInstructionActions
		case ofp.OFPITApplyActions:
			applyAction := ofpInstruction.(ofp.IInstructionApplyActions)
			var ofpInstructionActions openflow_13.OfpInstruction_Actions
			var ofpActions []*openflow_13.OfpAction
			actions := applyAction.GetActions()
			for i := 0; i < len(actions); i++ {
				action := actions[i]
				ofpAction := extractAction(action)
				ofpActions = append(ofpActions, ofpAction)
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

	*/
	flowUpdate.FlowMod = &flowMod
	grpcClient := *getGrpcClient()
	flowUpdateJs, _ := json.Marshal(flowUpdate)
	log.Printf("FLOW UPDATE %s", flowUpdateJs)
	empty, err := grpcClient.UpdateLogicalDeviceFlowTable(context.Background(), &flowUpdate)
	if err != nil {
		log.Printf("ERROR DOING FLOW MOD ADD %v", err)
	}
	emptyJs, _ := json.Marshal(empty)
	log.Printf("FLOW MOD RESPONSE %s", emptyJs)

}
