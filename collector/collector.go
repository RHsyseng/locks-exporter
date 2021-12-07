package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
	"github.com/sirupsen/logrus"
)

const Namespace = "locks"

// interface test
var _ prometheus.Collector = (*Collector)(nil)

// Exporter collects metrics from a local Raspberry Pi
type Collector struct {
	logger     *logrus.Logger
	procfsPath string
	fs         procfs.FS

	containerFileLocks *prometheus.Desc
}

// New returns an initialized collector
func New(logger *logrus.Logger, procfsPath string) (*Collector, error) {
	fs, err := procfs.NewFS(procfsPath)
	coll := &Collector{
		logger:     logger,
		procfsPath: procfsPath,
		fs:         fs,
		containerFileLocks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "pod", "file_locks"),
			"Number of file locks held by processes in container",
			[]string{"namespace", "pod", "container"},
			nil,
		),
	}
	return coll, err
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
	containerMeta := make(map[string]container)
	for pid, count := range locks {
		container := c.findContainerId(pid)
		// we'll ignore pids not running in containers
		if len(container) == 0 {
			continue
		}
		// populate metadata map if it doesn't already exist
		if _, ok := containerMeta[container]; !ok {
			meta := c.getContainerMetadata(pid)
			containerMeta[container] = meta
		}
		containers[container] += count
	}
	c.logger.Debugf("Mapped locks to %d containers", len(containers))

	// now that we have lock counts per container id, we will annotate them
	// with metadata (pod name, container name, namespace)
	for cId, count := range containers {
		meta := containerMeta[cId]
		ch <- prometheus.MustNewConstMetric(
			c.containerFileLocks,
			prometheus.GaugeValue,
			float64(count),
			meta.namespace,
			meta.podName,
			meta.containerName,
		)
	}
}
