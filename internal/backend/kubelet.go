package backend

import (
	"context"
	"fmt"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func (in *instance) waitKubeletReady(tainr *types.Container, container string, wait int) error {
	if !in.waitKubelet {
		return nil
	}

	limit := int64(1)
	opts := &corev1.PodLogOptions{Container: container, LimitBytes: &limit}

	var err error
	for max := 0; max < wait; max++ {
		var stream io.ReadCloser
		stream, err = in.cli.CoreV1().Pods(in.namespace).GetLogs(tainr.GetPodName(), opts).Stream(context.Background())
		if err == nil {
			stream.Close()
			return nil
		}
		klog.V(2).Infof("kubelet not ready for %s: %s", tainr.GetPodName(), err)
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timeout waiting for kubelet: %w", err)
}
