package backend

import (
	"context"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// GetLogs will write the logs for given container to given writer.
func (in *instance) GetLogs(tainr *types.Container, follow bool, count int, w io.Writer) error {
	tail := int64(count)
	options := v1.PodLogOptions{
		Container: "main",
		Follow:    follow,
		TailLines: &tail,
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

	stop := make(chan struct{}, 1)
	tainr.AddStopChannel(stop)

	for {
		// close when container is done
		select {
		case <-stop:
			return nil
		default:
		}
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
func (in *instance) getFirstPodName(tainr *types.Container) (string, error) {
	pods, err := in.getPods(tainr)
	if err != nil {
		return "", err
	}

	names := []string{}
	for _, p := range pods {
		names = append(names, p.Name)
	}

	if len(names) == 0 {
		return "", fmt.Errorf("no running pods for %s", in.getContainerName(tainr))
	}

	return names[0], nil
}
