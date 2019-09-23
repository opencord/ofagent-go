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
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/ofagent-go/openflow"
	pb "github.com/opencord/voltha-protos/go/voltha"
	ofp "github.com/skydive-project/goloxi/of13"
	"google.golang.org/grpc"
	"log"
)

func receivePacketIn(client pb.VolthaServiceClient) {
	opt := grpc.EmptyCallOption{}
	stream, err := client.ReceivePacketsIn(context.Background(), &empty.Empty{}, opt)
	if err != nil {
		log.Fatalln("Unable to establish stream")
	}
	for {
		packet, err := stream.Recv()
		packetIn := packet.GetPacketIn()

		if err != nil {
			log.Fatalf("error on stream.Rec %v", err)
		}
		ofPacketIn := ofp.NewPacketIn()
		ofPacketIn.SetVersion(uint8(4))
		ofPacketIn.SetXid(openflow.GetXid())
		ofPacketIn.SetBufferId(packetIn.GetBufferId())
		ofPacketIn.SetCookie(packetIn.GetCookie())
		ofPacketIn.SetData(packetIn.GetData())
		var outMatch ofp.Match
		inMatch := packetIn.GetMatch()
		outMatch.SetType(uint16(inMatch.GetType()))
		/*
			TODO not sure if anything further is needed
			fields := inMatch.GetOxmFields()
			var outFields []ofp.Oxm
			for i:=0;i< len(fields);i++{
				field := fields[i]
				outField := ofp.Oxm{}
				outField.SetTypeLen(field.OxmClass.)
				outField.SetTypeLen(field)
			}
			outMatch.SetOxmList(inMatch.GetOxmFields())
		*/

		ofPacketIn.SetMatch(outMatch)
		ofPacketIn.SetReason(uint8(packetIn.GetReason()))
		ofPacketIn.SetTableId(uint8(packetIn.GetTableId()))
		ofPacketIn.SetTotalLen(uint16(len(ofPacketIn.GetData())))
		//ofPacketIn.
	}

}
