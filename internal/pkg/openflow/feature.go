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

	ofp "github.com/opencord/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v5/pkg/log"
	"github.com/opencord/voltha-protos/v4/go/common"
)

func (ofc *OFConnection) handleFeatureRequest(ctx context.Context, request *ofp.FeaturesRequest) error {
	span, ctx := log.CreateChildSpan(ctx, "openflow-feature")
	defer span.Finish()

	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(request)
		logger.Debugw(ctx, "handleFeatureRequest called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"request":   js})
	}
	volthaClient := ofc.VolthaClient.Get()
	if volthaClient == nil {
		return NoVolthaConnectionError
	}
	var id = common.ID{Id: ofc.DeviceID}
	logicalDevice, err := volthaClient.GetLogicalDevice(log.WithSpanFromContext(context.Background(), ctx), &id)
	if err != nil {
		return err
	}
	reply := ofp.NewFeaturesReply()
	reply.SetVersion(4)
	reply.SetXid(request.GetXid())
	features := logicalDevice.GetSwitchFeatures()
	reply.SetDatapathId(logicalDevice.GetDatapathId())
	reply.SetNBuffers(features.GetNBuffers())
	reply.SetNTables(uint8(features.GetNTables()))
	reply.SetAuxiliaryId(uint8(features.GetAuxiliaryId()))
	capabilities := features.GetCapabilities()
	reply.SetCapabilities(ofp.Capabilities(capabilities))

	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(reply)
		logger.Debugw(ctx, "handleFeatureRequestReturn",
			log.Fields{
				"device-id": ofc.DeviceID,
				"reply":     js})
	}
	err = ofc.SendMessage(ctx, reply)
	return err
}
