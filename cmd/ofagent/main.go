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
	"context"
	"flag"
	"fmt"
	"github.com/opencord/ofagent-go/internal/pkg/ofagent"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-lib-go/v2/pkg/probe"
	"github.com/opencord/voltha-lib-go/v2/pkg/version"
	"os"
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

	if config.version {
		printVersion()
		os.Exit(0)
	}

	if config.banner {
		printBanner()
	}

	logLevel := log.WarnLevel
	switch strings.ToLower(config.logLevel) {
	case "fatal":
		log.SetLogLevel(log.FatalLevel)
	case "error":
		log.SetLogLevel(log.ErrorLevel)
	case "warn":
		log.SetLogLevel(log.WarnLevel)
	case "info":
		log.SetLogLevel(log.InfoLevel)
	case "debug":
		log.SetLogLevel(log.DebugLevel)
	default:
		log.Warn("unrecognized-log-level",
			log.Fields{
				"msg":       "Unrecognized log level setting defaulting to 'WARN'",
				"component": "ofagent",
				"value":     config.logLevel,
			})
		config.logLevel = "WARN"
		log.SetLogLevel(log.WarnLevel)
	}
	log.SetDefaultLogger(log.JSON, logLevel,
		log.Fields{
			"component": "ofagent",
		})

	log.Debugw("ofagent-startup-configuration",
		log.Fields{
			"ofControllerEndPoint": config.ofControllerEndPoint,
			"volthaApiEndPoint":    config.volthaApiEndPoint,
			"logLevel":             config.logLevel,
			"banner":               config.banner,
			"version":              config.version,
			"probeEndPoint":        config.probeEndPoint,
			"cpuProfile":           config.cpuProfile,
			"memProfile":           config.memProfile,
		})

	/*
	 * Create and start the liveness and readiness container management probes. This
	 * is done in the main function so just in case the main starts multiple other
	 * objects there can be a single probe end point for the process.
	 */
	p := &probe.Probe{}
	go p.ListenAndServe(config.probeEndPoint)

	/*
	 * Create a context which to start the services used by the applicaiton
	 * and attach the probe to that context
	 */
	ctx := context.WithValue(context.TODO(), probe.ProbeContextKey, p)

	ofa, _ := ofagent.NewOFAgent(config.ofControllerEndPoint, config.volthaApiEndPoint)
	ofa.Run(ctx)
}
