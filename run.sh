#!/bin/bash

if [ -z output ]
then
    rm output
fi
go build -v  -o build/ofagent-go -mod=vendor

if [  "$1" = "debug" ]
then
    echo DEBUGGING
    ./build/ofagent-go -memprofile=debugmem -cpuprofile=debugcpu -debug -ofaddress=localhost openflowPort=6653 -volthaAddress=localhost -volthaPort=50057  2>&1 |tee output
else
    echo NOT DEBUGGING
    ./build/ofagent-go -ofaddress=localhost openflowPort=6653 -volthaAddress=localhost -volthaPort=50057 -memprofile=mem -cpuprofile=cpu 2>&1 |tee output
fi

