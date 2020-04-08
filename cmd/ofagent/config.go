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
	"os"
	"strings"
	"time"
)

type Config struct {
	OFControllerEndPoints     multiFlag
	VolthaApiEndPoint         string
	LogLevel                  string
	Banner                    bool
	Version                   bool
	ProbeEndPoint             string
	CpuProfile                string
	MemProfile                string
	DeviceListRefreshInterval time.Duration
	ConnectionRetryDelay      time.Duration
	ConnectionMaxRetries      int
	KVStoreType               string
	KVStoreTimeout            time.Duration
	KVStoreHost               string
	KVStorePort               int
	InstanceID                string
}

type multiFlag []string

func (m *multiFlag) String() string {
	return "[" + strings.Join(*m, ", ") + "]"
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func parseCommandLineArguments() (*Config, error) {
	config := Config{}

	flag.BoolVar(&(config.Banner),
		"banner",
		true,
		"display application banner on startup")
	flag.BoolVar(&(config.Version),
		"version",
		false,
		"display application version and exit")
	flag.Var(&config.OFControllerEndPoints,
		"controller",
		"connection to the OF controller specified as host:port")
	flag.Var(&config.OFControllerEndPoints,
		"O",
		"(short) connection to the OF controller specified as host:port")
	flag.StringVar(&(config.VolthaApiEndPoint),
		"voltha",
		"voltha-api-server:50060",
		"connection to the VOLTHA API server specified as host:port")
	flag.StringVar(&(config.VolthaApiEndPoint),
		"A",
		"voltha-api-server:50060",
		"(short) connection to the VOLTHA API server specified as host:port")
	flag.StringVar(&(config.ProbeEndPoint),
		"probe",
		":8080",
		"address and port on which to listen for k8s live and ready probe requests")
	flag.StringVar(&(config.ProbeEndPoint),
		"P",
		":8080",
		"(short) address and port on which to listen for k8s live and ready probe requests")
	flag.StringVar(&(config.CpuProfile),
		"cpuprofile",
		"",
		"write cpu profile to 'file' if specified")
	flag.StringVar(&(config.MemProfile),
		"memprofile",
		"",
		"write memory profile to 'file' if specified")
	flag.DurationVar(&(config.ConnectionRetryDelay),
		"cd",
		3*time.Second,
		"(short) delay to wait before connection establishment retries")
	flag.DurationVar(&(config.ConnectionRetryDelay),
		"connnection-delay",
		3*time.Second,
		"delay to wait before connection establishment retries")
	flag.IntVar(&(config.ConnectionMaxRetries),
		"mr",
		0,
		"(short) number of retries when attempting to estblish a connection, 0 is unlimted")
	flag.IntVar(&(config.ConnectionMaxRetries),
		"connnection-retries",
		0,
		"number of retries when attempting to estblish a connection, 0 is unlimted")
	flag.DurationVar(&(config.DeviceListRefreshInterval),
		"dri",
		1*time.Minute,
		"(short) interval between attempts to synchronize devices from voltha to ofagent")
	flag.DurationVar(&(config.DeviceListRefreshInterval),
		"device-refresh-interval",
		1*time.Minute,
		"interval between attempts to synchronize devices from voltha to ofagent")
	flag.StringVar(&(config.KVStoreType), "kv_store_type", "etcd", "KV store type")

	flag.DurationVar(&(config.KVStoreTimeout), "kv_store_request_timeout", 5*time.Second, "The default timeout when making a kv store request")

	flag.StringVar(&(config.KVStoreHost), "kv_store_host", "voltha-etcd-cluster-client.voltha.svc.cluster.local", "KV store host")

	flag.IntVar(&(config.KVStorePort), "kv_store_port", 2379, "KV store port")

	flag.StringVar(&(config.LogLevel), "log_level", "WARN", "Log level")

	containerName := getContainerInfo()
	if len(containerName) > 0 {
		config.InstanceID = containerName
	} else {
		config.InstanceID = "openFlowAgent001"
	}

	return &config, nil
}

func getContainerInfo() string {
	return os.Getenv("HOSTNAME")
}
