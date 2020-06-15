/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package openflow

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/opencord/goloxi"
	"github.com/opencord/goloxi/of13"
	"github.com/opencord/ofagent-go/internal/pkg/holder"
	"github.com/opencord/ofagent-go/internal/pkg/mock"
	"github.com/opencord/voltha-protos/v3/go/openflow_13"
	"github.com/opencord/voltha-protos/v3/go/voltha"
	"github.com/stretchr/testify/assert"
)

var msgSizeLimit = 64000

func createEapolFlow(id int) openflow_13.OfpFlowStats {

	portField := openflow_13.OfpOxmField{
		OxmClass: openflow_13.OfpOxmClass_OFPXMC_OPENFLOW_BASIC,
		Field: &openflow_13.OfpOxmField_OfbField{
			OfbField: &openflow_13.OfpOxmOfbField{
				Type:  openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_IN_PORT,
				Value: &openflow_13.OfpOxmOfbField_Port{Port: 16},
			},
		},
	}

	ethField := openflow_13.OfpOxmField{
		OxmClass: openflow_13.OfpOxmClass_OFPXMC_OPENFLOW_BASIC,
		Field: &openflow_13.OfpOxmField_OfbField{
			OfbField: &openflow_13.OfpOxmOfbField{
				Type:  openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_ETH_TYPE,
				Value: &openflow_13.OfpOxmOfbField_EthType{EthType: 2048},
			},
		},
	}

	vlanField := openflow_13.OfpOxmField{
		OxmClass: openflow_13.OfpOxmClass_OFPXMC_OPENFLOW_BASIC,
		Field: &openflow_13.OfpOxmField_OfbField{
			OfbField: &openflow_13.OfpOxmOfbField{
				Type:  openflow_13.OxmOfbFieldTypes_OFPXMT_OFB_VLAN_VID,
				Value: &openflow_13.OfpOxmOfbField_VlanVid{VlanVid: 4096},
			},
		},
	}

	return openflow_13.OfpFlowStats{
		Id:       uint64(id),
		Priority: 10000,
		Flags:    1,
		Match: &openflow_13.OfpMatch{
			Type: openflow_13.OfpMatchType_OFPMT_OXM,
			OxmFields: []*openflow_13.OfpOxmField{
				&portField, &ethField, &vlanField,
			},
		},
		Instructions: []*openflow_13.OfpInstruction{
			{
				Type: uint32(openflow_13.OfpInstructionType_OFPIT_APPLY_ACTIONS),
				Data: &openflow_13.OfpInstruction_Actions{
					Actions: &openflow_13.OfpInstructionActions{
						Actions: []*openflow_13.OfpAction{
							{
								Type: openflow_13.OfpActionType_OFPAT_OUTPUT,
								Action: &openflow_13.OfpAction_Output{
									Output: &openflow_13.OfpActionOutput{
										Port: 4294967293,
									},
								},
							},
						},
					},
				},
			},
			{
				Type: uint32(openflow_13.OfpInstructionType_OFPIT_WRITE_METADATA),
				Data: &openflow_13.OfpInstruction_WriteMetadata{
					WriteMetadata: &openflow_13.OfpInstructionWriteMetadata{
						Metadata: 274877906944,
					},
				},
			},
			{
				Type: uint32(openflow_13.OfpInstructionType_OFPIT_METER),
				Data: &openflow_13.OfpInstruction_Meter{
					Meter: &openflow_13.OfpInstructionMeter{
						MeterId: 2,
					},
				},
			},
		},
	}
}

func createRandomFlows(count int) []*openflow_13.OfpFlowStats {

	var flows []*openflow_13.OfpFlowStats

	n := 0
	for n < count {

		flow := createEapolFlow(n)

		flows = append(flows, &flow)

		n = n + 1
	}
	return flows
}

func createRandomPorts(count int) []*voltha.LogicalPort {
	var ports []*voltha.LogicalPort

	n := 0
	for n < count {

		port := voltha.LogicalPort{
			Id: fmt.Sprintf("uni-%d", n),
			OfpPort: &openflow_13.OfpPort{
				PortNo:     uint32(n),
				HwAddr:     []uint32{8, 0, 0, 0, 0, uint32(n)},
				Name:       fmt.Sprintf("BBSM-%d", n),
				State:      uint32(openflow_13.OfpPortState_OFPPS_LIVE),
				Curr:       4128,
				Advertised: 4128,
				Peer:       4128,
				CurrSpeed:  32,
				MaxSpeed:   32,
				Config:     1,
				Supported:  1,
			},
		}

		ports = append(ports, &port)

		n = n + 1
	}
	return ports
}

func newTestOFConnection(flowsCount int, portsCount int) *OFConnection {

	flows := openflow_13.Flows{
		Items: createRandomFlows(flowsCount),
	}

	ports := voltha.LogicalPorts{
		Items: createRandomPorts(portsCount),
	}

	volthaClient := mock.MockVolthaClient{
		LogicalDeviceFlows: flows,
		LogicalPorts:       ports,
	}

	volthaClientHolder := &holder.VolthaServiceClientHolder{}
	volthaClientHolder.Set(volthaClient)

	return &OFConnection{
		VolthaClient:       volthaClientHolder,
		flowsChunkSize:     ofcFlowsChunkSize,
		portsChunkSize:     ofcPortsChunkSize,
		portsDescChunkSize: ofcPortsDescChunkSize,
	}
}

func TestHandleFlowStatsRequest(t *testing.T) {

	generatedFlowsCount := 2000

	ofc := newTestOFConnection(2000, 0)

	request := of13.NewFlowStatsRequest()

	replies, err := ofc.handleFlowStatsRequest(context.Background(), request)
	assert.Equal(t, err, nil)

	// check that the correct number of messages is generated
	assert.Equal(t, int(math.Ceil(float64(generatedFlowsCount)/ofcFlowsChunkSize)), len(replies))

	n := 1
	entriesCount := 0

	for _, r := range replies {
		json, _ := r.Flags.MarshalJSON()

		// check that the ReplyMore flag is correctly set in the messages
		if n == len(replies) {
			assert.Equal(t, string(json), "{}")
		} else {
			assert.Equal(t, string(json), "{\"OFPSFReplyMore\": true}")
		}

		// check the message size
		enc := goloxi.NewEncoder()
		if err := r.Serialize(enc); err != nil {
			t.Fail()
		}
		bytes := enc.Bytes()
		fmt.Println("FlowStats msg size: ", len(bytes))
		if len(bytes) > msgSizeLimit {
			t.Fatal("Message size is bigger than 64KB")
		}

		entriesCount = entriesCount + len(r.GetEntries())
		n++
	}

	// make sure all the generate item are included in the responses
	assert.Equal(t, generatedFlowsCount, entriesCount)
}

func TestHandlePortStatsRequest(t *testing.T) {

	generatedPortsCount := 2560

	ofc := newTestOFConnection(0, generatedPortsCount)

	request := of13.NewPortStatsRequest()

	// request stats for all ports
	request.PortNo = 0xffffffff

	replies, err := ofc.handlePortStatsRequest(request)
	assert.Equal(t, err, nil)

	assert.Equal(t, int(math.Ceil(float64(generatedPortsCount)/ofcPortsChunkSize)), len(replies))

	n := 1
	entriesCount := 0

	for _, r := range replies {
		json, _ := r.Flags.MarshalJSON()

		// check that the ReplyMore flag is correctly set in the messages
		if n == len(replies) {
			assert.Equal(t, string(json), "{}")
		} else {
			assert.Equal(t, string(json), "{\"OFPSFReplyMore\": true}")
		}

		// check the message size
		enc := goloxi.NewEncoder()
		if err := r.Serialize(enc); err != nil {
			t.Fail()
		}
		bytes := enc.Bytes()
		fmt.Println("PortStats msg size: ", len(bytes))
		if len(bytes) > msgSizeLimit {
			t.Fatal("Message size is bigger than 64KB")
		}

		entriesCount = entriesCount + len(r.GetEntries())
		n++
	}

	// make sure all the generate item are included in the responses
	assert.Equal(t, generatedPortsCount, entriesCount)
}

func TestHandlePortDescStatsRequest(t *testing.T) {

	generatedPortsCount := 2560

	ofc := newTestOFConnection(0, generatedPortsCount)

	request := of13.NewPortDescStatsRequest()

	replies, err := ofc.handlePortDescStatsRequest(request)
	assert.Equal(t, err, nil)

	// check that the correct number of messages is generated
	assert.Equal(t, int(math.Ceil(float64(generatedPortsCount)/ofcPortsDescChunkSize)), len(replies))

	n := 1
	entriesCount := 0

	for _, r := range replies {
		json, _ := r.Flags.MarshalJSON()

		// check that the ReplyMore flag is correctly set in the messages
		if n == len(replies) {
			assert.Equal(t, string(json), "{}")
		} else {
			assert.Equal(t, string(json), "{\"OFPSFReplyMore\": true}")
		}

		// check the message size
		enc := goloxi.NewEncoder()
		if err := r.Serialize(enc); err != nil {
			t.Fail()
		}
		bytes := enc.Bytes()
		fmt.Println("PortDesc msg size: ", len(bytes))
		if len(bytes) > msgSizeLimit {
			t.Fatal("Message size is bigger than 64KB")
		}

		entriesCount = entriesCount + len(r.GetEntries())
		n++
	}

	// make sure all the generate item are included in the responses
	assert.Equal(t, generatedPortsCount, entriesCount)
}
