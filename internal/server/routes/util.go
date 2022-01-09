package routes

import (
	"os"
	"strings"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/model/types"
	"k8s.io/klog"
)

// addNetworkAliases will add the networkaliases as defined in the provided
// EndpointConfig to the container.
func (cr *Router) addNetworkAliases(tainr *types.Container, endp EndpointConfig) {
	aliases := []string{}
	done := map[string]string{tainr.ShortID: tainr.ShortID}
	for _, l := range [][]string{tainr.NetworkAliases, endp.Aliases} {
		for _, a := range l {
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
			ip, err := cr.kub.GetServiceClusterIP(tainr)
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
