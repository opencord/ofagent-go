/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mock

import (
	"context"

	. "github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/voltha-protos/v3/go/common"
	"github.com/opencord/voltha-protos/v3/go/omci"
	"github.com/opencord/voltha-protos/v3/go/openflow_13"
	. "github.com/opencord/voltha-protos/v3/go/voltha"
	"google.golang.org/grpc"
)

type MockVolthaClient struct {
	LogicalDeviceFlows openflow_13.Flows
	LogicalPorts       LogicalPorts
}

func (c MockVolthaClient) GetMembership(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Membership, error) {
	return &Membership{}, nil
}

func (c MockVolthaClient) UpdateMembership(ctx context.Context, in *Membership, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) GetVoltha(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Voltha, error) {
	return &Voltha{}, nil
}

func (c MockVolthaClient) ListCoreInstances(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*CoreInstances, error) {
	return &CoreInstances{}, nil
}

func (c MockVolthaClient) GetCoreInstance(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*CoreInstance, error) {
	return &CoreInstance{}, nil
}

func (c MockVolthaClient) ListAdapters(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Adapters, error) {
	return &Adapters{}, nil
}

func (c MockVolthaClient) ListLogicalDevices(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*LogicalDevices, error) {
	return &LogicalDevices{}, nil
}

func (c MockVolthaClient) GetLogicalDevice(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*LogicalDevice, error) {
	return &LogicalDevice{}, nil
}

func (c MockVolthaClient) ListLogicalDevicePorts(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*LogicalPorts, error) {
	return &c.LogicalPorts, nil
}

func (c MockVolthaClient) GetLogicalDevicePort(ctx context.Context, in *LogicalPortId, opts ...grpc.CallOption) (*LogicalPort, error) {
	return &LogicalPort{}, nil
}

func (c MockVolthaClient) EnableLogicalDevicePort(ctx context.Context, in *LogicalPortId, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) DisableLogicalDevicePort(ctx context.Context, in *LogicalPortId, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) ListLogicalDeviceFlows(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*openflow_13.Flows, error) {
	return &c.LogicalDeviceFlows, nil
}

func (c MockVolthaClient) UpdateLogicalDeviceFlowTable(ctx context.Context, in *openflow_13.FlowTableUpdate, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) UpdateLogicalDeviceMeterTable(ctx context.Context, in *openflow_13.MeterModUpdate, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) ListLogicalDeviceMeters(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*openflow_13.Meters, error) {
	return &openflow_13.Meters{}, nil
}

func (c MockVolthaClient) ListLogicalDeviceFlowGroups(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*openflow_13.FlowGroups, error) {
	return &openflow_13.FlowGroups{}, nil
}

func (c MockVolthaClient) UpdateLogicalDeviceFlowGroupTable(ctx context.Context, in *openflow_13.FlowGroupTableUpdate, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) ListDevices(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Devices, error) {
	return &Devices{}, nil
}

func (c MockVolthaClient) ListDeviceIds(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*common.IDs, error) {
	return &common.IDs{}, nil
}

func (c MockVolthaClient) ReconcileDevices(ctx context.Context, in *common.IDs, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) GetDevice(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*Device, error) {
	return &Device{}, nil
}

func (c MockVolthaClient) CreateDevice(ctx context.Context, in *Device, opts ...grpc.CallOption) (*Device, error) {
	return &Device{}, nil
}

func (c MockVolthaClient) EnableDevice(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) DisableDevice(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) RebootDevice(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) DeleteDevice(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) DownloadImage(ctx context.Context, in *ImageDownload, opts ...grpc.CallOption) (*common.OperationResp, error) {
	return &common.OperationResp{}, nil
}

func (c MockVolthaClient) GetImageDownloadStatus(ctx context.Context, in *ImageDownload, opts ...grpc.CallOption) (*ImageDownload, error) {
	return &ImageDownload{}, nil
}

func (c MockVolthaClient) GetImageDownload(ctx context.Context, in *ImageDownload, opts ...grpc.CallOption) (*ImageDownload, error) {
	return &ImageDownload{}, nil
}

func (c MockVolthaClient) ListImageDownloads(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*ImageDownloads, error) {
	return &ImageDownloads{}, nil
}

func (c MockVolthaClient) CancelImageDownload(ctx context.Context, in *ImageDownload, opts ...grpc.CallOption) (*common.OperationResp, error) {
	return &common.OperationResp{}, nil
}

func (c MockVolthaClient) ActivateImageUpdate(ctx context.Context, in *ImageDownload, opts ...grpc.CallOption) (*common.OperationResp, error) {
	return &common.OperationResp{}, nil
}

func (c MockVolthaClient) RevertImageUpdate(ctx context.Context, in *ImageDownload, opts ...grpc.CallOption) (*common.OperationResp, error) {
	return &common.OperationResp{}, nil
}

func (c MockVolthaClient) ListDevicePorts(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*Ports, error) {
	return &Ports{}, nil
}

func (c MockVolthaClient) ListDevicePmConfigs(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*PmConfigs, error) {
	return &PmConfigs{}, nil
}

func (c MockVolthaClient) UpdateDevicePmConfigs(ctx context.Context, in *PmConfigs, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) ListDeviceFlows(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*openflow_13.Flows, error) {
	return &openflow_13.Flows{}, nil
}

func (c MockVolthaClient) ListDeviceFlowGroups(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*openflow_13.FlowGroups, error) {
	return &openflow_13.FlowGroups{}, nil
}

func (c MockVolthaClient) ListDeviceTypes(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*DeviceTypes, error) {
	return &DeviceTypes{}, nil
}

func (c MockVolthaClient) GetDeviceType(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*DeviceType, error) {
	return &DeviceType{}, nil
}

func (c MockVolthaClient) ListDeviceGroups(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*DeviceGroups, error) {
	return &DeviceGroups{}, nil
}

func (c MockVolthaClient) StreamPacketsOut(ctx context.Context, opts ...grpc.CallOption) (VolthaService_StreamPacketsOutClient, error) {
	return nil, nil
}

func (c MockVolthaClient) ReceivePacketsIn(ctx context.Context, in *Empty, opts ...grpc.CallOption) (VolthaService_ReceivePacketsInClient, error) {
	return nil, nil
}

func (c MockVolthaClient) ReceiveChangeEvents(ctx context.Context, in *Empty, opts ...grpc.CallOption) (VolthaService_ReceiveChangeEventsClient, error) {
	return nil, nil
}

func (c MockVolthaClient) GetDeviceGroup(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*DeviceGroup, error) {
	return &DeviceGroup{}, nil
}

func (c MockVolthaClient) CreateEventFilter(ctx context.Context, in *EventFilter, opts ...grpc.CallOption) (*EventFilter, error) {
	return &EventFilter{}, nil
}

func (c MockVolthaClient) GetEventFilter(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*EventFilters, error) {
	return &EventFilters{}, nil
}

func (c MockVolthaClient) UpdateEventFilter(ctx context.Context, in *EventFilter, opts ...grpc.CallOption) (*EventFilter, error) {
	return &EventFilter{}, nil
}

func (c MockVolthaClient) DeleteEventFilter(ctx context.Context, in *EventFilter, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) ListEventFilters(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*EventFilters, error) {
	return &EventFilters{}, nil
}

func (c MockVolthaClient) GetImages(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*Images, error) {
	return &Images{}, nil
}

func (c MockVolthaClient) SelfTest(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*SelfTestResponse, error) {
	return &SelfTestResponse{}, nil
}

func (c MockVolthaClient) GetMibDeviceData(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*omci.MibDeviceData, error) {
	return &omci.MibDeviceData{}, nil
}

func (c MockVolthaClient) GetAlarmDeviceData(ctx context.Context, in *common.ID, opts ...grpc.CallOption) (*omci.AlarmDeviceData, error) {
	return &omci.AlarmDeviceData{}, nil
}

func (c MockVolthaClient) SimulateAlarm(ctx context.Context, in *SimulateAlarmRequest, opts ...grpc.CallOption) (*common.OperationResp, error) {
	return &common.OperationResp{}, nil
}

func (c MockVolthaClient) Subscribe(ctx context.Context, in *OfAgentSubscriber, opts ...grpc.CallOption) (*OfAgentSubscriber, error) {
	return &OfAgentSubscriber{}, nil
}

func (c MockVolthaClient) EnablePort(ctx context.Context, in *Port, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) DisablePort(ctx context.Context, in *Port, opts ...grpc.CallOption) (*Empty, error) {
	return &Empty{}, nil
}

func (c MockVolthaClient) GetExtValue(ctx context.Context, in *common.ValueSpecifier, opts ...grpc.CallOption) (*common.ReturnValues, error) {
	return &common.ReturnValues{}, nil
}

func (c MockVolthaClient) StartOmciTestAction(ctx context.Context, in *OmciTestRequest, opts ...grpc.CallOption) (*TestResponse, error) {
	return &TestResponse{}, nil
}
