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
	pb "github.com/opencord/voltha-protos/go/voltha"
	"log"
)

func handlePacketOut(packetOut *ofp.PacketOut, deviceId string) {
	jsonMessage, _ := json.Marshal(packetOut)
	log.Printf("handlePacketOut called with %s", jsonMessage)
	pktOut := pb.OfpPacketOut{}
	pktOut.BufferId = packetOut.GetBufferId()
	pktOut.InPort = uint32(packetOut.GetInPort())
	var actions []*pb.OfpAction
	inActions := packetOut.GetActions()
	for i := 0; i < len(inActions); i++ {
		action := inActions[i]
		newAction := extractAction(action)
		/*var newAction = pb.OfpAction{}
		newAction.Type = pb.OfpActionType(action.GetType())
		action.

		*/
		actions = append(actions, newAction)
	}
	pktOut.Actions = actions
	pktOut.Data = packetOut.GetData()
	js, _ := json.Marshal(pktOut)
	log.Printf("PacketOut %s", js)
	pbPacketOut := pb.PacketOut{}
	pbPacketOut.PacketOut = &pktOut
	pbPacketOut.Id = deviceId
	packetOutChannel <- pbPacketOut
}
