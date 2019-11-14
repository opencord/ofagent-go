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
	"log"
)

func handleEchoRequest(request *ofp.EchoRequest, client *Client) {
	jsonMessage, _ := json.Marshal(request)
	log.Printf("handleEchoRequest called with %s", jsonMessage)
	reply := ofp.NewEchoReply()
	reply.SetXid(request.GetXid())
	reply.SetVersion(request.GetVersion())
	client.SendMessage(reply)

}
