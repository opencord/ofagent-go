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
	"github.com/opencord/goloxi"
	ofp "github.com/opencord/goloxi/of13"
	"github.com/opencord/voltha-lib-go/v4/pkg/log"
)

func (ofc *OFConnection) handleRoleRequest(ctx context.Context, request *ofp.RoleRequest) {
	span, ctx := log.CreateChildSpan(ctx, "openflow-role")
	defer span.Finish()

	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(request)
		logger.Debugw(ctx, "handleRoleRequest called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"request":   js})
	}

	if request.Role == ofp.OFPCRRoleNochange {
		reply := ofp.NewRoleReply()
		reply.SetXid(request.GetXid())
		reply.SetVersion(request.GetVersion())
		reply.SetRole(ofp.ControllerRole(ofc.role))
		reply.SetGenerationId(request.GetGenerationId())
		if err := ofc.SendMessage(ctx, reply); err != nil {
			logger.Errorw(ctx, "handle-role-request-send-message", log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
		}
	}

	ok := ofc.roleManager.UpdateRoles(ctx, ofc.OFControllerEndPoint, request)

	if ok {
		reply := ofp.NewRoleReply()
		reply.SetXid(request.GetXid())
		reply.SetVersion(request.GetVersion())
		reply.SetRole(request.GetRole())
		reply.SetGenerationId(request.GetGenerationId())
		if err := ofc.SendMessage(ctx, reply); err != nil {
			logger.Errorw(ctx, "handle-role-request-send-message", log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
		}
	} else {
		reply := ofp.NewRoleRequestFailedErrorMsg()
		reply.SetXid(request.GetXid())
		reply.SetVersion(request.GetVersion())
		reply.Code = ofp.OFPRRFCStale

		enc := goloxi.NewEncoder()
		err := request.Serialize(enc)
		if err == nil {
			reply.Data = enc.Bytes()
		}

		if err := ofc.SendMessage(ctx, reply); err != nil {
			logger.Errorw(ctx, "handle-role-request-send-message", log.Fields{
				"device-id": ofc.DeviceID,
				"error":     err})
		}
	}
}

func (ofc *OFConnection) sendRoleSlaveError(ctx context.Context, request ofp.IHeader) {
	reply := ofp.NewBadRequestErrorMsg()
	reply.SetXid(request.GetXid())
	reply.SetVersion(request.GetVersion())
	reply.Code = ofp.OFPCRRoleSlave

	enc := goloxi.NewEncoder()
	err := request.Serialize(enc)
	if err == nil {
		reply.Data = enc.Bytes()
	}

	if err := ofc.SendMessage(ctx, reply); err != nil {
		logger.Errorw(ctx, "send-role-slave-error", log.Fields{
			"device-id": ofc.DeviceID,
			"error":     err})
	}
}
