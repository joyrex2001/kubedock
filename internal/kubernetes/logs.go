package kubernetes

import (
	"context"
	"fmt"
	"io"

	"github.com/joyrex2001/kubedock/internal/container"
	v1 "k8s.io/api/core/v1"
)

// GetPodLogs will write the logs for given container to given writer.
func (in *instance) GetLogs(tainr container.Container, follow bool, w io.Writer) error {
	count := int64(100)
	options := v1.PodLogOptions{
		Container: tainr.GetKubernetesName(),
		Follow:    follow,
		TailLines: &count,
	}

	pod, err := in.GetPodNames(tainr)
	if err != nil {
		return err
	}

	if len(pod) == 0 {
		return fmt.Errorf("no running pods for %s", tainr.GetKubernetesName())
	}

	req := in.cli.CoreV1().
		Pods(in.namespace).
		GetLogs(pod[0], &options)
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
