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
	"github.com/opencord/voltha-protos/go/common"
	pb "github.com/opencord/voltha-protos/go/voltha"
	ofp "github.com/skydive-project/goloxi/of13"
	"log"
)

func handleFeatureRequest(request *ofp.FeaturesRequest, deviceId string, client *Client) error {
	message, _ := json.Marshal(request)
	log.Printf("handleFeatureRequest called with %s\n ", message)
	var grpcClient = *getGrpcClient()
	var id = common.ID{Id: deviceId}
	logicalDevice, err := grpcClient.GetLogicalDevice(context.Background(), &id)
	reply := createFeaturesRequestReply(request.GetXid(), logicalDevice)
	err = client.SendMessage(reply)
	return err
}
func createFeaturesRequestReply(xid uint32, device *pb.LogicalDevice) *ofp.FeaturesReply {

	featureReply := ofp.NewFeaturesReply()
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
