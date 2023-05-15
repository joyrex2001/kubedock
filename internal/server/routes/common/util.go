package common

import (
	"os"
	"time"

	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/model/types"
)

// StartContainer will start given container and saves the appropriate state
// in the database.
func StartContainer(cr *ContextRouter, tainr *types.Container) error {
	state, err := cr.Backend.StartContainer(tainr)
	if err != nil {
		if klog.V(2) {
			klog.Infof("container %s log output:", tainr.ShortID)
			stop := make(chan struct{}, 1)
			_ = cr.Backend.GetLogs(tainr, false, 100, stop, os.Stderr)
			close(stop)
		}
		return err
	}

	tainr.HostIP = "0.0.0.0"
	if cr.Config.PortForward {
		cr.Backend.CreatePortForwards(tainr)
	} else {
		if len(tainr.GetServicePorts()) > 0 {
			ip, err := cr.Backend.GetPodIP(tainr)
			if err != nil {
				return err
			}
			tainr.HostIP = ip
			if cr.Config.ReverseProxy {
				cr.Backend.CreateReverseProxies(tainr)
			}
		}
	}

	tainr.Stopped = false
	tainr.Killed = false
	tainr.Failed = (state == backend.DeployFailed)
	tainr.Completed = (state == backend.DeployCompleted)
	tainr.Running = (state == backend.DeployRunning)

	return cr.DB.SaveContainer(tainr)
}

// UpdateContainerStatus will check if the started container is finished and will
// update the container database record accordingly.
func UpdateContainerStatus(cr *ContextRouter, tainr *types.Container) {
	if tainr.Completed {
		return
	}
	if !cr.Limiter.Allow() {
		klog.V(2).Infof("rate-limited status request for container: %s", tainr.ID)
		return
	}
	status, err := cr.Backend.GetContainerStatus(tainr)
	if err != nil {
		klog.Warningf("container status error: %s", err)
		tainr.Failed = true
	}
	if status == backend.DeployCompleted {
		tainr.Finished = time.Now()
		tainr.Completed = true
		tainr.Running = false
	}
}
