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
	"github.com/opencord/goloxi"
	ofp "github.com/opencord/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v4/pkg/log"
	"github.com/opencord/voltha-protos/v4/go/openflow_13"
	"github.com/opencord/voltha-protos/v4/go/voltha"
)

func (ofc *OFConnection) handleGroupMod(ctx context.Context, groupMod ofp.IGroupMod) {
	span, ctx := log.CreateChildSpan(ctx, "openflow-group-modification")
	defer span.Finish()

	volthaClient := ofc.VolthaClient.Get()
	if volthaClient == nil {
		logger.Errorw(ctx, "no-voltha-connection",
			log.Fields{"device-id": ofc.DeviceID})
		return
	}

	groupUpdate := &openflow_13.FlowGroupTableUpdate{
		Id: ofc.DeviceID,
		GroupMod: &voltha.OfpGroupMod{
			Command: openflowGroupModCommandToVoltha(ctx, groupMod.GetCommand()),
			Type:    openflowGroupTypeToVoltha(ctx, groupMod.GetGroupType()),
			GroupId: groupMod.GetGroupId(),
			Buckets: openflowBucketsToVoltha(groupMod.GetBuckets()),
		},
	}

	_, err := volthaClient.UpdateLogicalDeviceFlowGroupTable(log.WithSpanFromContext(context.Background(), ctx), groupUpdate)
	if err != nil {
		logger.Errorw(ctx, "Error updating group table",
			log.Fields{"device-id": ofc.DeviceID, "error": err})
	}

}

func openflowGroupModCommandToVoltha(ctx context.Context, command ofp.GroupModCommand) openflow_13.OfpGroupModCommand {
	switch command {
	case ofp.OFPGCAdd:
		return openflow_13.OfpGroupModCommand_OFPGC_ADD
	case ofp.OFPGCModify:
		return openflow_13.OfpGroupModCommand_OFPGC_MODIFY
	case ofp.OFPGCDelete:
		return openflow_13.OfpGroupModCommand_OFPGC_DELETE
	}
	logger.Errorw(ctx, "Unknown group mod command", log.Fields{"command": command})
	return 0
}

func openflowGroupTypeToVoltha(ctx context.Context, t ofp.GroupType) openflow_13.OfpGroupType {
	switch t {
	case ofp.OFPGTAll:
		return openflow_13.OfpGroupType_OFPGT_ALL
	case ofp.OFPGTSelect:
		return openflow_13.OfpGroupType_OFPGT_SELECT
	case ofp.OFPGTIndirect:
		return openflow_13.OfpGroupType_OFPGT_INDIRECT
	case ofp.OFPGTFf:
		return openflow_13.OfpGroupType_OFPGT_FF
	}
	logger.Errorw(ctx, "Unknown openflow group type", log.Fields{"type": t})
	return 0
}

func volthaGroupTypeToOpenflow(ctx context.Context, t openflow_13.OfpGroupType) ofp.GroupType {
	switch t {
	case openflow_13.OfpGroupType_OFPGT_ALL:
		return ofp.OFPGTAll
	case openflow_13.OfpGroupType_OFPGT_SELECT:
		return ofp.OFPGTSelect
	case openflow_13.OfpGroupType_OFPGT_INDIRECT:
		return ofp.OFPGTIndirect
	case openflow_13.OfpGroupType_OFPGT_FF:
		return ofp.OFPGTFf
	}
	logger.Errorw(ctx, "Unknown voltha group type", log.Fields{"type": t})
	return 0
}

func openflowBucketsToVoltha(buckets []*ofp.Bucket) []*openflow_13.OfpBucket {
	outBuckets := make([]*openflow_13.OfpBucket, len(buckets))

	for i, bucket := range buckets {
		b := &openflow_13.OfpBucket{
			Weight:     uint32(bucket.Weight),
			WatchPort:  uint32(bucket.WatchPort),
			WatchGroup: bucket.WatchGroup,
			Actions:    openflowActionsToVoltha(bucket.Actions),
		}
		outBuckets[i] = b
	}

	return outBuckets
}

func openflowActionsToVoltha(actions []goloxi.IAction) []*openflow_13.OfpAction {
	outActions := make([]*openflow_13.OfpAction, len(actions))

	for i, action := range actions {
		outActions[i] = extractAction(action)
	}

	return outActions
}

func volthaBucketsToOpenflow(ctx context.Context, buckets []*openflow_13.OfpBucket) []*ofp.Bucket {
	outBuckets := make([]*ofp.Bucket, len(buckets))

	for i, bucket := range buckets {
		actions := volthaActionsToOpenflow(ctx, bucket.Actions)
		b := &ofp.Bucket{
			Weight:     uint16(bucket.Weight),
			WatchPort:  ofp.Port(bucket.WatchPort),
			WatchGroup: bucket.WatchGroup,
			Actions:    actions,
		}
		outBuckets[i] = b
	}

	return outBuckets
}

func volthaActionsToOpenflow(ctx context.Context, actions []*openflow_13.OfpAction) []goloxi.IAction {
	outActions := make([]goloxi.IAction, len(actions))

	for i, action := range actions {
		outActions[i] = parseAction(ctx, action)
	}

	return outActions
}
