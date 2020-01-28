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
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-protos/v3/go/common"
)

func (ofc *OFClient) handleFeatureRequest(request *ofp.FeaturesRequest) error {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(request)
		logger.Debugw("handleFeatureRequest called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"request":   js})
	}
	if ofc.VolthaClient == nil {
		return NoVolthaConnectionError
	}
	var id = common.ID{Id: ofc.DeviceID}
	logicalDevice, err := ofc.VolthaClient.GetLogicalDevice(context.Background(), &id)
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
		logger.Debugw("handleFeatureRequestReturn",
			log.Fields{
				"device-id": ofc.DeviceID,
				"reply":     js})
	}
	err = ofc.SendMessage(reply)
	return err
}
