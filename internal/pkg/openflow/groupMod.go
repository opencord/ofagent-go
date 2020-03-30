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
	"unsafe"

	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-protos/v3/go/openflow_13"
	"github.com/opencord/voltha-protos/v3/go/voltha"
)

func (ofc *OFConnection) handleGroupMod(groupMod *ofp.GroupMod, msgbuf *[]byte) {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(groupMod)
		logger.Debugw("handleGroupMod called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"params":    js})
	}

	if ofc.VolthaClient == nil {
		logger.Errorw("no-voltha-connection",
			log.Fields{"device-id": ofc.DeviceID})
		return
	}

	// Construct the buckets
	var buckets []*voltha.OfpBucket
	for _, bucket := range groupMod.GetBuckets() {
		var ofpBucket voltha.OfpBucket
		ofpBucket.Weight = uint32(bucket.Weight)
		ofpBucket.WatchPort = uint32(bucket.WatchPort)
		ofpBucket.WatchGroup = bucket.WatchGroup
		for _, action := range bucket.Actions {
			ofpBucket.Actions = append(ofpBucket.Actions, extractAction(action))
		}

		buckets = append(buckets, &ofpBucket)
	}

	// Construct the request
	groupUpdate := openflow_13.FlowGroupTableUpdate{
		Id: ofc.DeviceID,
		GroupMod: &voltha.OfpGroupMod{
			Command: func(cmd ofp.GroupModCommand) voltha.OfpGroupModCommand {
				switch cmd {
				case ofp.OFPGCAdd:
					return voltha.OfpGroupModCommand_OFPGC_ADD
				case ofp.OFPGCModify:
					return voltha.OfpGroupModCommand_OFPGC_MODIFY
				case ofp.OFPGCDelete:
					return voltha.OfpGroupModCommand_OFPGC_DELETE
				}
				return voltha.OfpGroupModCommand_OFPGC_ADD
			}(groupMod.Command),
			Type: func(groupType ofp.GroupType) voltha.OfpGroupType {
				switch groupType {
				case ofp.OFPGTAll:
					return voltha.OfpGroupType_OFPGT_ALL
				case ofp.OFPGTSelect:
					return voltha.OfpGroupType_OFPGT_SELECT
				case ofp.OFPGTIndirect:
					return voltha.OfpGroupType_OFPGT_INDIRECT
				case ofp.OFPGTFf:
					return voltha.OfpGroupType_OFPGT_FF
				}
				return voltha.OfpGroupType_OFPGT_ALL
			}(groupMod.GroupType),
			GroupId: groupMod.GroupId,
			Buckets: buckets,
		},
	}
	if logger.V(log.DebugLevel) {
		groupUpdateJs, _ := json.Marshal(groupUpdate)
		logger.Debugf("GroupMod being sent to Voltha",
			log.Fields{
				"device-id":         ofc.DeviceID,
				"group-mod-object":  groupUpdate,
				"group-mod-request": groupUpdateJs})
	}
	if _, err := ofc.VolthaClient.UpdateLogicalDeviceFlowGroupTable(context.Background(), &groupUpdate); err != nil {
		logger.Errorw("Error calling groupMod ",
			log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
		// Report failure to controller
		message := ofp.NewGroupModFailedErrorMsg()
		message.SetXid(groupMod.Xid)
		message.SetCode(ofp.OFPGMFCBadCommand)
		//OF 1.3
		message.SetVersion(4)
		message.SetData(*msgbuf)
		message.Length = uint16(unsafe.Sizeof(*message))
		err := ofc.SendMessage(message)
		if err != nil {
			logger.Errorw("Error reporting failure of GroupUpdate to controller",
				log.Fields{
					"device-id": ofc.DeviceID,
					"error":     err})
		}
	}
}
