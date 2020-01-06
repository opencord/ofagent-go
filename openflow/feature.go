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

	"github.com/opencord/ofagent-go/settings"

	ofp "github.com/donNewtonAlpha/goloxi/of13"
	l "github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/common"
	pb "github.com/opencord/voltha-protos/v2/go/voltha"
)

func handleFeatureRequest(request *ofp.FeaturesRequest, DeviceID string, client *Client) error {
	if settings.GetDebug(DeviceID) {
		js, _ := json.Marshal(request)
		logger.Debugw("handleFeatureRequest called", l.Fields{"DeviceID": DeviceID, "request": js})
	}
	var grpcClient = *getGrpcClient()
	var id = common.ID{Id: DeviceID}
	logicalDevice, err := grpcClient.GetLogicalDevice(context.Background(), &id)
	reply := createFeaturesRequestReply(request.GetXid(), logicalDevice)

	if settings.GetDebug(DeviceID) {
		js, _ := json.Marshal(reply)
		logger.Debugw("handleFeatureRequestReturn", l.Fields{"DeviceID": DeviceID, "reply": js})
	}
	err = client.SendMessage(reply)
	return err
}
func createFeaturesRequestReply(xid uint32, device *pb.LogicalDevice) *ofp.FeaturesReply {

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
