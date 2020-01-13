package main

import (
	"flag"
	"time"
)

type Config struct {
	OFControllerEndPoint string
	VolthaApiEndPoint    string
	LogLevel             string
	Banner               bool
	Version              bool
	ProbeEndPoint        string
	CpuProfile           string
	MemProfile           string
	ConnectionRetryDelay time.Duration
	ConnectionMaxRetries int
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
	flag.StringVar(&(config.OFControllerEndPoint),
		"controller",
		"onos-openflow:6653",
		"connection to the OF controller specified as host:port")
	flag.StringVar(&(config.OFControllerEndPoint),
		"O",
		"onos-openflow:6653",
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
	flag.StringVar(&(config.LogLevel),
		"loglevel",
		"WARN",
		"initial log level setting, overriden by any value set in configuration store")
	flag.StringVar(&(config.LogLevel),
		"L",
		"WARN",
		"(short) initial log level setting, overriden by any value set in configuration store")
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

	return &config, nil
}
