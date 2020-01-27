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

package ofagent

import (
	"context"
	"encoding/json"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"google.golang.org/grpc"
)

func (ofa *OFAgent) streamPacketOut(ctx context.Context) {
	if logger.V(log.DebugLevel) {
		logger.Debug("GrpcClient streamPacketOut called")
	}
	opt := grpc.EmptyCallOption{}
	streamCtx, streamDone := context.WithCancel(context.Background())
	outClient, err := ofa.volthaClient.StreamPacketsOut(streamCtx, opt)
	defer streamDone()
	if err != nil {
		logger.Errorw("streamPacketOut Error creating packetout stream ", log.Fields{"error": err})
		ofa.events <- ofaEventVolthaDisconnected
	}
	for {
		select {
		case <-ctx.Done():
			return
		case ofPacketOut := <-ofa.packetOutChannel:
			if logger.V(log.DebugLevel) {
				js, _ := json.Marshal(ofPacketOut)
				logger.Debugw("streamPacketOut Receive PacketOut from Channel", log.Fields{"PacketOut": js})
			}
			outClient.Send(ofPacketOut)
		}
	}
}
