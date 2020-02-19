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
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-lib-go/v3/pkg/probe"
	"github.com/opencord/voltha-lib-go/v3/pkg/version"
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

func init() {
	_, err := log.AddPackage(log.JSON, log.DebugLevel, nil)
	if err != nil {
		log.Errorw("unable-to-register-package-to-the-log-map", log.Fields{"error": err})
	}
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

	if err := log.SetLogLevel(log.DebugLevel); err != nil {
		log.Errorw("set-log-level", log.Fields{
			"level": log.DebugLevel,
			"error": err})
	}
	if _, err := log.SetDefaultLogger(log.JSON, log.DebugLevel, log.Fields{"component": "ofagent"}); err != nil {
		log.Errorw("set-default-log-level", log.Fields{
			"level": log.DebugLevel,
			"error": err})

	}

	var logLevel int
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
		logLevel = log.WarnLevel
	}
	if _, err := log.SetDefaultLogger(log.JSON, logLevel, log.Fields{"component": "ofagent"}); err != nil {
		log.Errorw("set-default-log-level", log.Fields{
			"level": logLevel,
			"error": err})

	}
	log.SetAllLogLevel(logLevel)

	log.Infow("ofagent-startup-configuration",
		log.Fields{
			"configuration": fmt.Sprintf("%+v", *config),
		})

	_, err := log.AddPackage(log.JSON, logLevel, nil)
	if err != nil {
		log.Errorw("unable-to-register-package-to-the-log-map", log.Fields{"error": err})
	}

	/*
	 * Create and start the liveness and readiness container management probes. This
	 * is done in the main function so just in case the main starts multiple other
	 * objects there can be a single probe end point for the process.
	 */
	p := &probe.Probe{}
	go p.ListenAndServe(config.ProbeEndPoint)

	/*
	 * Create a context which to start the services used by the applicaiton
	 * and attach the probe to that context
	 */
	ctx := context.WithValue(context.Background(), probe.ProbeContextKey, p)

	ofa, err := ofagent.NewOFAgent(&ofagent.OFAgent{
		OFControllerEndPoint:      config.OFControllerEndPoint,
		VolthaApiEndPoint:         config.VolthaApiEndPoint,
		DeviceListRefreshInterval: config.DeviceListRefreshInterval,
		ConnectionMaxRetries:      config.ConnectionMaxRetries,
		ConnectionRetryDelay:      config.ConnectionRetryDelay,
	})
	if err != nil {
		log.Fatalw("failed-to-create-ofagent",
			log.Fields{
				"error": err})
	}
	ofa.Run(ctx)
}
