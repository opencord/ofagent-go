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
	ofp "github.com/donNewtonAlpha/goloxi/of13"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/ofagent-go/openflow"
	pb "github.com/opencord/voltha-protos/v2/go/voltha"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

func receiveChangeEvent(client pb.VolthaServiceClient) {
	opt := grpc.EmptyCallOption{}
	stream, err := client.ReceiveChangeEvents(context.Background(), &empty.Empty{}, opt)
	if err != nil {
		log.Fatalln("Unable to establish Receive Change Event Stream")
	}
	for {
		changeEvent, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving change event %v", err)
		}
		js, _ := json.Marshal(changeEvent)

		log.Printf("Received Change Event   \n\n\n %s \n\n\n", js)
		deviceId := changeEvent.GetId()
		portStatus := changeEvent.GetPortStatus()
		if portStatus == nil {
			jsonMessage, _ := json.Marshal(changeEvent.GetEvent())
			log.Printf("Received change event that was not port status %v", jsonMessage)
			break
		}
		ofPortStatus := ofp.NewPortStatus()
		ofPortStatus.SetXid(openflow.GetXid())
		ofPortStatus.SetVersion(4)

		ofReason := ofp.PortReason(portStatus.GetReason())
		ofPortStatus.SetReason(ofReason)
		ofDesc := ofp.NewPortDesc()

		desc := portStatus.GetDesc()
		ofDesc.SetAdvertised(ofp.PortFeatures(desc.GetAdvertised()))
		ofDesc.SetConfig(ofp.PortConfig(0))
		ofDesc.SetCurr(ofp.PortFeatures(desc.GetAdvertised()))
		ofDesc.SetCurrSpeed(desc.GetCurrSpeed())
		intArray := desc.GetHwAddr()
		var octets []byte
		for i := 0; i < len(intArray); i++ {
			octets = append(octets, byte(intArray[i]))
		}
		addr := net.HardwareAddr(octets)
		ofDesc.SetHwAddr(addr)
		ofDesc.SetMaxSpeed(desc.GetMaxSpeed())
		ofDesc.SetName(openflow.PadString(desc.GetName(), 16))
		ofDesc.SetPeer(ofp.PortFeatures(desc.GetPeer()))
		ofDesc.SetPortNo(ofp.Port(desc.GetPortNo()))
		ofDesc.SetState(ofp.PortState(desc.GetState()))
		ofDesc.SetSupported(ofp.PortFeatures(desc.GetSupported()))
		ofPortStatus.SetDesc(*ofDesc)
		var client = clientMap[deviceId]
		if client == nil {
			client = addClient(deviceId)
			time.Sleep(2 * time.Second)
		}
		client.SendMessage(ofPortStatus)
	}
}
