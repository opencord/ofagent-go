# Ofagent-go

Ofagent-go provides an OpenFlow management interface for Voltha.  It is a rewrite in Golang of the original ofagent that was written in python / twisted.  The main driver behind the work was to introduce true concurrency to the agent for performance/scalability reasons.

#### Building

1. Outside $GOPATH 

   1. Read-only `git clone https://github.com/opencord/ofagent-go.git` 
   2. To Contribute `git clone https://gerrit.opencord.org/ofagent-go` 

2. Compile `go build -mod=vendor -o ./build/ofagent-go`

   

#### Running

- Normal Logging 
  - `./build/ofagent-go -ofaddress=localhost openflowPort=6653 -volthaAddress=localhost -volthaPort=50057`
- Debug Logging
  - `./build/ofagent-go -debug -ofaddress=localhost openflowPort=6653 -volthaAddress=localhost -volthaPort=50057`