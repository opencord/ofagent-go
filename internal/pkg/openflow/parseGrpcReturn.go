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
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/opencord/goloxi"
	ofp "github.com/opencord/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-protos/v5/go/openflow_13"
	"github.com/opencord/voltha-protos/v5/go/voltha"
)

func parseOxm(ctx context.Context, ofbField *openflow_13.OfpOxmOfbField) (goloxi.IOxm, error) {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(ofbField)
		logger.Debugw(ctx, "parseOxm called",
			log.Fields{"ofbField": js})
	}

	switch ofbField.Type {
	case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IN_PORT:
		ofpInPort := ofp.NewOxmInPort()
		val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_Port)
		ofpInPort.Value = ofp.Port(val.Port)
		return ofpInPort, nil
	case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_ETH_TYPE:
		ofpEthType := ofp.NewOxmEthType()
		val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_EthType)
		ofpEthType.Value = ofp.EthernetType(val.EthType)
		return ofpEthType, nil
	case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IN_PHY_PORT:
		ofpInPhyPort := ofp.NewOxmInPhyPort()
		val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_PhysicalPort)
		ofpInPhyPort.Value = ofp.Port(val.PhysicalPort)
		return ofpInPhyPort, nil
	case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IP_PROTO:
		ofpIpProto := ofp.NewOxmIpProto()
		val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_IpProto)
		ofpIpProto.Value = ofp.IpPrototype(val.IpProto)
		return ofpIpProto, nil
	case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IPV4_DST:
		ofpIpv4Dst := ofp.NewOxmIpv4Dst()
		val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_Ipv4Dst)
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.BigEndian, val.Ipv4Dst)
		if err != nil {
			logger.Errorw(ctx, "error writing ipv4 address %v",
				log.Fields{"error": err})
		}
		ofpIpv4Dst.Value = buf.Bytes()
		return ofpIpv4Dst, nil
	case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_UDP_SRC:
		ofpUdpSrc := ofp.NewOxmUdpSrc()
		val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_UdpSrc)
		ofpUdpSrc.Value = uint16(val.UdpSrc)
		return ofpUdpSrc, nil
	case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_UDP_DST:
		ofpUdpDst := ofp.NewOxmUdpDst()
		val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_UdpDst)
		ofpUdpDst.Value = uint16(val.UdpDst)
		return ofpUdpDst, nil
	case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID:
		ofpVlanVid := ofp.NewOxmVlanVid()
		val := ofbField.GetValue()
		if val == nil {
			ofpVlanVid.Value = uint16(0)
			return ofpVlanVid, nil
		}
		vlanId := val.(*openflow_13.OfpOxmOfbField_VlanVid)
		if ofbField.HasMask {
			ofpVlanVidMasked := ofp.NewOxmVlanVidMasked()
			valMask := ofbField.GetMask()
			vlanMask := valMask.(*openflow_13.OfpOxmOfbField_VlanVidMask)
			if vlanId.VlanVid == 4096 && vlanMask.VlanVidMask == 4096 {
				ofpVlanVidMasked.Value = uint16(vlanId.VlanVid)
				ofpVlanVidMasked.ValueMask = uint16(vlanMask.VlanVidMask)
			} else {
				ofpVlanVidMasked.Value = uint16(vlanId.VlanVid) | 0x1000
				ofpVlanVidMasked.ValueMask = uint16(vlanMask.VlanVidMask)

			}
			return ofpVlanVidMasked, nil
		}
		ofpVlanVid.Value = uint16(vlanId.VlanVid) | 0x1000
		return ofpVlanVid, nil
	case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_METADATA:
		ofpMetadata := ofp.NewOxmMetadata()
		val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_TableMetadata)
		ofpMetadata.Value = val.TableMetadata
		return ofpMetadata, nil
	case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_PCP:
		ofpVlanPcp := ofp.NewOxmVlanPcp()
		val := ofbField.GetValue()
		vlanPcp := val.(*openflow_13.OfpOxmOfbField_VlanPcp)
		ofpVlanPcp.Value = uint8(vlanPcp.VlanPcp)
		return ofpVlanPcp, nil
	case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_ETH_DST:
		ofpEthDst := ofp.NewOxmEthDst()
		val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_EthDst)
		ofpEthDst.Value = val.EthDst
		return ofpEthDst, nil
	case openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_ETH_SRC:
		ofpEthSrc := ofp.NewOxmEthSrc()
		val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_EthSrc)
		ofpEthSrc.Value = val.EthSrc
		return ofpEthSrc, nil
	default:
		if logger.V(log.WarnLevel) {
			js, _ := json.Marshal(ofbField)
			logger.Warnw(ctx, "ParseOXM Unhandled OxmField",
				log.Fields{"OfbField": js})
		}
	}
	return nil, fmt.Errorf("can't-parse-oxm %+v", ofbField)
}

func parseInstructions(ctx context.Context, ofpInstruction *openflow_13.OfpInstruction) (ofp.IInstruction, error) {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(ofpInstruction)
		logger.Debugw(ctx, "parseInstructions called",
			log.Fields{"Instruction": js,
				"ofp-Instruction": ofpInstruction})
	}
	instType := ofpInstruction.Type
	data := ofpInstruction.GetData()
	switch instType {
	case ofp.OFPITWriteMetadata:
		instruction := ofp.NewInstructionWriteMetadata()
		metadata := data.(*openflow_13.OfpInstruction_WriteMetadata).WriteMetadata
		instruction.Metadata = uint64(metadata.Metadata)
		return instruction, nil
	case ofp.OFPITMeter:
		instruction := ofp.NewInstructionMeter()
		meter := data.(*openflow_13.OfpInstruction_Meter).Meter
		instruction.MeterId = meter.MeterId
		return instruction, nil
	case ofp.OFPITGotoTable:
		instruction := ofp.NewInstructionGotoTable()
		gotoTable := data.(*openflow_13.OfpInstruction_GotoTable).GotoTable
		instruction.TableId = uint8(gotoTable.TableId)
		return instruction, nil
	case ofp.OFPITApplyActions:
		instruction := ofp.NewInstructionApplyActions()

		var actions []goloxi.IAction
		for _, ofpAction := range ofpInstruction.GetActions().Actions {
			action, err := parseAction(ctx, ofpAction)
			if err == nil {
				actions = append(actions, action)
			} else {
				return nil, fmt.Errorf("can't-parse-action %v", err)
			}
		}
		instruction.Actions = actions
		if logger.V(log.DebugLevel) {
			js, _ := json.Marshal(instruction)
			logger.Debugw(ctx, "parseInstructions returning",
				log.Fields{
					"parsed-instruction": js})
		}
		return instruction, nil
	}
	//shouldn't have reached here :<
	return nil, fmt.Errorf("can't-parse-instruction %+v", ofpInstruction)
}

func parseAction(ctx context.Context, ofpAction *openflow_13.OfpAction) (goloxi.IAction, error) {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(ofpAction)
		logger.Debugw(ctx, "parseAction called",
			log.Fields{"action": js})
	}
	switch ofpAction.Type {
	case openflow_13.OfpActionType_OFPAT_OUTPUT:
		ofpOutputAction := ofpAction.GetOutput()
		outputAction := ofp.NewActionOutput()
		outputAction.Port = ofp.Port(ofpOutputAction.Port)
		outputAction.MaxLen = uint16(ofpOutputAction.MaxLen)
		return outputAction, nil
	case openflow_13.OfpActionType_OFPAT_PUSH_VLAN:
		ofpPushVlanAction := ofp.NewActionPushVlan()
		ofpPushVlanAction.Ethertype = uint16(ofpAction.GetPush().Ethertype)
		return ofpPushVlanAction, nil
	case openflow_13.OfpActionType_OFPAT_POP_VLAN:
		ofpPopVlanAction := ofp.NewActionPopVlan()
		return ofpPopVlanAction, nil
	case openflow_13.OfpActionType_OFPAT_SET_FIELD:
		ofpActionSetField := ofpAction.GetSetField()
		setFieldAction := ofp.NewActionSetField()

		iOxm, err := parseOxm(ctx, ofpActionSetField.GetField().GetOfbField())
		if err == nil {
			setFieldAction.Field = iOxm
		} else {
			return nil, fmt.Errorf("can't-parse-oxm %v", err)
		}
		return setFieldAction, nil
	case openflow_13.OfpActionType_OFPAT_GROUP:
		ofpGroupAction := ofpAction.GetGroup()
		groupAction := ofp.NewActionGroup()
		groupAction.GroupId = ofpGroupAction.GroupId
		return groupAction, nil
	default:
		if logger.V(log.WarnLevel) {
			js, _ := json.Marshal(ofpAction)
			logger.Warnw(ctx, "parseAction unknow action",
				log.Fields{"action": js})
		}
	}
	return nil, fmt.Errorf("can't-parse-action %+v", ofpAction)
}

func parsePortStats(port *voltha.LogicalPort) *ofp.PortStatsEntry {
	stats := port.OfpPortStats
	port.OfpPort.GetPortNo()
	var entry ofp.PortStatsEntry
	entry.SetPortNo(ofp.Port(port.OfpPort.GetPortNo()))
	entry.SetRxPackets(stats.GetRxPackets())
	entry.SetTxPackets(stats.GetTxPackets())
	entry.SetRxBytes(stats.GetRxBytes())
	entry.SetTxBytes(stats.GetTxBytes())
	entry.SetRxDropped(stats.GetRxDropped())
	entry.SetTxDropped(stats.GetTxDropped())
	entry.SetRxErrors(stats.GetRxErrors())
	entry.SetTxErrors(stats.GetTxErrors())
	entry.SetRxFrameErr(stats.GetRxFrameErr())
	entry.SetRxOverErr(stats.GetRxOverErr())
	entry.SetRxCrcErr(stats.GetRxCrcErr())
	entry.SetCollisions(stats.GetCollisions())
	entry.SetDurationSec(stats.GetDurationSec())
	entry.SetDurationNsec(stats.GetDurationNsec())
	return &entry
}
