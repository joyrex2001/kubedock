package kubernetes

import (
	"context"
	"fmt"
	"io"

	"github.com/joyrex2001/kubedock/internal/container"
	v1 "k8s.io/api/core/v1"
)

// GetPodLogs will write the logs for given container to given writer.
func (in *instance) GetLogs(tainr *container.Container, follow bool, w io.Writer) error {
	count := int64(100)
	options := v1.PodLogOptions{
		Container: tainr.GetKubernetesName(),
		Follow:    follow,
		TailLines: &count,
	}

	name, err := in.getFirstPodName(tainr)
	if err != nil {
		return err
	}

	req := in.cli.CoreV1().Pods(in.namespace).GetLogs(name, &options)
	stream, err := req.Stream(context.TODO())
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		// read log input
		buf := make([]byte, 2000)
		n, err := stream.Read(buf)
		if n == 0 {
			if !follow {
				break
			}
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// write log to output
		if n, err = w.Write(buf[:n]); n == 0 || err != nil {
			break
		}
	}

	return nil
}

// getFirstPodName returns the pod name of the first pod that matches
// the container deployment.
func (in *instance) getFirstPodName(tainr *container.Container) (string, error) {
	pods, err := in.GetPods(tainr)
	if err != nil {
		return "", err
	}

	names := []string{}
	for _, p := range pods {
		names = append(names, p.Name)
	}

	if len(names) == 0 {
		return "", fmt.Errorf("no running pods for %s", tainr.GetKubernetesName())
	}

	return names[0], nil
}
