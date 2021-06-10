package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// ContainerLogs - get container logs.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerLogs
// POST "/containers/:id/logs"
func (cr *Router) ContainerLogs(c *gin.Context) {
	id := c.Param("id")
	follow, _ := strconv.ParseBool(c.Query("follow"))
	// TODO: implement since
	// TODO: implement until
	// TODO: implement tail
	tainr, err := cr.db.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	if !tainr.Running && !tainr.Completed {
		httputil.Error(c, http.StatusNotFound, fmt.Errorf("container %s is not running", tainr.ShortID))
		return
	}

	r := c.Request
	w := c.Writer
	w.WriteHeader(http.StatusOK)

	if !follow {
		stop := make(chan struct{}, 1)
		if err := cr.kub.GetLogs(tainr, follow, 100, stop, w); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
		close(stop)
		return
	}

	in, out, err := httputil.HijackConnection(w)
	if err != nil {
		klog.Errorf("error during hijack connection: %s", err)
		return
	}
	defer httputil.CloseStreams(in, out)
	httputil.UpgradeConnection(r, out)

	stop := make(chan struct{}, 1)
	tainr.AddStopChannel(stop)

	if err := cr.kub.GetLogs(tainr, follow, 100, stop, out); err != nil {
		klog.Errorf("error retrieving logs: %s", err)
		return
	}
}
