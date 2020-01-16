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
	"fmt"
	"github.com/opencord/ofagent-go/internal/pkg/ofagent"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-lib-go/v2/pkg/version"
	"net"
	"os"
	"strconv"
	"strings"
)

func printBanner() {
	fmt.Println(`   ___  _____ _                    _   `)
	fmt.Println(`  / _ \|  ___/ \   __ _  ___ _ __ | |_ `)
	fmt.Println(` | | | | |_ / _ \ / _' |/ _ \ '_ \| __|`)
	fmt.Println(` | |_| |  _/ ___ \ (_| |  __/ | | | |_ `)
	fmt.Println(`  \___/|_|/_/   \_\__, |\___|_| |_|\__|`)
	fmt.Println(`                  |___/                `)
}

func printVersion() {
	fmt.Println("OFAgent")
	fmt.Println(version.VersionInfo.String("  "))
}

func main() {

	config, _ := parseCommandLineArguments()
	flag.Parse()

	if config.Version {
		printVersion()
		os.Exit(0)
	}

	if config.Banner {
		printBanner()
	}

	log.SetLogLevel(log.DebugLevel)
	log.SetDefaultLogger(log.JSON, log.DebugLevel,
		log.Fields{
			"component": "ofagent",
		})

	logLevel := log.WarnLevel
	switch strings.ToLower(config.LogLevel) {
	case "fatal":
		logLevel = log.FatalLevel
	case "error":
		logLevel = log.ErrorLevel
	case "warn":
		logLevel = log.WarnLevel
	case "info":
		logLevel = log.InfoLevel
	case "debug":
		logLevel = log.DebugLevel
	default:
		log.Warn("unrecognized-log-level",
			log.Fields{
				"msg":       "Unrecognized log level setting defaulting to 'WARN'",
				"component": "ofagent",
				"value":     config.LogLevel,
			})
		config.LogLevel = "WARN"
		logLevel = log.FatalLevel
	}
	log.SetDefaultLogger(log.JSON, logLevel,
		log.Fields{
			"component": "ofagent",
		})
	log.SetAllLogLevel(logLevel)

	log.Infow("ofagent-startup-configuration",
		log.Fields{
			"configuration": fmt.Sprintf("%+v", *config),
		})

	_, err := log.AddPackage(log.JSON, logLevel, nil)
	if err != nil {
		log.Errorw("unable-to-register-package-to-the-log-map", log.Fields{"error": err})
	}

	// For now parse the endpoint specifications into host + port
	vAddress, vPortS, err := net.SplitHostPort(config.VolthaApiEndPoint)
	if err != nil {
		log.Fatalw("invalid-voltha-api-endpoint",
			log.Fields{
				"endpoint": config.VolthaApiEndPoint})
	}
	vPort, err := strconv.ParseUint(vPortS, 10, 16)
	if err != nil {
		log.Fatalw("invalid-of-controller-port",
			log.Fields{
				"port":  vPortS,
				"error": err})
	}

	ofagent.StartClient(vAddress, uint16(vPort), config.OFControllerEndPoint)
}
