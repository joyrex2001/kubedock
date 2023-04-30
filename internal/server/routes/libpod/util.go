package libpod

import (
	"os"
	"time"

	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/model/types"
)

// startContainer will start given container and saves the appropriate state
// in the database.
func (cr *Router) startContainer(tainr *types.Container) error {
	state, err := cr.kub.StartContainer(tainr)
	if err != nil {
		if klog.V(2) {
			klog.Infof("container %s log output:", tainr.ShortID)
			stop := make(chan struct{}, 1)
			_ = cr.kub.GetLogs(tainr, false, 100, stop, os.Stderr)
			close(stop)
		}
		return err
	}

	tainr.HostIP = "0.0.0.0"
	if cr.cfg.PortForward {
		cr.kub.CreatePortForwards(tainr)
	} else {
		if len(tainr.GetServicePorts()) > 0 {
			ip, err := cr.kub.GetPodIP(tainr)
			if err != nil {
				return err
			}
			tainr.HostIP = ip
			if cr.cfg.ReverseProxy {
				cr.kub.CreateReverseProxies(tainr)
			}
		}
	}

	tainr.Stopped = false
	tainr.Killed = false
	tainr.Failed = (state == backend.DeployFailed)
	tainr.Completed = (state == backend.DeployCompleted)
	tainr.Running = (state == backend.DeployRunning)

	return cr.db.SaveContainer(tainr)
}

// updateContainerStatus will check if the started container is finished and will
// update the container database record accordingly.
func (cr *Router) updateContainerStatus(tainr *types.Container) {
	if tainr.Completed {
		return
	}
	if !cr.plim.Allow() {
		klog.V(2).Infof("rate-limited status request for container: %s", tainr.ID)
		return
	}
	status, err := cr.kub.GetContainerStatus(tainr)
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
