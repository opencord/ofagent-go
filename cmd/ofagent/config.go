package main

import (
	"flag"
)

type Config struct {
	ofControllerEndPoint string
	volthaApiEndPoint    string
	logLevel             string
	banner               bool
	version              bool
	probeEndPoint        string
	cpuProfile           string
	memProfile           string
}

func parseCommandLineArguments() (*Config, error) {
	config := Config{}

	flag.BoolVar(&(config.banner),
		"banner",
		true,
		"display application banner on startup")
	flag.BoolVar(&(config.version),
		"version",
		false,
		"display application version and exit")
	flag.StringVar(&(config.ofControllerEndPoint),
		"controller",
		"onos-openflow:6653",
		"connection to the OF controller specified as host:port")
	flag.StringVar(&(config.ofControllerEndPoint),
		"O",
		"onos-openflow:6653",
		"(short) connection to the OF controller specified as host:port")
	flag.StringVar(&(config.volthaApiEndPoint),
		"voltha",
		"voltha-api-server:50060",
		"connection to the VOLTHA API server specified as host:port")
	flag.StringVar(&(config.volthaApiEndPoint),
		"A",
		"voltha-api-server:50060",
		"(short) connection to the VOLTHA API server specified as host:port")
	flag.StringVar(&(config.probeEndPoint),
		"probe",
		":8080",
		"address and port on which to listen for k8s live and ready probe requests")
	flag.StringVar(&(config.probeEndPoint),
		"P",
		":8080",
		"(short) address and port on which to listen for k8s live and ready probe requests")
	flag.StringVar(&(config.logLevel),
		"loglevel",
		"WARN",
		"initial log level setting, overriden by any value set in configuration store")
	flag.StringVar(&(config.logLevel),
		"L",
		"WARN",
		"(short) initial log level setting, overriden by any value set in configuration store")
	flag.StringVar(&(config.cpuProfile),
		"cpuprofile",
		"",
		"write cpu profile to 'file' if specified")
	flag.StringVar(&(config.memProfile),
		"memprofile",
		"",
		"write memory profile to 'file' if specified")

	return &config, nil
}
