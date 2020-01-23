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
set GO111MODULE=on
go build -mod=vendor -o /build/ofagent \
        -ldflags \
        "-X github.com/opencord/voltha-lib-go/v2/pkg/version.version=$org_label_schema_version \
         -X github.com/opencord/voltha-lib-go/v2/pkg/version.vcsRef=$org_label_schema_vcs_ref  \
         -X github.com/opencord/voltha-lib-go/v2/pkg/version.vcsDirty=$org_opencord_vcs_dirty \
         -X github.com/opencord/voltha-lib-go/v2/pkg/version.goVersion=$(go version 2>&1 | sed -E  's/.*go([0-9]+\.[0-9]+\.[0-9]+).*/\1/g') \
         -X github.com/opencord/voltha-lib-go/v2/pkg/version.os=$(go env GOHOSTOS) \
         -X github.com/opencord/voltha-lib-go/v2/pkg/version.arch=$(go env GOHOSTARCH) \
         -X github.com/opencord/voltha-lib-go/v2/pkg/version.buildTime=$org_label_schema_build_date" \
        ./cmd/ofagent

#go build ./cmd/ofagent -mod=vendor -o build/ofagent-go

if [  "$1" = "debug" ]
then
    echo DEBUGGING
    ./build/ofagent-go -memprofile=debugmem -cpuprofile=debugcpu -debug -ofaddress=localhost openflowPort=6653 -volthaAddress=localhost -volthaPort=50057  2>&1 |tee output
else
    echo NOT DEBUGGING
    ./build/ofagent-go -ofaddress=localhost openflowPort=6653 -volthaAddress=localhost -volthaPort=50057 -memprofile=mem -cpuprofile=cpu 2>&1 |tee output
fi

