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

// The implementation of the open flow agent (ofagent) uses a GRPC connection
// to VOLTHA accross several implementaton packages including the ofagent and
// openflow packages. This GRPC connection is shared through packages and type
// instances via injection.
//
// As the ofagent executes within a micro-service environment it is
// possible that the GRPC connection is reset (re-established). When the connection
// is re-established, because injection is used, the new value needs to be
// re-injected across the implementation.
//
// To help simply the re-injection or value change scenario a holder for the
// GRPC connection is established so that the reference to the holder can
// stay [sic] consistant over the lifetime of the ofagent while the underlying
// GRPC connection can change without walking the entire runtime structure.
package holder
