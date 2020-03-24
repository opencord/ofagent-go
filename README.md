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

## Building with a Local Copy of `voltha-protos` or `voltha-lib-go`
If you want to build/test using a local copy or `voltha-protos` or `voltha-lib-go`
this can be accomplished by using the environment variables `LOCAL_PROTOS` and
`LOCAL_LIB_GO`. These environment variables should be set to the filesystem
path where the local source is located, e.g.

```bash
LOCAL\_PROTOS=$HOME/src/voltha-protos
LOCAL\_LIB\_GO=$HOME/src/voltha-lib-go
```

When these environment variables are set the vendored versions of these packages
will be removed from the `vendor` directory and replaced by coping the files from
the specified locations to the `vendor` directory. *NOTE:* _this means that
the files in the `vendor` directory are no longer what is in the `git` repository
and it will take manual `git` intervention to put the original files back._

