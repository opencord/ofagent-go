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
	"errors"
	"github.com/donNewtonAlpha/goloxi"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/common"
	"github.com/opencord/voltha-protos/v2/go/openflow_13"
	"net"
	"unsafe"
)

func (ofc *OFClient) handleStatsRequest(request ofp.IHeader, statType uint16) error {
	if log.GetDefaultLogLevel() <= log.DebugLevel {
		js, _ := json.Marshal(request)
		log.Debugw("handleStatsRequest called",
			log.Fields{
				"device-id":  ofc.DeviceID,
				"stats-type": statType,
				"request":    js})
	}

	switch statType {
	case ofp.OFPSTDesc:
		response, err := ofc.handleDescStatsRequest(request.(ofp.IDescStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IDescStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTFlow:
		response, err := ofc.handleFlowStatsRequest(request.(ofp.IFlowStatsRequest))
		if err != nil {
			return err
		}
		response.Length = uint16(unsafe.Sizeof(*response))
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IFlowStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTAggregate:
		response, err := ofc.handleAggregateStatsRequest(request.(ofp.IAggregateStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IAggregateStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"DeviceID": ofc.DeviceID,
					"request":  reqJs,
					"response": resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTTable:
		response, err := ofc.handleTableStatsRequest(request.(ofp.ITableStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.ITableStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTPort:
		response, err := ofc.handlePortStatsRequest(request.(ofp.IPortStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IPortStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)

	case ofp.OFPSTQueue:
		response, err := ofc.handleQueueStatsRequest(request.(ofp.IQueueStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IQueueStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTGroup:
		response, err := ofc.handleGroupStatsRequest(request.(ofp.IGroupStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IGroupStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		ofc.SendMessage(response)
	case ofp.OFPSTGroupDesc:
		response, err := ofc.handleGroupStatsDescRequest(request.(ofp.IGroupDescStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IGroupDescStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)

	case ofp.OFPSTGroupFeatures:
		response, err := ofc.handleGroupFeatureStatsRequest(request.(ofp.IGroupFeaturesStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IGroupFeaturesStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTMeter:
		response, err := ofc.handleMeterStatsRequest(request.(ofp.IMeterStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IMeterStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTMeterConfig:
		response, err := ofc.handleMeterConfigStatsRequest(request.(ofp.IMeterConfigStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IMeterConfigStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTMeterFeatures:
		response, err := ofc.handleMeterFeatureStatsRequest(request.(ofp.IMeterFeaturesStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IMeterFeaturesStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTTableFeatures:
		response, err := ofc.handleTableFeaturesStatsRequest(request.(ofp.ITableFeaturesStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.ITableFeaturesStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTPortDesc:
		response, err := ofc.handlePortDescStatsRequest(request.(ofp.IPortDescStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IPortDescStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTExperimenter:
		response, err := ofc.handleExperimenterStatsRequest(request.(ofp.IExperimenterStatsRequest))
		if err != nil {
			return err
		}
		if log.GetDefaultLogLevel() <= log.DebugLevel {
			reqJs, _ := json.Marshal(request.(ofp.IExperimenterStatsRequest))
			resJs, _ := json.Marshal(response)
			log.Debugw("HandleStatsRequest GRPC",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	default:
		// TODO: Should return error
		return errors.New("unsupported-stats-request")
	}
	return nil
}

func (ofc *OFClient) handleDescStatsRequest(request ofp.IDescStatsRequest) (*ofp.DescStatsReply, error) {
	response := ofp.NewDescStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(request.GetVersion())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))

	resp, err := ofc.volthaClient.GetLogicalDevice(context.Background(), &common.ID{Id: ofc.DeviceID})
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
func (ofc *OFClient) handleFlowStatsRequest(request ofp.IFlowStatsRequest) (*ofp.FlowStatsReply, error) {
	response := ofp.NewFlowStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(4)
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	resp, err := ofc.volthaClient.ListLogicalDeviceFlows(context.Background(), &common.ID{Id: ofc.DeviceID})
	if err != nil {
		return nil, err
	}
	var flow []*ofp.FlowStatsEntry
	for _, item := range resp.GetItems() {
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
		entrySize := uint16(48)
		match := ofp.NewMatchV3()
		pbMatch := item.GetMatch()
		var fields []goloxi.IOxm
		match.SetType(uint16(pbMatch.GetType()))
		size := uint16(4)
		for _, oxmField := range pbMatch.GetOxmFields() {
			iOxm, oxmSize := parseOxm(
				oxmField.GetField().(*openflow_13.OfpOxmField_OfbField).OfbField,
				ofc.DeviceID)
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
		for _, ofpInstruction := range item.Instructions {
			instruction, size := parseInstructions(ofpInstruction, ofc.DeviceID)
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
func (ofc *OFClient) handleAggregateStatsRequest(request ofp.IAggregateStatsRequest) (*ofp.AggregateStatsReply, error) {
	response := ofp.NewAggregateStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	response.SetFlowCount(0)
	//TODO wire this to voltha core when it implements
	return response, nil
}
func (ofc *OFClient) handleGroupStatsRequest(request ofp.IGroupStatsRequest) (*ofp.GroupStatsReply, error) {
	response := ofp.NewGroupStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	reply, err := ofc.volthaClient.ListLogicalDeviceFlowGroups(context.Background(), &common.ID{Id: ofc.DeviceID})
	if err != nil {
		return nil, err
	}

	var groupStatsEntries []*ofp.GroupStatsEntry
	for _, item := range reply.GetItems() {
		stats := item.GetStats()
		var entry ofp.GroupStatsEntry
		entry.SetByteCount(stats.GetByteCount())
		entry.SetPacketCount(stats.GetPacketCount())
		entry.SetDurationNsec(stats.GetDurationNsec())
		entry.SetDurationSec(stats.GetDurationSec())
		entry.SetRefCount(stats.GetRefCount())
		entry.SetGroupId(stats.GetGroupId())
		bucketStats := stats.GetBucketStats()
		var bucketStatsList []*ofp.BucketCounter
		for _, bucketStat := range bucketStats {
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
func (ofc *OFClient) handleGroupStatsDescRequest(request ofp.IGroupDescStatsRequest) (*ofp.GroupDescStatsReply, error) {
	response := ofp.NewGroupDescStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	reply, err := ofc.volthaClient.ListLogicalDeviceFlowGroups(context.Background(), &common.ID{Id: ofc.DeviceID})
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
			var bucketList []ofp.IBucket
			for j:=0;j<len(buckets);j++{

			}

			groupDesc.SetBuckets(bucketList)
		*/
		groupDescStatsEntries = append(groupDescStatsEntries, &groupDesc)
	}
	response.SetEntries(groupDescStatsEntries)
	return response, nil
}
func (ofc *OFClient) handleGroupFeatureStatsRequest(request ofp.IGroupFeaturesStatsRequest) (*ofp.GroupFeaturesStatsReply, error) {
	response := ofp.NewGroupFeaturesStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}
func (ofc *OFClient) handleMeterStatsRequest(request ofp.IMeterStatsRequest) (*ofp.MeterStatsReply, error) {
	response := ofp.NewMeterStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	resp, err := ofc.volthaClient.ListLogicalDeviceMeters(context.Background(), &common.ID{Id: ofc.DeviceID})
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
func (ofc *OFClient) handleMeterConfigStatsRequest(request ofp.IMeterConfigStatsRequest) (*ofp.MeterConfigStatsReply, error) {
	response := ofp.NewMeterConfigStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}
func (ofc *OFClient) handleTableFeaturesStatsRequest(request ofp.ITableFeaturesStatsRequest) (*ofp.TableFeaturesStatsReply, error) {
	response := ofp.NewTableFeaturesStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}
func (ofc *OFClient) handleTableStatsRequest(request ofp.ITableStatsRequest) (*ofp.TableStatsReply, error) {
	var response = ofp.NewTableStatsReply()
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	return response, nil
}
func (ofc *OFClient) handleQueueStatsRequest(request ofp.IQueueStatsRequest) (*ofp.QueueStatsReply, error) {
	response := ofp.NewQueueStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}
func (ofc *OFClient) handlePortStatsRequest(request ofp.IPortStatsRequest) (*ofp.PortStatsReply, error) {
	response := ofp.NewPortStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(request.GetVersion())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	reply, err := ofc.volthaClient.ListLogicalDevicePorts(context.Background(), &common.ID{Id: ofc.DeviceID})
	//reply,err := ofc.volthaClient.GetLogicalDevicePort(context.Background(),&common.ID{Id:ofc.DeviceID})
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

func (ofc *OFClient) handlePortDescStatsRequest(request ofp.IPortDescStatsRequest) (*ofp.PortDescStatsReply, error) {
	response := ofp.NewPortDescStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	logicalDevice, err := ofc.volthaClient.GetLogicalDevice(context.Background(), &common.ID{Id: ofc.DeviceID})
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
func (ofc *OFClient) handleMeterFeatureStatsRequest(request ofp.IMeterFeaturesStatsRequest) (*ofp.MeterFeaturesStatsReply, error) {
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
func (ofc *OFClient) handleExperimenterStatsRequest(request ofp.IExperimenterStatsRequest) (*ofp.ExperimenterStatsReply, error) {
	response := ofp.NewExperimenterStatsReply(request.GetExperimenter())
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}
