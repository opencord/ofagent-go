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

	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/common"
	"github.com/opencord/voltha-protos/v2/go/voltha"
)

func (ofc *OFClient) handleFeatureRequest(request *ofp.FeaturesRequest) error {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(request)
		logger.Debugw("handleFeatureRequest called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"request":   js})
	}
	var id = common.ID{Id: ofc.DeviceID}
	logicalDevice, err := ofc.volthaClient.GetLogicalDevice(context.Background(), &id)
	reply := ofc.createFeaturesRequestReply(request.GetXid(), logicalDevice)

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
func (ofc *OFClient) createFeaturesRequestReply(xid uint32, device *voltha.LogicalDevice) *ofp.FeaturesReply {

	featureReply := ofp.NewFeaturesReply()
	featureReply.SetVersion(4)
	featureReply.SetXid(xid)
	features := device.GetSwitchFeatures()
	featureReply.SetDatapathId(device.GetDatapathId())
	featureReply.SetNBuffers(features.GetNBuffers())
	featureReply.SetNTables(uint8(features.GetNTables()))
	featureReply.SetAuxiliaryId(uint8(features.GetAuxiliaryId()))
	capabilities := features.GetCapabilities()
	featureReply.SetCapabilities(ofp.Capabilities(capabilities))
	return featureReply
}
