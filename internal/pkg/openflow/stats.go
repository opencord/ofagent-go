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
	"github.com/donNewtonAlpha/goloxi"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/common"
	"github.com/opencord/voltha-protos/v2/go/openflow_13"
	"net"
	"unsafe"
)

func (ofc *OFClient) handleStatsRequest(request ofp.IHeader, statType uint16) error {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(request)
		logger.Debugw("handleStatsRequest called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"stat-type": statType,
				"request":   js})
	}

	switch statType {
	case ofp.OFPSTDesc:
		statsReq := request.(*ofp.DescStatsRequest)
		response, err := ofc.handleDescStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-desc",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTFlow:
		statsReq := request.(*ofp.FlowStatsRequest)
		response, err := ofc.handleFlowStatsRequest(statsReq)
		if err != nil {
			return err
		}
		response.Length = uint16(unsafe.Sizeof(*response))
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-flow",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)

	case ofp.OFPSTAggregate:
		statsReq := request.(*ofp.AggregateStatsRequest)
		response, err := ofc.handleAggregateStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-aggregate",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTTable:
		statsReq := request.(*ofp.TableStatsRequest)
		response, e := ofc.handleTableStatsRequest(statsReq)
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-table",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		if e != nil {
			return e
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTPort:
		statsReq := request.(*ofp.PortStatsRequest)
		response, err := ofc.handlePortStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-port",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTQueue:
		statsReq := request.(*ofp.QueueStatsRequest)
		response, err := ofc.handleQueueStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-queue",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTGroup:
		statsReq := request.(*ofp.GroupStatsRequest)
		response, err := ofc.handleGroupStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-group",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		ofc.SendMessage(response)
	case ofp.OFPSTGroupDesc:
		statsReq := request.(*ofp.GroupDescStatsRequest)
		response, err := ofc.handleGroupStatsDescRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-group-desc",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)

	case ofp.OFPSTGroupFeatures:
		statsReq := request.(*ofp.GroupFeaturesStatsRequest)
		response, err := ofc.handleGroupFeatureStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-group-features",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTMeter:
		statsReq := request.(*ofp.MeterStatsRequest)
		response, err := ofc.handleMeterStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-meter",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTMeterConfig:
		statsReq := request.(*ofp.MeterConfigStatsRequest)
		response, err := ofc.handleMeterConfigStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-meter-config",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTMeterFeatures:
		statsReq := request.(*ofp.MeterFeaturesStatsRequest)
		response, err := ofc.handleMeterFeatureStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-meter-features",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTTableFeatures:
		statsReq := request.(*ofp.TableFeaturesStatsRequest)
		response, err := ofc.handleTableFeaturesStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-table-features",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	case ofp.OFPSTPortDesc:
		statsReq := request.(*ofp.PortDescStatsRequest)
		response, err := ofc.handlePortDescStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-port-desc",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)

	case ofp.OFPSTExperimenter:
		statsReq := request.(*ofp.ExperimenterStatsRequest)
		response, err := ofc.handleExperimenterStatsRequest(statsReq)
		if err != nil {
			return err
		}
		if logger.V(log.DebugLevel) {
			reqJs, _ := json.Marshal(statsReq)
			resJs, _ := json.Marshal(response)
			logger.Debugw("handle-stats-request-experimenter",
				log.Fields{
					"device-id": ofc.DeviceID,
					"request":   reqJs,
					"response":  resJs})
		}
		return ofc.SendMessage(response)
	}
	return nil
}

func (ofc *OFClient) handleDescStatsRequest(request *ofp.DescStatsRequest) (*ofp.DescStatsReply, error) {
	response := ofp.NewDescStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(request.GetVersion())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))

	resp, err := ofc.VolthaClient.GetLogicalDevice(context.Background(),
		&common.ID{Id: ofc.DeviceID})
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

func (ofc *OFClient) handleFlowStatsRequest(request *ofp.FlowStatsRequest) (*ofp.FlowStatsReply, error) {
	response := ofp.NewFlowStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(4)
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	resp, err := ofc.VolthaClient.ListLogicalDeviceFlows(context.Background(),
		&common.ID{Id: ofc.DeviceID})
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
		match.SetType(uint16(pbMatch.GetType()))
		size := uint16(4)
		var fields []goloxi.IOxm
		for _, oxmField := range pbMatch.GetOxmFields() {
			field := oxmField.GetField()
			ofbField := field.(*openflow_13.OfpOxmField_OfbField).OfbField
			iOxm, oxmSize := parseOxm(ofbField, ofc.DeviceID)
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

func (ofc *OFClient) handleAggregateStatsRequest(request *ofp.AggregateStatsRequest) (*ofp.AggregateStatsReply, error) {
	response := ofp.NewAggregateStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	response.SetFlowCount(0)
	//TODO wire this to voltha core when it implements
	return response, nil
}

func (ofc *OFClient) handleGroupStatsRequest(request *ofp.GroupStatsRequest) (*ofp.GroupStatsReply, error) {
	response := ofp.NewGroupStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	reply, err := ofc.VolthaClient.ListLogicalDeviceFlowGroups(context.Background(),
		&common.ID{Id: ofc.DeviceID})
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
		var bucketStatsList []*ofp.BucketCounter
		for _, bucketStat := range stats.GetBucketStats() {
			bucketCounter := ofp.BucketCounter{}
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

func (ofc *OFClient) handleGroupStatsDescRequest(request *ofp.GroupDescStatsRequest) (*ofp.GroupDescStatsReply, error) {
	response := ofp.NewGroupDescStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	reply, err := ofc.VolthaClient.ListLogicalDeviceFlowGroups(context.Background(),
		&common.ID{Id: ofc.DeviceID})
	if err != nil {
		return nil, err
	}
	var groupDescStatsEntries []*ofp.GroupDescStatsEntry
	for _, item := range reply.GetItems() {
		stats := item.GetStats()
		var groupDesc ofp.GroupDescStatsEntry
		groupDesc.SetGroupId(stats.GetGroupId())
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

func (ofc *OFClient) handleGroupFeatureStatsRequest(request *ofp.GroupFeaturesStatsRequest) (*ofp.GroupFeaturesStatsReply, error) {
	response := ofp.NewGroupFeaturesStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}

func (ofc *OFClient) handleMeterStatsRequest(request *ofp.MeterStatsRequest) (*ofp.MeterStatsReply, error) {
	response := ofp.NewMeterStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	resp, err := ofc.VolthaClient.ListLogicalDeviceMeters(context.Background(),
		&common.ID{Id: ofc.DeviceID})
	if err != nil {
		return nil, err
	}
	size := uint16(5) // size of stats header
	var meterStats []*ofp.MeterStats
	for _, item := range resp.Items {
		entrySize := uint16(40) // size of entry header
		meterStat := ofp.NewMeterStats()
		stats := item.Stats
		meterStat.DurationNsec = stats.DurationNsec
		meterStat.DurationSec = stats.DurationSec
		meterStat.ByteInCount = stats.ByteInCount
		meterStat.FlowCount = stats.FlowCount
		meterStat.MeterId = stats.MeterId
		var bandStats []*ofp.MeterBandStats
		for _, bStat := range stats.BandStats {
			bandStat := ofp.NewMeterBandStats()
			bandStat.ByteBandCount = bStat.ByteBandCount
			bandStat.PacketBandCount = bStat.PacketBandCount
			bandStats = append(bandStats, bandStat)
			entrySize += uint16(16) // size of each band stat
		}
		meterStat.SetBandStats(bandStats)
		meterStat.Len = entrySize
		meterStats = append(meterStats, meterStat)
		size += entrySize
	}
	response.SetEntries(meterStats)
	response.SetLength(size)
	return response, nil
}

func (ofc *OFClient) handleMeterConfigStatsRequest(request *ofp.MeterConfigStatsRequest) (*ofp.MeterConfigStatsReply, error) {
	response := ofp.NewMeterConfigStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}

func (ofc *OFClient) handleTableFeaturesStatsRequest(request *ofp.TableFeaturesStatsRequest) (*ofp.TableFeaturesStatsReply, error) {
	response := ofp.NewTableFeaturesStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}

func (ofc *OFClient) handleTableStatsRequest(request *ofp.TableStatsRequest) (*ofp.TableStatsReply, error) {
	var response = ofp.NewTableStatsReply()
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	return response, nil
}

func (ofc *OFClient) handleQueueStatsRequest(request *ofp.QueueStatsRequest) (*ofp.QueueStatsReply, error) {
	response := ofp.NewQueueStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}

func (ofc *OFClient) handlePortStatsRequest(request *ofp.PortStatsRequest) (*ofp.PortStatsReply, error) {
	response := ofp.NewPortStatsReply()
	response.SetXid(request.GetXid())
	response.SetVersion(request.GetVersion())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	reply, err := ofc.VolthaClient.ListLogicalDevicePorts(context.Background(),
		&common.ID{Id: ofc.DeviceID})
	if err != nil {
		return nil, err
	}
	var entries []*ofp.PortStatsEntry
	if request.GetPortNo() == 0xffffffff { //all ports
		for _, port := range reply.GetItems() {
			entries = append(entries, parsePortStats(port))
		}
	} else { //find right port that is requested
		for _, port := range reply.GetItems() {
			if port.GetOfpPortStats().GetPortNo() == uint32(request.GetPortNo()) {
				entries = append(entries, parsePortStats(port))
			}
		}
	}
	response.SetEntries(entries)
	return response, nil
}

func (ofc *OFClient) handlePortDescStatsRequest(request *ofp.PortDescStatsRequest) (*ofp.PortDescStatsReply, error) {
	response := ofp.NewPortDescStatsReply()
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	logicalDevice, err := ofc.VolthaClient.GetLogicalDevice(context.Background(),
		&common.ID{Id: ofc.DeviceID})
	if err != nil {
		return nil, err
	}
	var entries []*ofp.PortDesc
	for _, port := range logicalDevice.GetPorts() {
		ofpPort := port.GetOfpPort()
		var entry ofp.PortDesc
		entry.SetPortNo(ofp.Port(ofpPort.GetPortNo()))

		var octets []byte
		for _, val := range ofpPort.GetHwAddr() {
			octets = append(octets, byte(val))
		}
		hwAddr := net.HardwareAddr(octets)
		entry.SetHwAddr(hwAddr)
		entry.SetName(PadString(ofpPort.GetName(), 16))
		entry.SetConfig(ofp.PortConfig(ofpPort.GetConfig()))
		entry.SetState(ofp.PortState(ofpPort.GetState()))
		entry.SetCurr(ofp.PortFeatures(ofpPort.GetCurr()))
		entry.SetAdvertised(ofp.PortFeatures(ofpPort.GetAdvertised()))
		entry.SetSupported(ofp.PortFeatures(ofpPort.GetSupported()))
		entry.SetPeer(ofp.PortFeatures(ofpPort.GetPeer()))
		entry.SetCurrSpeed(ofpPort.GetCurrSpeed())
		entry.SetMaxSpeed(ofpPort.GetMaxSpeed())

		entries = append(entries, &entry)
	}

	response.SetEntries(entries)
	//TODO call voltha and get port descriptions etc
	return response, nil

}

func (ofc *OFClient) handleMeterFeatureStatsRequest(request *ofp.MeterFeaturesStatsRequest) (*ofp.MeterFeaturesStatsReply, error) {
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

func (ofc *OFClient) handleExperimenterStatsRequest(request *ofp.ExperimenterStatsRequest) (*ofp.ExperimenterStatsReply, error) {
	response := ofp.NewExperimenterStatsReply(request.GetExperimenter())
	response.SetVersion(request.GetVersion())
	response.SetXid(request.GetXid())
	response.SetFlags(ofp.StatsReplyFlags(request.GetFlags()))
	//TODO wire this to voltha core when it implements
	return response, nil
}
