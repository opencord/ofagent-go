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
	"fmt"
	"net"
	"strings"
	"sync"

	ofp "github.com/opencord/goloxi/of13"
	"github.com/opencord/voltha-protos/v5/go/openflow_13"
)

var mu sync.Mutex
var xid uint32 = 1

func GetXid() uint32 {
	mu.Lock()
	defer mu.Unlock()
	xid++
	return xid
}
func PadString(value string, padSize int) string {
	size := len(value)
	nullsNeeded := padSize - size
	null := fmt.Sprintf("%c", '\000')
	padded := strings.Repeat(null, nullsNeeded)
	return fmt.Sprintf("%s%s", value, padded)
}

func extractAction(action ofp.IAction) *openflow_13.OfpAction {
	var ofpAction openflow_13.OfpAction
	switch action.GetType() {
	case ofp.OFPATOutput:
		var outputAction openflow_13.OfpAction_Output
		loxiOutputAction := action.(*ofp.ActionOutput)
		var output openflow_13.OfpActionOutput
		output.Port = uint32(loxiOutputAction.GetPort())
		/*
			var maxLen uint16
			maxLen = loxiOutputAction.GetMaxLen()
			output.MaxLen = uint32(maxLen)

		*/
		output.MaxLen = 0
		outputAction.Output = &output
		ofpAction.Action = &outputAction
		ofpAction.Type = openflow_13.OfpActionType_OFPAT_OUTPUT
	case ofp.OFPATCopyTtlOut: //CopyTtltOut
	case ofp.OFPATCopyTtlIn: //CopyTtlIn
	case ofp.OFPATSetMplsTtl: //SetMplsTtl
		mplsTtl := action.(*ofp.ActionSetMplsTtl)
		setMplsTtl := openflow_13.OfpAction_MplsTtl{
			MplsTtl: &openflow_13.OfpActionMplsTtl{
				MplsTtl: uint32(mplsTtl.MplsTtl),
			},
		}
		ofpAction.Type = openflow_13.OfpActionType_OFPAT_SET_MPLS_TTL
		ofpAction.Action = &setMplsTtl
	case ofp.OFPATDecMplsTtl: //DecMplsTtl
	case ofp.OFPATPushVLAN: //PushVlan
		var pushVlan openflow_13.OfpAction_Push
		loxiPushAction := action.(*ofp.ActionPushVlan)
		var push openflow_13.OfpActionPush
		push.Ethertype = uint32(loxiPushAction.Ethertype) //TODO This should be available in the fields
		pushVlan.Push = &push
		ofpAction.Type = openflow_13.OfpActionType_OFPAT_PUSH_VLAN
		ofpAction.Action = &pushVlan
	case ofp.OFPATPopVLAN: //PopVlan
		ofpAction.Type = openflow_13.OfpActionType_OFPAT_POP_VLAN
	case ofp.OFPATPushMpls: //PushMpls
		var pushMpls openflow_13.OfpAction_Push
		mplsPushAction := action.(*ofp.ActionPushMpls)
		var push openflow_13.OfpActionPush
		push.Ethertype = uint32(mplsPushAction.Ethertype)
		pushMpls.Push = &push
		ofpAction.Type = openflow_13.OfpActionType_OFPAT_PUSH_MPLS
		ofpAction.Action = &pushMpls
	case ofp.OFPATPopMpls: //PopMpls
		ofpAction.Type = openflow_13.OfpActionType_OFPAT_POP_MPLS
	case ofp.OFPATSetQueue: //SetQueue
	case ofp.OFPATGroup: //ActionGroup
		ofpAction.Type = openflow_13.OfpActionType_OFPAT_GROUP
		group := action.(*ofp.ActionGroup)
		ofpAction.Action = &openflow_13.OfpAction_Group{
			Group: &openflow_13.OfpActionGroup{
				GroupId: group.GroupId,
			},
		}
	case ofp.OFPATSetNwTtl: //SetNwTtl
	case ofp.OFPATDecNwTtl: //DecNwTtl
	case ofp.OFPATSetField: //SetField
		ofpAction.Type = openflow_13.OfpActionType_OFPAT_SET_FIELD
		var ofpAction_SetField openflow_13.OfpAction_SetField
		var ofpActionSetField openflow_13.OfpActionSetField
		var ofpOxmField openflow_13.OfpOxmField
		ofpOxmField.OxmClass = openflow_13.OfpOxmClass_OFPXMC_OPENFLOW_BASIC
		var ofpOxmField_OfbField openflow_13.OfpOxmField_OfbField
		var ofpOxmOfbField openflow_13.OfpOxmOfbField
		loxiSetField := action.(*ofp.ActionSetField)
		oxmName := loxiSetField.Field.GetOXMName()
		switch oxmName {
		//TODO handle set field sith other fields
		case "vlan_vid":
			ofpOxmOfbField.Type = openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID
			var vlanVid openflow_13.OfpOxmOfbField_VlanVid
			var VlanVid = loxiSetField.Field.GetOXMValue().(uint16)
			vlanVid.VlanVid = uint32(VlanVid)
			ofpOxmOfbField.Value = &vlanVid
		case "vlan_pcp":
			ofpOxmOfbField.Type = openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_PCP
			var vlanPcp openflow_13.OfpOxmOfbField_VlanPcp
			var VlanPcp = loxiSetField.Field.GetOXMValue().(uint8)
			vlanPcp.VlanPcp = uint32(VlanPcp)
			ofpOxmOfbField.Value = &vlanPcp
		case "mpls_label":
			ofpOxmOfbField.Type = openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_MPLS_LABEL
			var mplsLabel openflow_13.OfpOxmOfbField_MplsLabel
			label := loxiSetField.Field.GetOXMValue().(uint32)
			mplsLabel.MplsLabel = label
			ofpOxmOfbField.Value = &mplsLabel
		case "mpls_bos":
			ofpOxmOfbField.Type = openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_MPLS_BOS
			var mplsBos openflow_13.OfpOxmOfbField_MplsBos
			bos := loxiSetField.Field.GetOXMValue().(uint8)
			mplsBos.MplsBos = uint32(bos)
			ofpOxmOfbField.Value = &mplsBos
		case "eth_src":
			ofpOxmOfbField.Type = openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_ETH_SRC
			var ethSrc openflow_13.OfpOxmOfbField_EthSrc
			src := loxiSetField.Field.GetOXMValue().(net.HardwareAddr)
			ethSrc.EthSrc = src
			ofpOxmOfbField.Value = &ethSrc
		case "eth_dst":
			ofpOxmOfbField.Type = openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_ETH_DST
			var ethDst openflow_13.OfpOxmOfbField_EthDst
			dst := loxiSetField.Field.GetOXMValue().(net.HardwareAddr)
			ethDst.EthDst = dst
			ofpOxmOfbField.Value = &ethDst
		}
		ofpOxmField_OfbField.OfbField = &ofpOxmOfbField
		ofpOxmField.Field = &ofpOxmField_OfbField
		ofpActionSetField.Field = &ofpOxmField
		ofpAction_SetField.SetField = &ofpActionSetField
		ofpAction.Action = &ofpAction_SetField

	case ofp.OFPATPushPbb: //PushPbb
	case ofp.OFPATPopPbb: //PopPbb
	case ofp.OFPATExperimenter: //Experimenter

	}
	return &ofpAction

}
