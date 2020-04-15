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
	ofp "github.com/opencord/goloxi/of13"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testRoleManager struct {
	role         ofp.ControllerRole
	generationId uint64
	from         string

	generateError bool
}

func (trm *testRoleManager) UpdateRoles(from string, request *ofp.RoleRequest) bool {
	trm.from = from
	trm.role = request.Role
	trm.generationId = request.GenerationId

	return !trm.generateError
}

func createOFClient() *OFClient {
	endpoints := []string{"e1", "e2", "e3"}

	// create a client object
	ofclient := &OFClient{
		OFControllerEndPoints: endpoints,
		connections:           make(map[string]*OFConnection),
	}

	// create a connection object for each endpoint
	for _, ep := range endpoints {
		conn := &OFConnection{OFControllerEndPoint: ep, role: ofcRoleNone}
		ofclient.connections[ep] = conn
	}

	return ofclient
}

func createRoleRequest(role ofp.ControllerRole, genId uint64) *ofp.RoleRequest {
	rr := ofp.NewRoleRequest()
	rr.Role = role
	rr.GenerationId = genId
	return rr
}

func TestRoleChange(t *testing.T) {
	endpoints := []string{"e1", "e2", "e3"}
	ofclient := &OFClient{
		OFControllerEndPoints: endpoints, connections: make(map[string]*OFConnection),
	}

	for _, ep := range endpoints {
		conn := &OFConnection{OFControllerEndPoint: ep, role: ofcRoleNone}
		ofclient.connections[ep] = conn
	}

	assert.Equal(t, ofclient.connections["e1"].role, ofcRoleNone)

	// change role of e1 to master
	rr := createRoleRequest(ofp.OFPCRRoleMaster, 1)

	ok := ofclient.UpdateRoles("e1", rr)
	assert.True(t, ok)
	assert.Equal(t, ofclient.connections["e1"].role, ofcRoleMaster)

	// change role of e2 to master
	ok = ofclient.UpdateRoles("e2", rr)
	assert.True(t, ok)
	assert.Equal(t, ofclient.connections["e2"].role, ofcRoleMaster)
	// e1 should now have reverted to slave
	assert.Equal(t, ofclient.connections["e1"].role, ofcRoleSlave)

	// change role of e2 to slave
	rr = createRoleRequest(ofp.OFPCRRoleSlave, 1)

	ok = ofclient.UpdateRoles("e2", rr)
	assert.True(t, ok)
	assert.Equal(t, ofclient.connections["e2"].role, ofcRoleSlave)
}

func TestStaleRoleRequest(t *testing.T) {
	ofclient := createOFClient()

	rr1 := createRoleRequest(ofp.OFPCRRoleMaster, 2)

	ok := ofclient.UpdateRoles("e1", rr1)
	assert.True(t, ok)
	assert.Equal(t, ofclient.connections["e1"].role, ofcRoleMaster)

	// 'stale' role request
	rr2 := createRoleRequest(ofp.OFPCRRoleSlave, 1)

	ok = ofclient.UpdateRoles("e1", rr2)
	// should not have succeeded
	assert.False(t, ok)
	// role should remain master
	assert.Equal(t, ofclient.connections["e1"].role, ofcRoleMaster)
}

func TestHandleRoleRequest(t *testing.T) {

	trm := &testRoleManager{}

	connection := &OFConnection{
		OFControllerEndPoint: "e1",
		roleManager:          trm,
		sendChannel:          make(chan Message, 10),
	}

	rr := createRoleRequest(ofp.OFPCRRoleMaster, 1)

	connection.handleRoleRequest(rr)

	assert.Equal(t, "e1", trm.from)
	assert.EqualValues(t, ofp.OFPCRRoleMaster, trm.role)
	assert.EqualValues(t, 1, trm.generationId)

	resp := (<-connection.sendChannel).(*ofp.RoleReply)

	assert.EqualValues(t, resp.Role, ofp.OFPCRRoleMaster)
}

func TestHandleRoleRequestError(t *testing.T) {

	trm := &testRoleManager{generateError: true}

	connection := &OFConnection{
		OFControllerEndPoint: "e1",
		roleManager:          trm,
		sendChannel:          make(chan Message, 10),
	}

	rr := createRoleRequest(ofp.OFPCRRoleMaster, 1)

	connection.handleRoleRequest(rr)

	resp := (<-connection.sendChannel).(*ofp.RoleRequestFailedErrorMsg)

	assert.EqualValues(t, resp.Code, ofp.OFPRRFCStale)
}
