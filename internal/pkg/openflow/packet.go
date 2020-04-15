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
	"encoding/json"
	ofp "github.com/opencord/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-protos/v3/go/voltha"
)

func (ofc *OFConnection) handlePacketOut(packetOut *ofp.PacketOut) {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(packetOut)
		logger.Debugw("handlePacketOut called",
			log.Fields{
				"device-id":  ofc.DeviceID,
				"packet-out": js})
	}

	// Collection actions
	var actions []*voltha.OfpAction
	for _, action := range packetOut.GetActions() {
		actions = append(actions, extractAction(action))
	}

	// Build packet out
	pbPacketOut := voltha.PacketOut{
		Id: ofc.DeviceID,
		PacketOut: &voltha.OfpPacketOut{
			BufferId: packetOut.GetBufferId(),
			InPort:   uint32(packetOut.GetInPort()),
			Actions:  actions,
			Data:     packetOut.GetData(),
		},
	}

	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(pbPacketOut)
		logger.Debugw("handlePacketOut sending",
			log.Fields{
				"device-id":  ofc.DeviceID,
				"packet-out": js})
	}

	// Queue it
	ofc.PacketOutChannel <- &pbPacketOut
}
