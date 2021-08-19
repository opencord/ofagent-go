module github.com/opencord/ofagent-go

go 1.16

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/opencord/voltha-lib-go/v6 => ../voltha-lib-go
	github.com/opencord/voltha-protos/v4 => ../voltha-protos
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.25.1
)

require (
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/opencord/goloxi v1.0.1
	github.com/opencord/voltha-lib-go/v6 v6.0.2
	github.com/opencord/voltha-protos/v4 v4.3.0-dev
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	google.golang.org/grpc v1.39.1
)
