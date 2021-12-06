VERSION = $(shell git rev-parse --short HEAD)
 
build:
	go build -o bin/locks-exporter -ldflags="-X 'main.version=${VERSION}'" cmd/locks-exporter/main.go

container:
	buildah build -t lock-exporter
