package collector

import (
	"context"

	pb "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type container struct {
	id            string
	containerName string
	podName       string
	namespace     string
}

const containerLabel = "io.kubernetes.container.name"
const podLabel = "io.kubernetes.pod.name"
const namespaceLabel = "io.kubernetes.pod.namespace"

func (c *Collector) getContainerMetadata(id string) (container, error) {
	resp, err := c.criClient.ContainerStatus(context.Background(),
		&pb.ContainerStatusRequest{ContainerId: id})
	if err != nil {
		return container{}, err
	}

	// return a smaller struct about the container to avoid passing around
	// a huge ContainerStatusResponse
	out := container{
		id:            id,
		containerName: resp.GetStatus().GetLabels()[containerLabel],
		podName:       resp.GetStatus().GetLabels()[podLabel],
		namespace:     resp.GetStatus().GetLabels()[namespaceLabel],
	}
	return out, nil
}
