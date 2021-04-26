package kubernetes

import (
	"context"
	"fmt"
	"io"
	"log"

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
		buf := make([]byte, 2000)
		num, err := stream.Read(buf)
		if num == 0 {
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
		w.Write(buf[:num])
	}

	log.Printf("log done... for %s", tainr.GetKubernetesName())

	return nil
}
