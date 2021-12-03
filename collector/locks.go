package collector

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

var procLockRE = regexp.MustCompile(`(?m)^(?:\S+\s+){4}(\S+)`)
var cgroupRE = regexp.MustCompile(`(?m)^1:.+/crio-([\w\d]+)\.scope`)

func (c *Collector) getLocks() (map[int]int, error) {
	file, err := os.ReadFile(c.procfsPath + "/locks")
	if err != nil {
		return nil, err
	}

	// iterate through all locks
	locks := make(map[int]int)
	for _, lock := range procLockRE.FindAllSubmatch(file, -1) {
		pid, err := strconv.Atoi(string(lock[1]))
		if err != nil {
			c.logger.Debugf("Error parsing pid: %s", err)
			continue
		}
		// a pid of -1 is used for open file description locks,
		// but can be ignored for our purposes of mapping to containers
		if pid < 0 {
			continue
		}
		locks[pid] += 1
	}

	return locks, nil
}

func (c *Collector) findContainer(pid int) string {
	// find the cgroup from the /proc/<pid>/cgroup file
	file, err := os.ReadFile(fmt.Sprintf("%s/%d/cgroup", c.procfsPath, pid))
	if err != nil {
		c.logger.Debugf("Failed to open proc cgroup for pid %d: %s", pid, err)
		return ""
	}

	// use a regex pattern to extract the container id from cgroup name
	match := cgroupRE.FindSubmatch(file)
	if len(match) == 0 {
		c.logger.Debugf("No crio container found for pid %d", pid)
		return ""
	}
	return string(match[1])
}
