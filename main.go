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

package main

import (
	"flag"
	"log"

	"github.com/opencord/ofagent-go/grpc"
	//"log"
)

func main() {
	logConfig := flag.String("logconfig", "", "logConfigFile")
	log.Printf("LogConfig:%s", *logConfig)
	openflowAddress := flag.String("ofaddress", "localhost", "address of the openflow server")
	openflowPort := flag.Uint("openflowPort", 6653, "port the openflow server is listening on")
	volthaAddress := flag.String("volthaAddress", "localhost", "address for voltha core / afrouter")
	volthaPort := flag.Uint("volthaPort", 50057, "port that voltha core / afrouter listens on ")
	flag.Parse()

	grpc.StartClient(*volthaAddress, uint16(*volthaPort), *openflowAddress, uint16(*openflowPort))
}
