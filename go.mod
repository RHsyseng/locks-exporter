module github.com/RHSyseng/locks-exporter

go 1.16

require github.com/sirupsen/logrus v1.8.1

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
)

require (
	github.com/prometheus/client_golang v1.11.0
	google.golang.org/grpc v1.38.0 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/cri-api v0.22.4
)
