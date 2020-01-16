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
	"encoding/json"
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-protos/v2/go/voltha"
)

func (ofc *OFClient) handlePacketOut(packetOut *ofp.PacketOut) {
	if log.V(log.DebugLevel) {
		js, _ := json.Marshal(packetOut)
		logger.Debugw("handlePacketOut called", log.Fields{"device-id": ofc.DeviceID, "packetOut": js})
	}
	pktOut := voltha.OfpPacketOut{}
	pktOut.BufferId = packetOut.GetBufferId()
	pktOut.InPort = uint32(packetOut.GetInPort())
	var actions []*voltha.OfpAction
	inActions := packetOut.GetActions()
	for i := 0; i < len(inActions); i++ {
		action := inActions[i]
		newAction := extractAction(action)
		actions = append(actions, newAction)
	}
	pktOut.Actions = actions
	pktOut.Data = packetOut.GetData()
	pbPacketOut := voltha.PacketOut{}
	pbPacketOut.PacketOut = &pktOut
	pbPacketOut.Id = ofc.DeviceID
	if log.V(log.DebugLevel) {
		js, _ := json.Marshal(pbPacketOut)
		logger.Debugw("handlePacketOut sending", log.Fields{"device-id": ofc.DeviceID, "packetOut": js})
	}
	packetOutChannel <- pbPacketOut
}
