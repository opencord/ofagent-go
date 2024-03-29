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

	ofp "github.com/opencord/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
)

func (ofc *OFConnection) handleEchoRequest(ctx context.Context, request *ofp.EchoRequest) {
	span, ctx := log.CreateChildSpan(ctx, "openflow-echo")
	defer span.Finish()

	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(request)
		logger.Debugw(ctx, "handleEchoRequest called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"request":   js})
	}
	reply := ofp.NewEchoReply()
	reply.SetXid(request.GetXid())
	reply.SetVersion(request.GetVersion())
	if err := ofc.SendMessage(ctx, reply); err != nil {
		logger.Errorw(ctx, "handle-echo-request-send-message", log.Fields{"error": err})
	}
}
