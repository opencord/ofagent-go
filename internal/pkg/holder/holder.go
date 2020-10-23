/*
 *   Copyright 2020-present Open Networking Foundation
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package holder

import (
	"sync"

	"github.com/opencord/voltha-protos/v4/go/voltha"
)

// VolthaServiceClientHolder provides a consistent (voluntarily unmutable) reference
// point for a mutable value that represents a GRPC service interface to
// VOLTHA
type VolthaServiceClientHolder struct {
	mutex           sync.RWMutex
	volthaSvcClient voltha.VolthaServiceClient
}

type VolthaServiceClientReference struct {
}

// Clear sets the held value to nil (not set)
func (h *VolthaServiceClientHolder) Clear() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.volthaSvcClient = nil
}

// Set assigns the value being held to the specified value
func (h *VolthaServiceClientHolder) Set(client voltha.VolthaServiceClient) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.volthaSvcClient = client
}

// Get returns the currently held value
func (h *VolthaServiceClientHolder) Get() voltha.VolthaServiceClient {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.volthaSvcClient
}
