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

package grpc

import (
	"context"
	"encoding/json"

	"github.com/opencord/ofagent-go/openflow"
	"github.com/opencord/ofagent-go/settings"
	l "github.com/opencord/voltha-lib-go/v2/pkg/log"
	pb "github.com/opencord/voltha-protos/v2/go/voltha"
	"google.golang.org/grpc"
)

func streamPacketOut(client pb.VolthaServiceClient) {
	if settings.GetDebug(grpcDeviceID) {
		logger.Debugln("GrpcClient streamPacketOut called")
	}
	opt := grpc.EmptyCallOption{}
	outClient, err := client.StreamPacketsOut(context.Background(), opt)
	if err != nil {
		logger.Fatalw("streamPacketOut Error creating packetout stream ", l.Fields{"Error": err})
	}
	packetOutChannel := make(chan pb.PacketOut)
	openflow.SetPacketOutChannel(packetOutChannel)
	for {
		ofPacketOut := <-packetOutChannel
		if settings.GetDebug(grpcDeviceID) {
			js, _ := json.Marshal(ofPacketOut)
			logger.Debugw("streamPacketOut Receive PacketOut from Channel", l.Fields{"PacketOut": js})
		}
		outClient.Send(&ofPacketOut)
	}

}
