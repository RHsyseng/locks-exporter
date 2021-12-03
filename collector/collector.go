package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	pb "k8s.io/cri-api/pkg/apis/runtime/v1"
)

const Namespace = "locks"

var _ prometheus.Collector = (*Collector)(nil)

// Exporter collects metrics from a local Raspberry Pi
type Collector struct {
	logger     *logrus.Logger
	procfsPath string
	criClient  pb.RuntimeServiceClient

	containerFileLocks *prometheus.Desc
}

// New returns an initialized collector
func New(logger *logrus.Logger, procfsPath string, criClient pb.RuntimeServiceClient) *Collector {
	return &Collector{
		logger:     logger,
		procfsPath: procfsPath,
		criClient:  criClient,
		containerFileLocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "pod", "file_locks"),
			"Number of file locks held by processes in container",
			[]string{"pod", "container", "namespace"},
			nil,
		),
	}
}

// Describe returns all possible metric descriptions
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.containerFileLocks
}

// Collect fetches file lock statistics from the local system
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Debug("Beginning collection")

	// get a count of locks from /proc/locks by pid
	locks, err := c.getLocks()
	if err != nil {
		c.logger.Errorf("Unable to get locks: %s", err)
		return
	}
	c.logger.Debugf("Found %d pids with locks", len(locks))

	// map pids to crio containers since it can be a many->one relationship
	containers := make(map[string]int)
	for pid, count := range locks {
		container := c.findContainer(pid)
		// we'll ignore pids not running in containers
		if len(container) == 0 {
			continue
		}
		containers[container] += count
	}
	c.logger.Debugf("Mapped locks to %d containers", len(containers))

	// now that we have lock counts per container id, we will annotate them
	// with metadata (pod name, container name, namespace) from crio
	for cId, count := range containers {
		container, err := c.getContainerMetadata(cId)
		if err != nil {
			c.logger.Warnf("Failed to get container id %s from crio: %s", cId, err)
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.containerFileLocks,
			prometheus.GaugeValue,
			float64(count),
			container.podName,
			container.containerName,
			container.namespace,
		)
	}
}
