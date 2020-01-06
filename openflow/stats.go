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

	"net"

	"unsafe"

	"github.com/opencord/ofagent-go/settings"
	l "github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/openflow_13"

	"github.com/donNewtonAlpha/goloxi"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-protos/v2/go/common"
)

func handleStatsRequest(request ofp.IHeader, statType uint16, DeviceID string, client *Client) error {
	if settings.GetDebug(DeviceID) {
		js, _ := json.Marshal(request)
		logger.Debugw("handleStatsRequest called", l.Fields{"DeviceID": DeviceID, "StatType": statType, "rquest": js})
	}
	var id = common.ID{Id: DeviceID}
	switch statType {
	case ofp.OFPSTDesc:
		statsReq := request.(*ofp.DescStatsRequest)
		response, err := handleDescStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)
	case ofp.OFPSTFlow:
		statsReq := request.(*ofp.FlowStatsRequest)
		response, err := handleFlowStatsRequest(statsReq, id, DeviceID)
		if err != nil {
			return err
		}
		response.Length = uint16(unsafe.Sizeof(*response))
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)

	case ofp.OFPSTAggregate:
		statsReq := request.(*ofp.AggregateStatsRequest)
		response, err := handleAggregateStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)
	case ofp.OFPSTTable:
		statsReq := request.(*ofp.TableStatsRequest)
		response, e := handleTableStatsRequest(statsReq, id)
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		if e != nil {
			return e
		}
		return client.SendMessage(response)
	case ofp.OFPSTPort:
		statsReq := request.(*ofp.PortStatsRequest)
		response, err := handlePortStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)

	case ofp.OFPSTQueue:
		statsReq := request.(*ofp.QueueStatsRequest)
		response, err := handleQueueStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)
	case ofp.OFPSTGroup:
		statsReq := request.(*ofp.GroupStatsRequest)
		response, err := handleGroupStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		client.SendMessage(response)
	case ofp.OFPSTGroupDesc:
		statsReq := request.(*ofp.GroupDescStatsRequest)
		response, err := handleGroupStatsDescRequest(statsReq, id)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)

	case ofp.OFPSTGroupFeatures:
		statsReq := request.(*ofp.GroupFeaturesStatsRequest)
		response, err := handleGroupFeatureStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)
	case ofp.OFPSTMeter:
		statsReq := request.(*ofp.MeterStatsRequest)
		response, err := handleMeterStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)
	case ofp.OFPSTMeterConfig:
		statsReq := request.(*ofp.MeterConfigStatsRequest)
		response, err := handleMeterConfigStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)
	case ofp.OFPSTMeterFeatures:
		statsReq := request.(*ofp.MeterFeaturesStatsRequest)
		response, err := handleMeterFeatureStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)
	case ofp.OFPSTTableFeatures:
		statsReq := request.(*ofp.TableFeaturesStatsRequest)
		response, err := handleTableFeaturesStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)
	case ofp.OFPSTPortDesc:
		statsReq := request.(*ofp.PortDescStatsRequest)
		response, err := handlePortDescStatsRequest(statsReq, DeviceID)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)

	case ofp.OFPSTExperimenter:
		statsReq := request.(*ofp.ExperimenterStatsRequest)
		response, err := handleExperimenterStatsRequest(statsReq, id)
		if err != nil {
			return err
		}
		if settings.GetDebug(DeviceID) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("HandleStatsRequest GRPC", l.Fields{"DeviceID": DeviceID, "Req": reqJs, "Res": resJs})
		}
		return client.SendMessage(response)
	}
	return nil
}

func handleDescStatsRequest(request *ofp.DescStatsRequest, id common.ID) (*ofp.DescStatsReply, error) {
	response := ofp.NewDescStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(request.GetVersion())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))

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
	return response, nil
}
func handleFlowStatsRequest(request *ofp.FlowStatsRequest, id common.ID, DeviceID string) (*ofp.FlowStatsReply, error) {
	response := ofp.NewFlowStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(4)
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	client := *getGrpcClient()
	resp, err := client.ListLogicalDeviceFlows(context.Background(), &id)
	if err != nil {
		return nil, err
	}
	var flow []*ofp.FlowStatsEntry
	items := resp.GetItems()
	for i := 0; i < len(items); i++ {
		item := items[i]
		entry := ofp.NewFlowStatsEntry()
		entry.SetTableId(uint8(item.GetTableId()))
		entry.SetDurationSec(item.GetDurationSec())
		entry.SetDurationNsec(item.GetDurationNsec())
		entry.SetPriority(uint16(item.GetPriority()))
		entry.SetIdleTimeout(uint16(item.GetIdleTimeout()))
		entry.SetHardTimeout(uint16(item.GetHardTimeout()))
		entry.SetFlags(ofp.FlowModFlags(item.GetFlags()))
		entry.SetCookie(item.GetCookie())
		entry.SetPacketCount(item.GetPacketCount())
		entry.SetByteCount(item.GetByteCount())
		var entrySize uint16
		entrySize = 48
		match := ofp.NewMatchV3()
		pbMatch := item.GetMatch()
		var fields []goloxi.IOxm
		match.SetType(uint16(pbMatch.GetType()))
		oxFields := pbMatch.GetOxmFields()
		var size uint16
		size = 4
		for i := 0; i < len(oxFields); i++ {
			oxmField := oxFields[i]
			field := oxmField.GetField()
			ofbField := field.(*openflow_13.OfpOxmField_OfbField).OfbField
			iOxm, oxmSize := parseOxm(ofbField, DeviceID)
			fields = append(fields, iOxm)
			if oxmSize > 0 {
				size += 4 //header for oxm
			}
			size += oxmSize
		}

		match.OxmList = fields
		match.Length = uint16(size)
		//account for 8 byte alignment
		if size%8 != 0 {
			size = ((size / 8) + 1) * 8
		}
		entrySize += size
		entry.SetMatch(*match)
		var instructions []ofp.IInstruction
		ofpInstructions := item.Instructions
		for i := 0; i < len(ofpInstructions); i++ {
			instruction, size := parseInstructions(ofpInstructions[i], DeviceID)
			instructions = append(instructions, instruction)
			entrySize += size
		}
		entry.Instructions = instructions
		entry.Length = entrySize
		entrySize = 0
		flow = append(flow, entry)
	}
	response.SetEntries(flow)
	return response, nil
}
func handleAggregateStatsRequest(request *ofp.AggregateStatsRequest, id common.ID) (*ofp.AggregateStatsReply, error) {
	response := ofp.NewAggregateStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	response.SetFlowCount(0)
	//TODO wire this to voltha core when it implements
	return response, nil
}
func handleGroupStatsRequest(request *ofp.GroupStatsRequest, id common.ID) (*ofp.GroupStatsReply, error) {
	response := ofp.NewGroupStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
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
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
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
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}
func handleMeterStatsRequest(request *ofp.MeterStatsRequest, id common.ID) (*ofp.MeterStatsReply, error) {
	response := ofp.NewMeterStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	client := *getGrpcClient()
	resp, err := client.ListLogicalDeviceMeters(context.Background(), &id)
	if err != nil {
		return nil, err
	}
	size := uint16(40)
	items := resp.Items
	var meterStats []*ofp.MeterStats
	for i := 0; i < len(items); i++ {
		meterStat := ofp.NewMeterStats()
		item := items[i]
		stats := item.Stats
		meterStat.DurationNsec = stats.DurationNsec
		meterStat.DurationSec = stats.DurationSec
		meterStat.ByteInCount = stats.ByteInCount
		meterStat.FlowCount = stats.FlowCount
		meterStat.MeterId = stats.MeterId
		var bandStats []*ofp.MeterBandStats
		bStats := stats.BandStats
		for j := 0; j < len(bStats); j++ {
			bStat := bStats[j]
			bandStat := ofp.NewMeterBandStats()
			bandStat.ByteBandCount = bStat.ByteBandCount
			bandStat.PacketBandCount = bStat.PacketBandCount
			bandStats = append(bandStats, bandStat)
			size += 16
		}
		meterStat.SetBandStats(bandStats)
		meterStat.Len = size
		meterStats = append(meterStats, meterStat)
	}
	response.SetEntries(meterStats)
	return response, nil
}
func handleMeterConfigStatsRequest(request *ofp.MeterConfigStatsRequest, id common.ID) (*ofp.MeterConfigStatsReply, error) {
	response := ofp.NewMeterConfigStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}
func handleTableFeaturesStatsRequest(request *ofp.TableFeaturesStatsRequest, id common.ID) (*ofp.TableFeaturesStatsReply, error) {
	response := ofp.NewTableFeaturesStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}
func handleTableStatsRequest(request *ofp.TableStatsRequest, id common.ID) (*ofp.TableStatsReply, error) {
	var response = ofp.NewTableStatsReply()
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	return response, nil
}
func handleQueueStatsRequest(request *ofp.QueueStatsRequest, id common.ID) (*ofp.QueueStatsReply, error) {
	response := ofp.NewQueueStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}
func handlePortStatsRequest(request *ofp.PortStatsRequest, id common.ID) (*ofp.PortStatsReply, error) {
	response := ofp.NewPortStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(request.GetVersion())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	client := *getGrpcClient()
	reply, err := client.ListLogicalDevicePorts(context.Background(), &id)
	//reply,err := client.GetLogicalDevicePort(context.Background(),&id)
	if err != nil {
		return nil, err
	}
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
	return response, nil
}

func handlePortDescStatsRequest(request *ofp.PortDescStatsRequest, DeviceID string) (*ofp.PortDescStatsReply, error) {
	response := ofp.NewPortDescStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	var grpcClient = *getGrpcClient()
	var id = common.ID{Id: DeviceID}
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
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	meterFeatures := ofp.NewMeterFeatures()
	meterFeatures.Capabilities = ofp.OFPMFKbps
	meterFeatures.BandTypes = ofp.OFPMBTDrop
	meterFeatures.MaxMeter = 0xffffffff
	meterFeatures.MaxBands = 0xff
	meterFeatures.MaxColor = 0xff
	response.Features = *meterFeatures
	return response, nil
}
func handleExperimenterStatsRequest(request *ofp.ExperimenterStatsRequest, id common.ID) (*ofp.ExperimenterStatsReply, error) {
	response := ofp.NewExperimenterStatsReply(request.GetExperimenter())
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}
