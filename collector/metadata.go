package collector

import (
	"regexp"

	"github.com/prometheus/procfs"
)

type container struct {
	containerName string
	podName       string
	namespace     string
}

const nameArg = "-n"

// container name matches k8s_<container_name>_<pod_name>_<namespace>_***
var namePattern = regexp.MustCompile(`^k8s_([^_]+)_([^_]+)_([^_]+)_`)

func (c *Collector) getContainerMetadata(pid int) *container {
	proc, err := c.fs.Proc(pid)
	if err != nil {
		c.logger.Warnf("Failed to retrieve pid from procfs: %s", err)
		return nil
	}
	// get parent proc running conmon
	parent, err := c.getParent(proc)
	if err != nil {
		c.logger.Warnf("Failed to find parent from procfs: %s", err)
		return nil
	}

	// get full command line of conmon proc
	cmd, err := parent.CmdLine()
	if err != nil {
		c.logger.Warnf("Failed to find command line from procfs: %s", err)
		return nil
	}

	// extract the "name" argument and get the container metadata
	for i, arg := range cmd {
		if arg == nameArg {
			// the "name" value will come immediately after "-n"
			match := namePattern.FindStringSubmatch(cmd[i+1])
			if match != nil {
				return &container{
					containerName: match[1],
					podName:       match[2],
					namespace:     match[3],
				}
			}
		}
	}

	// bar all else, we return nil
	return nil
}

func (c *Collector) getParent(p procfs.Proc) (procfs.Proc, error) {
	stat, err := p.Stat()
	if err != nil {
		return procfs.Proc{}, err
	}

	// return if already top level
	if stat.PPID == 1 {
		return p, nil
	}

	parent, err := c.fs.Proc(stat.PPID)
	if err != nil {
		return procfs.Proc{}, err
	}

	// recurse until parent pid is 1
	return c.getParent(parent)
}
