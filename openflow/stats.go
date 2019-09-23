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
	"github.com/opencord/voltha-protos/go/openflow_13"
	"github.com/skydive-project/goloxi"
	"log"
	"net"
	"unsafe"

	"github.com/opencord/voltha-protos/go/common"
	pb "github.com/opencord/voltha-protos/go/voltha"
	ofp "github.com/skydive-project/goloxi/of13"
)

func handleStatsRequest(request ofp.IHeader, statType uint16, deviceId string, client *Client) error {
	message, _ := json.Marshal(request)
	log.Printf("handleStatsRequest called with %s\n ", message)
	var id = common.ID{Id: deviceId}

	switch statType {
	case 0:
		statsReq := request.(*ofp.DescStatsRequest)
		response, err := handleDescStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		client.SendMessage(response)

	case 1:
		statsReq := request.(*ofp.FlowStatsRequest)
		response, _ := handleFlowStatsRequest(statsReq, id)
		err := client.SendMessage(response)
		if err != nil {
			return err
		}

	case 2:
		statsReq := request.(*ofp.AggregateStatsRequest)
		aggregateStatsReply, err := handleAggregateStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		client.SendMessage(aggregateStatsReply)
	case 3:
		statsReq := request.(*ofp.TableStatsRequest)
		tableStatsReply, e := handleTableStatsRequest(statsReq, id)
		if e != nil {
			return e
		}
		client.SendMessage(tableStatsReply)
	case 4:
		statsReq := request.(*ofp.PortStatsRequest)
		response, err := handlePortStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		client.SendMessage(response)

	case 5:
		statsReq := request.(*ofp.QueueStatsRequest)
		response, err := handleQueueStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		client.SendMessage(response)
	case 6:
		statsReq := request.(*ofp.GroupStatsRequest)
		response, err := handleGroupStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		client.SendMessage(response)
	case 7:
		statsReq := request.(*ofp.GroupDescStatsRequest)
		response, err := handleGroupStatsDescRequest(statsReq, id)
		if err != nil {
			return err
		}
		client.SendMessage(response)
	case 8:
		statsReq := request.(*ofp.GroupFeaturesStatsRequest)
		response, err := handleGroupFeatureStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		client.SendMessage(response)
	case 9:
		statsReq := request.(*ofp.MeterStatsRequest)
		response, err := handleMeterStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		client.SendMessage(response)
	case 10:
		statsReq := request.(*ofp.MeterConfigStatsRequest)
		response, err := handleMeterConfigStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		client.SendMessage(response)
	case 11:
		statsReq := request.(*ofp.MeterFeaturesStatsRequest)
		response, err := handleMeterFeatureStatsRequest(statsReq)
		if err != nil {
			return err
		}
		client.SendMessage(response)
	case 12:
		statsReq := request.(*ofp.TableFeaturesStatsRequest)
		response, err := handleTableFeaturesStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		client.SendMessage(response)
	case 13:
		statsReq := request.(*ofp.PortDescStatsRequest)
		response, err := handlePortDescStatsRequest(statsReq, deviceId)
		if err != nil {
			return err
		}
		client.SendMessage(response)

	case 65535:
		statsReq := request.(*ofp.ExperimenterStatsRequest)
		response, err := handleExperimenterStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		client.SendMessage(response)
	}
	return nil
}

func handleDescStatsRequest(request *ofp.DescStatsRequest, id common.ID) (*ofp.DescStatsReply, error) {
	response := ofp.NewDescStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(request.GetVersion())
	client := *getGrpcClient()
	resp, err := client.GetLogicalDevice(context.Background(), &id)
	if err != nil {
		return nil, err
	}
	desc := resp.GetDesc()

	response.SetMfrDesc(PadString(desc.GetMfrDesc(), 256))
	response.SetHwDesc(PadString(desc.GetHwDesc(), 256))
	response.SetSwDesc(PadString(desc.GetSwDesc(), 256))
	response.SetSerialNum(PadString(desc.GetSerialNum(), 32))
	response.SetDpDesc(PadString(desc.GetDpDesc(), 256))
	//jsonRes,_ := json.Marshal(response)
	//log.Printf("handleDescStatsRequest response : %s",jsonRes)
	return response, nil
}
func handleFlowStatsRequest(request *ofp.FlowStatsRequest, id common.ID) (*ofp.FlowStatsReply, error) {
	log.Println("****************************************\n***********************************")
	response := ofp.NewFlowStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(request.GetVersion())
	client := *getGrpcClient()
	resp, err := client.ListLogicalDeviceFlows(context.Background(), &id)
	if err != nil {
		log.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		log.Printf("err in handleFlowStatsRequest calling ListLogicalDeviceFlows %v", err)
		return nil, err
	}
	log.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	js, _ := json.Marshal(resp.GetItems())
	log.Printf("||||||||||||||||||||||||||||||||||||HandFlowStat %s|||||||||||||||||||||||||||||||||||||||||", js)
	var flow []*ofp.FlowStatsEntry
	items := resp.GetItems()

	for i := 0; i < len(items); i++ {
		item := items[1]
		var entry ofp.FlowStatsEntry

		entry.SetTableId(uint8(item.GetId()))
		entry.SetDurationSec(item.GetDurationSec())
		entry.SetDurationNsec(item.GetDurationNsec())
		entry.SetPriority(uint16(item.GetPriority()))
		entry.SetIdleTimeout(uint16(item.GetIdleTimeout()))
		entry.SetHardTimeout(uint16(item.GetHardTimeout()))
		entry.SetFlags(ofp.FlowModFlags(item.GetFlags()))
		entry.SetCookie(item.GetCookie())
		entry.SetPacketCount(item.GetPacketCount())
		entry.SetByteCount(item.GetByteCount())
		var match ofp.Match
		pbMatch := item.GetMatch()

		var fields []goloxi.IOxm
		match.SetType(uint16(pbMatch.GetType()))
		oxFields := pbMatch.GetOxmFields()
		for i := 0; i < len(oxFields); i++ {
			js, _ := json.Marshal(oxFields[i])
			log.Printf("oxfields %s", js)
			oxmField := oxFields[i]
			field := oxmField.GetField()
			ofbField := field.(*openflow_13.OfpOxmField_OfbField).OfbField
			switch ofbField.Type {
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_IN_PORT:
				ofpInPort := ofp.NewOxmInPort()
				val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_Port)
				ofpInPort.Value = ofp.Port(val.Port)
				fields = append(fields, ofpInPort)
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_IN_PHY_PORT:
				ofpInPhyPort := ofp.NewOxmInPhyPort()
				val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_PhysicalPort)
				ofpInPhyPort.Value = ofp.Port(val.PhysicalPort)
				fields = append(fields, ofpInPhyPort)
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_IP_PROTO:
				ofpIpProto := ofp.NewOxmIpProto()
				val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_IpProto)
				ofpIpProto.Value = ofp.IpPrototype(val.IpProto)
				fields = append(fields, ofpIpProto)
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_UDP_SRC:
				ofpUdpSrc := ofp.NewOxmUdpSrc()
				val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_UdpSrc)
				ofpUdpSrc.Value = uint16(val.UdpSrc)
				fields = append(fields, ofpUdpSrc)
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_UDP_DST:
				ofpUdpDst := ofp.NewOxmUdpSrc()
				val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_UdpDst)
				ofpUdpDst.Value = uint16(val.UdpDst)
				fields = append(fields, ofpUdpDst)
			case pb.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID:
				ofpVlanVid := ofp.NewOxmVlanVid()
				val := ofbField.GetValue().(*openflow_13.OfpOxmOfbField_VlanVid)
				ofpVlanVid.Value = uint16(val.VlanVid)
				fields = append(fields, ofpVlanVid)
			default:
				log.Printf("handleFlowStatsRequest   Unhandled OxmField %v", ofbField.Type)
			}

		}
		match.OxmList = fields
		match.Length = uint16(unsafe.Sizeof(match))
		entry.SetMatch(match)
		var instructions []ofp.IInstruction
		ofpInstructions := item.Instructions
		for i := 0; i < len(ofpInstructions); i++ {
			ofpInstruction := ofpInstructions[i]
			instType := ofpInstruction.Type
			switch instType {
			case uint32(openflow_13.OfpInstructionType_OFPIT_APPLY_ACTIONS):
				ofpActions := ofpInstruction.GetActions().Actions
				for i := 0; i < len(ofpActions); i++ {
					ofpAction := ofpActions[i]
					var actions []*ofp.Action
					switch ofpAction.Type {
					case openflow_13.OfpActionType_OFPAT_OUTPUT:
						ofpOutputAction := ofpAction.GetOutput()
						outputAction := ofp.NewActionOutput()
						outputAction.Port = ofp.Port(ofpOutputAction.Port)
						outputAction.MaxLen = uint16(ofpOutputAction.MaxLen)
						actions = append(actions, outputAction.Action)
					}
				}
				instruction := ofp.NewInstruction(uint16(instType))
				instructions = append(instructions, instruction)
			}

		}
		entry.Instructions = instructions
		entry.Length = uint16(unsafe.Sizeof(entry))
		flow = append(flow, &entry)

	}
	response.SetEntries(flow)
	return response, nil
}
func handleAggregateStatsRequest(request *ofp.AggregateStatsRequest, id common.ID) (*ofp.AggregateStatsReply, error) {
	response := ofp.NewAggregateStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlowCount(0)
	//TODO wire this to voltha core when it implements
	return response, nil
}
func handleGroupStatsRequest(request *ofp.GroupStatsRequest, id common.ID) (*ofp.GroupStatsReply, error) {
	response := ofp.NewGroupStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	client := *getGrpcClient()
	reply, err := client.ListLogicalDeviceFlowGroups(context.Background(), &id)
	if err != nil {
		return nil, err
	}

	var groupStatsEntries []*ofp.GroupStatsEntry
	items := reply.GetItems()
	for i := 0; i < len(items); i++ {
		item := items[i].GetStats()
		var entry ofp.GroupStatsEntry
		entry.SetByteCount(item.GetByteCount())
		entry.SetPacketCount(item.GetPacketCount())
		entry.SetDurationNsec(item.GetDurationNsec())
		entry.SetDurationSec(item.GetDurationSec())
		entry.SetRefCount(item.GetRefCount())
		entry.SetGroupId(item.GetGroupId())
		bucketStats := item.GetBucketStats()
		var bucketStatsList []*ofp.BucketCounter
		for j := 0; j < len(bucketStats); j++ {
			bucketStat := bucketStats[i]
			var bucketCounter ofp.BucketCounter
			bucketCounter.SetPacketCount(bucketStat.GetPacketCount())
			bucketCounter.SetByteCount(bucketStat.GetByteCount())
			bucketStatsList = append(bucketStatsList, &bucketCounter)
		}
		entry.SetBucketStats(bucketStatsList)
		groupStatsEntries = append(groupStatsEntries, &entry)
	}
	response.SetEntries(groupStatsEntries)
	return response, nil
}
func handleGroupStatsDescRequest(request *ofp.GroupDescStatsRequest, id common.ID) (*ofp.GroupDescStatsReply, error) {
	response := ofp.NewGroupDescStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	client := *getGrpcClient()
	reply, err := client.ListLogicalDeviceFlowGroups(context.Background(), &id)
	if err != nil {
		return nil, err
	}
	entries := reply.GetItems()
	var groupDescStatsEntries []*ofp.GroupDescStatsEntry
	for i := 0; i < len(entries); i++ {
		item := entries[i].GetStats()
		var groupDesc ofp.GroupDescStatsEntry
		groupDesc.SetGroupId(item.GetGroupId())
		/*
			buckets := item.g
			var bucketList []*ofp.Bucket
			for j:=0;j<len(buckets);j++{

			}

			groupDesc.SetBuckets(bucketList)
		*/
		groupDescStatsEntries = append(groupDescStatsEntries, &groupDesc)
	}
	response.SetEntries(groupDescStatsEntries)
	return response, nil
}
func handleGroupFeatureStatsRequest(request *ofp.GroupFeaturesStatsRequest, id common.ID) (*ofp.GroupFeaturesStatsReply, error) {
	response := ofp.NewGroupFeaturesStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	//TODO wire this to voltha core when it implements
	return response, nil
}
func handleMeterStatsRequest(request *ofp.MeterStatsRequest, id common.ID) (*ofp.MeterStatsReply, error) {
	response := ofp.NewMeterStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	//TODO wire this to voltha core when it implements
	return response, nil
}

//statsReq := request.(*ofp.MeterConfigStatsRequest)
func handleMeterConfigStatsRequest(request *ofp.MeterConfigStatsRequest, id common.ID) (*ofp.MeterConfigStatsReply, error) {
	response := ofp.NewMeterConfigStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	//TODO wire this to voltha core when it implements
	return response, nil
}

//statsReq := request.(*ofp.TableFeaturesStatsRequest)
func handleTableFeaturesStatsRequest(request *ofp.TableFeaturesStatsRequest, id common.ID) (*ofp.TableFeaturesStatsReply, error) {
	response := ofp.NewTableFeaturesStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	//TODO wire this to voltha core when it implements
	return response, nil
}
func handleTableStatsRequest(request *ofp.TableStatsRequest, id common.ID) (*ofp.TableStatsReply, error) {
	var tableStatsReply = ofp.NewTableStatsReply()
	tableStatsReply.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	tableStatsReply.SetVersion(request.GetVersion())
	tableStatsReply.SetXid(request.GetXid())
	return tableStatsReply, nil
}
func handleQueueStatsRequest(request *ofp.QueueStatsRequest, id common.ID) (*ofp.QueueStatsReply, error) {
	response := ofp.NewQueueStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	//TODO wire this to voltha core when it implements
	return response, nil
}
func handlePortStatsRequest(request *ofp.PortStatsRequest, id common.ID) (*ofp.PortStatsReply, error) {
	log.Println("HERE")
	response := ofp.NewPortStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(request.GetVersion())
	//response.SetFlags(ofp.flagsrequest.GetFlags())
	client := *getGrpcClient()
	reply, err := client.ListLogicalDevicePorts(context.Background(), &id)
	//reply,err := client.GetLogicalDevicePort(context.Background(),&id)
	if err != nil {
		log.Printf("error calling ListDevicePorts %v", err)
		return nil, err
	}
	js, _ := json.Marshal(reply)
	log.Printf("PORT STATS REPLY %s", js)
	ports := reply.GetItems()
	var entries []*ofp.PortStatsEntry
	if request.GetPortNo() == 0xffffffff { //all ports
		for i := 0; i < len(ports); i++ {
			port := ports[i]
			entry := parsePortStats(port)
			entries = append(entries, &entry)
		}
	} else { //find right port that is requested
		for i := 0; i < len(ports); i++ {
			if ports[i].GetOfpPortStats().GetPortNo() == uint32(request.GetPortNo()) {
				entry := parsePortStats(ports[i])
				entries = append(entries, &entry)
			}
		}
	}
	response.SetEntries(entries)
	js, _ = json.Marshal(response)
	log.Printf("handlePortStatsResponse %s", js)
	return response, nil

}
func parsePortStats(port *pb.LogicalPort) ofp.PortStatsEntry {
	stats := port.OfpPortStats
	var entry ofp.PortStatsEntry
	entry.SetPortNo(ofp.Port(stats.GetPortNo()))
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
	return entry
}
func handlePortDescStatsRequest(request *ofp.PortDescStatsRequest, deviceId string) (*ofp.PortDescStatsReply, error) {
	response := ofp.NewPortDescStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	var grpcClient = *getGrpcClient()
	var id = common.ID{Id: deviceId}
	logicalDevice, err := grpcClient.GetLogicalDevice(context.Background(), &id)
	if err != nil {
		return nil, err
	}
	ports := logicalDevice.GetPorts()
	var entries []*ofp.PortDesc
	for i := 0; i < len(ports); i++ {
		var entry ofp.PortDesc
		port := ports[i].GetOfpPort()
		entry.SetPortNo(ofp.Port(port.GetPortNo()))

		intArray := port.GetHwAddr()
		var octets []byte
		for i := 0; i < len(intArray); i++ {
			octets = append(octets, byte(intArray[i]))
		}
		hwAddr := net.HardwareAddr(octets)
		entry.SetHwAddr(hwAddr)
		entry.SetName(PadString(port.GetName(), 16))
		entry.SetConfig(ofp.PortConfig(port.GetConfig()))
		entry.SetState(ofp.PortState(port.GetState()))
		entry.SetCurr(ofp.PortFeatures(port.GetCurr()))
		entry.SetAdvertised(ofp.PortFeatures(port.GetAdvertised()))
		entry.SetSupported(ofp.PortFeatures(port.GetSupported()))
		entry.SetPeer(ofp.PortFeatures(port.GetPeer()))
		entry.SetCurrSpeed(port.GetCurrSpeed())
		entry.SetMaxSpeed(port.GetMaxSpeed())

		entries = append(entries, &entry)
	}

	response.SetEntries(entries)
	//TODO call voltha and get port descriptions etc
	return response, nil

}
func handleMeterFeatureStatsRequest(request *ofp.MeterFeaturesStatsRequest) (*ofp.MeterFeaturesStatsReply, error) {
	response := ofp.NewMeterFeaturesStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(request.GetVersion())
	return response, nil
}
func handleExperimenterStatsRequest(request *ofp.ExperimenterStatsRequest, id common.ID) (*ofp.ExperimenterStatsReply, error) {
	response := ofp.NewExperimenterStatsReply(request.GetExperimenter())
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	//TODO wire this to voltha core when it implements
	return response, nil
}
