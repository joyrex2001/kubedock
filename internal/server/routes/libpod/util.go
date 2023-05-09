package libpod

import (
	"os"
	"strings"
	"time"

	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/routes"
)

// addNetworkAliases will add the networkaliases as defined in the provided
// NetworksProperty to the container.
func addNetworkAliases(tainr *types.Container, networks map[string]NetworksProperty) {
	aliases := []string{}
	done := map[string]string{tainr.ShortID: tainr.ShortID}
	for _, netwp := range networks {
		for _, a := range netwp.Aliases {
			if _, ok := done[a]; !ok {
				alias := strings.ToLower(a)
				aliases = append(aliases, alias)
				done[alias] = alias
			}
		}
	}
	tainr.NetworkAliases = aliases
}

// startContainer will start given container and saves the appropriate state
// in the database.
func startContainer(cr *routes.ContextRouter, tainr *types.Container) error {
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

// updateContainerStatus will check if the started container is finished and will
// update the container database record accordingly.
func updateContainerStatus(cr *routes.ContextRouter, tainr *types.Container) {
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
