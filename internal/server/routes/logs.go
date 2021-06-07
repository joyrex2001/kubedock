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

	running, err := cr.kub.IsContainerRunning(tainr)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	if !running && !cr.kub.IsContainerCompleted(tainr) {
		httputil.Error(c, http.StatusNotFound, fmt.Errorf("container %s not running", id))
		return
	}

	r := c.Request
	w := c.Writer
	w.WriteHeader(http.StatusOK)

	if !follow {
		if err := cr.kub.GetLogs(tainr, follow, 100, w); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
		return
	}

	in, out, err := httputil.HijackConnection(w)
	if err != nil {
		klog.Errorf("error during hijack connection: %s", err)
		return
	}
	defer httputil.CloseStreams(in, out)
	httputil.UpgradeConnection(r, out)

	if err := cr.kub.GetLogs(tainr, follow, 100, out); err != nil {
		klog.Errorf("error retrieving logs: %s", err)
		return
	}
}
