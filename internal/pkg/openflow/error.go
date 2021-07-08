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
	"github.com/opencord/voltha-lib-go/v5/pkg/log"
)

func (ofc *OFConnection) handleErrMsg(ctx context.Context, message ofp.IErrorMsg) {
	if logger.V(log.DebugLevel) {
		js, _ := json.Marshal(message)
		logger.Debugw(ctx, "handleErrMsg called",
			log.Fields{
				"device-id": ofc.DeviceID,
				"request":   js})
	}
	//Not sure yet what if anything to do here
}
