package main

import (
	"log"
	"net/http"
	"os"

	"github.com/RHSyseng/locks-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"
	pb "k8s.io/cri-api/pkg/apis/runtime/v1"
)

var version = "develop"

func main() {
	var (
		procfsPath    = kingpin.Flag("lock.procfsPath", "Path to procfs filesystem.").Default("/proc").String()
		crioSocket    = kingpin.Flag("lock.crioSocket", "Path to cri-o socket.").Default("/var/run/crio/crio.sock").String()
		logLevel      = kingpin.Flag("log.level", "Log level.").Default("info").String()
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":3000").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	)
	kingpin.HelpFlag.Short('h')
	kingpin.Version("locks-exporter version: " + version)
	kingpin.Parse()

	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		log.Printf("Unable to parse log level. Falling back to info.")
		level = logrus.InfoLevel
	}

	logger := logrus.New()
	logger.SetLevel(level)

	// we must use insecure since there is no auth outside of filesystem permissions here
	cc, err := grpc.Dial("unix:///"+*crioSocket, grpc.WithInsecure())
	if err != nil {
		logger.Fatalf("Failed to connect to crio: %s", err)
	}

	// use a blank registry to remove the default collectors for Go and promhttp
	reg := prometheus.NewRegistry()
	reg.MustRegister(collector.New(logger, *procfsPath, pb.NewRuntimeServiceClient(cc)))

	http.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	// serve a friendly page at index with a link to the proper metrics endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>File Locks Exporter</title></head>
             <body>
             <h1>File Locks Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	logger.Infof("Listening on address %s", *listenAddress)
	srv := &http.Server{Addr: *listenAddress}

	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalf("Failed to run HTTP server: %s", err)
		os.Exit(1)
	}
}
