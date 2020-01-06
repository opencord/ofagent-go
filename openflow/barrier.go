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

	"github.com/opencord/ofagent-go/settings"

	ofp "github.com/donNewtonAlpha/goloxi/of13"
	l "github.com/opencord/voltha-lib-go/v2/pkg/log"
)

func handleBarrierRequest(request *ofp.BarrierRequest, DeviceID string, client *Client) {

	if settings.GetDebug(DeviceID) {
		js, _ := json.Marshal(request)
		logger.Debugw("handleBarrierRequest called with %s", l.Fields{"DeviceID": DeviceID, "request": js})
	}
	reply := ofp.NewBarrierReply()
	reply.SetVersion(4)
	reply.SetXid(request.GetXid())
	client.SendMessage(reply)
}
