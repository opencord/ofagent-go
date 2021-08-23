/*
   Copyright 2021 the original author or authors.

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
	"fmt"
	"github.com/opencord/ofagent-go/internal/pkg/holder"
	"github.com/opencord/ofagent-go/internal/pkg/mock"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"

	"github.com/opencord/goloxi"
	"github.com/opencord/goloxi/of13"
	ofp "github.com/opencord/goloxi/of13"
)

func TestOFConnection_handleFlowAdd(t *testing.T) {
	ofc := &OFConnection{
		VolthaClient: &holder.VolthaServiceClientHolder{},
	}
	ofc.VolthaClient.Set(mock.MockVolthaClient{})

	callables := []func(t *testing.T) *ofp.FlowAdd{getUpstreamOnuFlow, getUpstreamOLTFlow, getOnuDownstreamFlow,
		getOLTDownstreamSingleMplsTagFlow, getOLTDownstreamDoubleMplsTagFlow}
	for _, callable := range callables {
		ofc.handleFlowAdd(context.Background(), callable(t))
	}
}

func getUpstreamOnuFlow(t *testing.T) *ofp.FlowAdd {
	flowAdd := &ofp.FlowAdd{
		FlowMod: &ofp.FlowMod{
			Header: &ofp.Header{
				Version: 0,
				Type:    ofp.OFPFCAdd,
				Length:  0,
				Xid:     10,
			},
			Cookie:   114560316889735549,
			TableId:  0,
			Priority: 1000,
			Match: ofp.Match{
				OxmList: []goloxi.IOxm{&of13.NxmInPort{
					Oxm: &of13.Oxm{
						TypeLen: 0,
					},
					Value: 16,
				},
					&of13.OxmVlanVid{
						Oxm: &of13.Oxm{
							TypeLen: 0,
						},
						Value: 4096,
					}},
			},
			Instructions: []ofp.IInstruction{
				&ofp.InstructionGotoTable{
					Instruction: &ofp.Instruction{
						Type: ofp.OFPITGotoTable,
						Len:  8,
					},
					TableId: 1,
				},
				&ofp.InstructionMeter{
					Instruction: &ofp.Instruction{
						Type: ofp.OFPITMeter,
					},
					MeterId: 1,
				},
			},
		},
	}
	flowAddJson, err := json.Marshal(flowAdd)
	assert.NoError(t, err)

	fmt.Printf("Onu Flow Add Json: %s\n", flowAddJson)
	return flowAdd
}

func getUpstreamOLTFlow(t *testing.T) *ofp.FlowAdd {
	flowAdd := &ofp.FlowAdd{
		FlowMod: &ofp.FlowMod{
			Header: &ofp.Header{
				Version: 0,
				Type:    ofp.OFPFCAdd,
				Length:  0,
				Xid:     10,
			},
			Cookie:   114560316889735549,
			TableId:  1,
			Priority: 1000,
			Match: ofp.Match{
				OxmList: []goloxi.IOxm{
					&of13.OxmInPort{
						Value: ofp.Port(16),
					},
					&of13.OxmVlanVid{
						Oxm: &of13.Oxm{
							TypeLen: 0,
						},
						Value: 4096,
					}},
			},
			Instructions: []ofp.IInstruction{
				&ofp.InstructionMeter{
					Instruction: &ofp.Instruction{
						Type: ofp.OFPITMeter,
					},
					MeterId: 1,
				},
			},
		},
	}

	actions := &ofp.InstructionApplyActions{
		Instruction: &ofp.Instruction{
			Type: ofp.OFPITApplyActions,
		},
	}
	actions.Actions = append(actions.Actions, &ofp.ActionPushVlan{
		Action: &ofp.Action{
			Type: ofp.OFPATPushVLAN,
		},
		Ethertype: 0x8100,
	},
		&ofp.ActionPushMpls{
			Action: &ofp.Action{
				Type: ofp.OFPATPushMpls,
			},
			Ethertype: 0x8847,
		},
		&ofp.ActionSetField{
			Action: &ofp.Action{
				Type: ofp.OFPATSetField,
			},
			Field: &ofp.OxmMplsLabel{
				Value: 10,
			},
		},
		&ofp.ActionSetField{
			Action: &ofp.Action{
				Type: ofp.OFPATSetField,
			},
			Field: &ofp.OxmMplsBos{
				Value: 1,
			},
		},
		&ofp.ActionSetField{
			Action: &ofp.Action{
				Type: ofp.OFPATSetField,
			},
			Field: &ofp.OxmEthSrc{
				Value: net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee},
			},
		},
		&ofp.ActionSetField{
			Action: &ofp.Action{
				Type: ofp.OFPATSetField,
			},
			Field: &ofp.OxmEthDst{
				Value: net.HardwareAddr{0xee, 0xdd, 0xcc, 0xbb, 0xaa},
			},
		},
		&ofp.ActionSetMplsTtl{
			Action: &ofp.Action{
				Type: ofp.OFPATSetMplsTtl,
			},
			MplsTtl: 64,
		},
	)

	flowAdd.Instructions = append(flowAdd.Instructions, actions)
	flowAddJson, err := json.Marshal(flowAdd)
	assert.NoError(t, err)

	fmt.Printf("OLT Flow Add Json: %s\n", flowAddJson)
	return flowAdd
}

func getOnuDownstreamFlow(t *testing.T) *ofp.FlowAdd {
	flowAdd := &ofp.FlowAdd{
		FlowMod: &ofp.FlowMod{
			Header: &ofp.Header{
				Version: 3,
				Type:    0,
				Xid:     0,
			},
			Cookie:   114560316889735549,
			TableId:  2,
			Priority: 1000,
			Flags:    0,
			Match: ofp.Match{
				Type:   0,
				Length: 0,
				OxmList: []goloxi.IOxm{
					&of13.OxmInPort{
						Value: ofp.Port(65536),
					},
					&of13.OxmVlanVid{
						Value: 4096,
					}},
			},
			Instructions: []ofp.IInstruction{
				&ofp.InstructionMeter{
					Instruction: &ofp.Instruction{
						Type: ofp.OFPITMeter,
					},
					MeterId: 1,
				},
			},
		},
	}

	actions := &ofp.InstructionApplyActions{
		Instruction: &ofp.Instruction{
			Type: ofp.OFPITApplyActions,
		},
	}
	actions.Actions = append(actions.Actions,
		&ofp.ActionOutput{
			Action: &ofp.Action{
				Type: ofp.OFPATOutput,
			},
			Port: 16,
		})

	flowAdd.Instructions = append(flowAdd.Instructions, actions)
	flowAddJson, err := json.Marshal(flowAdd)
	assert.NoError(t, err)

	fmt.Printf("Onu Downstream Flow Add Json: %s\n", flowAddJson)
	return flowAdd

}

func getOLTDownstreamSingleMplsTagFlow(t *testing.T) *ofp.FlowAdd {
	flowAdd := &ofp.FlowAdd{
		FlowMod: &ofp.FlowMod{
			Header: &ofp.Header{
				Version: 3,
				Type:    0,
				Xid:     0,
			},
			Cookie:   114560316889735549,
			TableId:  0,
			Priority: 1000,
			Flags:    0,
			Match: ofp.Match{
				OxmList: []goloxi.IOxm{
					&of13.OxmInPort{
						Value: ofp.Port(65536),
					},
					&of13.OxmEthType{
						Value: ofp.EthernetType(0x8847),
					},
					&of13.OxmMplsBos{
						Value: 1,
					},
					&of13.OxmVlanVid{
						Oxm: &of13.Oxm{
							TypeLen: 0,
						},
						Value: 4096,
					},
					&of13.OxmEthSrc{
						Value: net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd},
					},
				},
			},
			Instructions: []ofp.IInstruction{
				&ofp.InstructionGotoTable{
					Instruction: &ofp.Instruction{
						Type: ofp.OFPITGotoTable,
					},
					TableId: 1,
				},
				&ofp.InstructionMeter{
					Instruction: &ofp.Instruction{
						Type: ofp.OFPITMeter,
					},
					MeterId: 1,
				},
			},
		},
	}

	actions := &ofp.InstructionApplyActions{
		Instruction: &ofp.Instruction{
			Type: ofp.OFPITApplyActions,
		},
	}
	actions.Actions = append(actions.Actions,
		&ofp.ActionOutput{
			Action: &ofp.Action{
				Type: ofp.OFPATOutput,
			},
			Port: 16,
		},
		&ofp.ActionDecMplsTtl{
			Action: &ofp.Action{
				Type: ofp.OFPATDecMplsTtl,
			},
		},
		&ofp.ActionSetMplsTtl{
			Action: &ofp.Action{
				Type: ofp.OFPATSetMplsTtl,
			},
			MplsTtl: 64,
		},
		&ofp.ActionPopMpls{
			Action: &ofp.Action{
				Type: ofp.OFPATPopMpls,
			},
			Ethertype: 0x8847,
		},
	)

	flowAdd.Instructions = append(flowAdd.Instructions, actions)
	flowAddJson, err := json.Marshal(flowAdd)
	assert.NoError(t, err)

	fmt.Printf("Olt Downstream (Single MPLS tag) Flow Add Json: %s\n", flowAddJson)
	return flowAdd
}

func getOLTDownstreamDoubleMplsTagFlow(t *testing.T) *ofp.FlowAdd {
	flowAdd := &ofp.FlowAdd{
		FlowMod: &ofp.FlowMod{
			Header: &ofp.Header{
				Version: 3,
				Type:    0,
				Xid:     0,
			},
			Cookie:   114560316889735549,
			TableId:  0,
			Priority: 1000,
			Flags:    0,
			Match: ofp.Match{
				OxmList: []goloxi.IOxm{
					&of13.OxmInPort{
						Value: ofp.Port(65536),
					},
					&of13.OxmEthType{
						Oxm: &of13.Oxm{
							TypeLen: 0,
						},
						Value: ofp.EthernetType(0x8847),
					},
					&of13.OxmMplsBos{
						Value: 1,
					},
					&of13.OxmVlanVid{
						Value: 4096,
					},
					&of13.OxmEthSrc{
						Value: net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd},
					},
				},
			},
			Instructions: []ofp.IInstruction{
				&ofp.InstructionGotoTable{
					Instruction: &ofp.Instruction{
						Type: ofp.OFPITGotoTable,
					},
					TableId: 1,
				},
				&ofp.InstructionMeter{
					Instruction: &ofp.Instruction{
						Type: ofp.OFPITMeter,
					},
					MeterId: 1,
				},
			},
		},
	}

	actions := &ofp.InstructionApplyActions{
		Instruction: &ofp.Instruction{
			Type: ofp.OFPITApplyActions,
		},
	}
	actions.Actions = append(actions.Actions,
		&ofp.ActionOutput{
			Action: &ofp.Action{
				Type: ofp.OFPATOutput,
			},
			Port: 16,
		},
		&ofp.ActionDecMplsTtl{
			Action: &ofp.Action{
				Type: ofp.OFPATDecMplsTtl,
			},
		},
		&ofp.ActionSetMplsTtl{
			Action: &ofp.Action{
				Type: ofp.OFPATSetMplsTtl,
			},
			MplsTtl: 64,
		},
		&ofp.ActionPopMpls{
			Action: &ofp.Action{
				Type: ofp.OFPATPopMpls,
			},
			Ethertype: 0x8847,
		},
		&ofp.ActionPopMpls{
			Action: &ofp.Action{
				Type: ofp.OFPATPopMpls,
			},
			Ethertype: 0x8847,
		},
	)

	flowAdd.Instructions = append(flowAdd.Instructions, actions)
	flowAddJson, err := json.Marshal(flowAdd)
	assert.NoError(t, err)
	fmt.Printf("Olt Downstream (Double MPLS tag) Flow Add Json: %s\n", flowAddJson)
	return flowAdd
}
