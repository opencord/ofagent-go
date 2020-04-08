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
	"errors"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-lib-go/v3/pkg/probe"
	"github.com/opencord/voltha-protos/v3/go/voltha"
	"google.golang.org/grpc"
)

func (ofa *OFAgent) establishConnectionToVoltha(p *probe.Probe) error {
	if p != nil {
		p.UpdateStatus("voltha", probe.ServiceStatusPreparing)
	}

	if ofa.volthaConnection != nil {
		ofa.volthaConnection.Close()
	}

	ofa.volthaConnection = nil
	ofa.volthaClient.Clear()
	try := 1
	for ofa.ConnectionMaxRetries == 0 || try < ofa.ConnectionMaxRetries {
		conn, err := grpc.Dial(ofa.VolthaApiEndPoint, grpc.WithInsecure())
		if err == nil {
			svc := voltha.NewVolthaServiceClient(conn)
			if svc != nil {
				if _, err = svc.GetVoltha(context.Background(), &empty.Empty{}); err == nil {
					logger.Debugw("Established connection to Voltha",
						log.Fields{
							"VolthaApiEndPoint": ofa.VolthaApiEndPoint,
						})
					ofa.volthaConnection = conn
					ofa.volthaClient.Set(svc)
					if p != nil {
						p.UpdateStatus("voltha", probe.ServiceStatusRunning)
					}
					ofa.events <- ofaEventVolthaConnected
					return nil
				}
			}
		}
		logger.Warnw("Failed to connect to voltha",
			log.Fields{
				"VolthaApiEndPoint": ofa.VolthaApiEndPoint,
				"error":             err.Error(),
			})
		if ofa.ConnectionMaxRetries == 0 || try < ofa.ConnectionMaxRetries {
			if ofa.ConnectionMaxRetries != 0 {
				try += 1
			}
			time.Sleep(ofa.ConnectionRetryDelay)
		}
	}
	if p != nil {
		p.UpdateStatus("voltha", probe.ServiceStatusFailed)
	}
	return errors.New("failed-to-connect-to-voltha")
}
