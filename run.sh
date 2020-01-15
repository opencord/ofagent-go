#!/bin/bash
<<COMMENT
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
COMMENT
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

