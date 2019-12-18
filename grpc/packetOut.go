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
	"encoding/json"
	"github.com/opencord/ofagent-go/openflow"
	pb "github.com/opencord/voltha-protos/v2/go/voltha"
	"google.golang.org/grpc"
	"log"
)
import "context"

func streamPacketOut(client pb.VolthaServiceClient) {
	opt := grpc.EmptyCallOption{}
	outClient, err := client.StreamPacketsOut(context.Background(), opt)
	if err != nil {

		log.Printf("Error creating packetout stream %v", err)
	}
	packetOutChannel := make(chan pb.PacketOut)
	openflow.SetPacketOutChannel(packetOutChannel)
	for {
		ofPacketOut := <-packetOutChannel
		js, _ := json.Marshal(ofPacketOut)
		log.Printf("RECEIVED PACKET OUT FROM CHANNEL %s", js)
		outClient.Send(&ofPacketOut)
	}

}
