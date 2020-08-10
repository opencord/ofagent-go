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

package ofagent

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/ofagent-go/internal/pkg/openflow"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
)

func (ofa *OFAgent) synchronizeDeviceList(ctx context.Context) {
	// Refresh once to get everything started
	ofa.refreshDeviceList(ctx)

	tick := time.NewTicker(ofa.DeviceListRefreshInterval)
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-tick.C:
			ofa.refreshDeviceList(ctx)
		}
	}
	tick.Stop()
}

func (ofa *OFAgent) refreshDeviceList(ctx context.Context) {
	span, ctx := log.CreateChildSpan(ctx, "refresh-device-list")
	defer span.Finish()

	// If we exit, assume disconnected
	if ofa.volthaClient == nil {
		logger.Error(ctx, "no-voltha-connection")
		ofa.events <- ofaEventVolthaDisconnected
		return
	}
	deviceList, err := ofa.volthaClient.Get().ListLogicalDevices(log.WithSpanFromContext(context.Background(), ctx), &empty.Empty{})
	if err != nil {
		logger.Errorw(ctx, "ofagent failed to query device list from voltha",
			log.Fields{"error": err})
		ofa.events <- ofaEventVolthaDisconnected
		return
	}
	devices := deviceList.GetItems()

	var toAdd []string
	var toDel []string
	var deviceIDMap = make(map[string]string)
	for i := 0; i < len(devices); i++ {
		deviceID := devices[i].GetId()
		deviceIDMap[deviceID] = deviceID
		if ofa.clientMap[deviceID] == nil {
			toAdd = append(toAdd, deviceID)
		}
	}
	for key := range ofa.clientMap {
		if deviceIDMap[key] == "" {
			toDel = append(toDel, key)
		}
	}
	logger.Debugw(ctx, "GrpcClient refreshDeviceList", log.Fields{"ToAdd": toAdd, "ToDel": toDel})
	for i := 0; i < len(toAdd); i++ {
		ofa.addOFClient(ctx, toAdd[i]) // client is started in addOFClient
	}
	for i := 0; i < len(toDel); i++ {
		ofa.clientMap[toDel[i]].Stop()
		ofa.mapLock.Lock()
		delete(ofa.clientMap, toDel[i])
		ofa.mapLock.Unlock()
	}
}

func (ofa *OFAgent) addOFClient(ctx context.Context, deviceID string) *openflow.OFClient {
	logger.Debugw(ctx, "GrpcClient addClient called ", log.Fields{"device-id": deviceID})
	ofa.mapLock.Lock()
	ofc := ofa.clientMap[deviceID]
	if ofc == nil {
		ofc = openflow.NewOFClient(ctx, &openflow.OFClient{
			DeviceID:              deviceID,
			OFControllerEndPoints: ofa.OFControllerEndPoints,
			VolthaClient:          ofa.volthaClient,
			PacketOutChannel:      ofa.packetOutChannel,
			ConnectionMaxRetries:  ofa.ConnectionMaxRetries,
			ConnectionRetryDelay:  ofa.ConnectionRetryDelay,
		})
		ofc.Run(context.Background())
		ofa.clientMap[deviceID] = ofc
	}
	ofa.mapLock.Unlock()
	logger.Debugw(ctx, "Finished with addClient", log.Fields{"deviceID": deviceID})
	return ofc
}

func (ofa *OFAgent) getOFClient(ctx context.Context, deviceID string) *openflow.OFClient {
	ofc := ofa.clientMap[deviceID]
	if ofc == nil {
		ofc = ofa.addOFClient(ctx, deviceID)
	}
	return ofc
}
