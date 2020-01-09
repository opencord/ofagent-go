/*
   Copyright 2019 the original author or authors.

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

package main

import (
	"flag"

	"github.com/opencord/ofagent-go/settings"

	"github.com/opencord/ofagent-go/grpc"
	l "github.com/opencord/voltha-lib-go/v2/pkg/log"
)

func main() {
	debug := flag.Bool("debug", false, "Set Debug Level Logging")
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

	openflowAddress := flag.String("ofaddress", "localhost", "address of the openflow server")
	openflowPort := flag.Uint("openflowPort", 6653, "port the openflow server is listening on")
	volthaAddress := flag.String("volthaAddress", "localhost", "address for voltha core / afrouter")
	volthaPort := flag.Uint("volthaPort", 50057, "port that voltha core / afrouter listens on ")
	flag.Parse()

	logLevel := l.InfoLevel
	if *debug {
		logLevel = l.DebugLevel
		settings.SetDebug(true)
	} else {
		settings.SetDebug(false)
	}
	_, err := l.AddPackage(l.JSON, logLevel, nil)
	if err != nil {
		l.Errorw("unable-to-register-package-to-the-log-map", l.Fields{"error": err})
	}
	l.Infow("ofagent-startup params", l.Fields{"OpenFlowAddress": *openflowAddress,
		"OpenFlowPort":  *openflowPort,
		"VolthaAddress": *volthaAddress,
		"VolthaPort":    *volthaPort,
		"LogLevel":      logLevel,
		"CpuProfile":    *cpuprofile,
		"MemProfile":    *memprofile})

	grpc.StartClient(*volthaAddress, uint16(*volthaPort), *openflowAddress, uint16(*openflowPort))
}
