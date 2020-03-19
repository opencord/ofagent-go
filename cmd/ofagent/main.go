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
	conf "github.com/opencord/voltha-lib-go/v3/pkg/config"
	"github.com/opencord/voltha-lib-go/v3/pkg/db/kvstore"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-lib-go/v3/pkg/probe"
	"github.com/opencord/voltha-lib-go/v3/pkg/version"
	"os"
	"strconv"
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

func setLogConfig(ctx context.Context, kvStoreHost, kvStoreType string, kvStorePort, kvStoreTimeout int) (kvstore.Client, error) {
	client, err := kvstore.NewEtcdClient(kvStoreHost+":"+strconv.Itoa(kvStorePort), kvStoreTimeout)
	if err != nil {
		return nil, err
	}

	cm := conf.NewConfigManager(client, kvStoreType, kvStoreHost, kvStorePort, kvStoreTimeout)
	go conf.StartLogLevelConfigProcessing(cm, ctx)
	return client, nil
}

func stop(ctx context.Context, kvClient kvstore.Client) {

	// Cleanup - applies only if we had a kvClient
	if kvClient != nil {
		// Release all reservations
		if err := kvClient.ReleaseAllReservations(ctx); err != nil {
			log.Infow("fail-to-release-all-reservations", log.Fields{"error": err})
		}
		// Close the DB connection
		kvClient.Close()
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

	// Setup logging

	logLevel, err := log.StringToLogLevel(config.LogLevel)
	if err != nil {
		log.Fatalf("Cannot setup logging, %s", err)
	}

	// Setup default logger - applies for packages that do not have specific logger set
	if _, err := log.SetDefaultLogger(log.JSON, logLevel, log.Fields{"instanceId": config.InstanceID}); err != nil {
		log.With(log.Fields{"error": err}).Fatal("Cannot setup logging")
	}

	// Update all loggers (provisionned via init) with a common field
	if err := log.UpdateAllLoggers(log.Fields{"instanceId": config.InstanceID}); err != nil {
		log.With(log.Fields{"error": err}).Fatal("Cannot setup logging")
	}

	log.SetAllLogLevel(logLevel)

	defer func() {
		err := log.CleanUp()
		if err != nil {
			log.Errorw("unable-to-flush-any-buffered-log-entries", log.Fields{"error": err})
		}
	}()

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

	client, err := setLogConfig(ctx, config.KVStoreHost, config.KVStoreType, config.KVStorePort, config.KVStoreTimeout)
	if err != nil {
		log.Warnw("unable-to-create-kvstore-client", log.Fields{"error": err})
	}

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
	stop(ctx, client)
}
