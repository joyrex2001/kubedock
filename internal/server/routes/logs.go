package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// POST "/containers/:id/logs"
func (cr *Router) ContainerLogs(c *gin.Context) {
	id := c.Param("id")
	follow, _ := strconv.ParseBool(c.Query("follow"))
	// TODO: implement since
	// TODO: implement until
	// TODO: implement tail
	tainr, err := cr.factory.Load(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	running, err := cr.kubernetes.IsContainerRunning(tainr)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	if !running {
		httputil.Error(c, http.StatusNotFound, fmt.Errorf("container %s not running", id))
		return
	}

	r := c.Request
	w := c.Writer

	w.WriteHeader(http.StatusOK)

	if !follow {
		if err := cr.kubernetes.GetLogs(tainr, follow, w); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
		return
	}

	in, out, err := httputil.HijackConnection(w)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	defer httputil.CloseStreams(in, out)

	if _, ok := r.Header["Upgrade"]; ok {
		fmt.Fprint(out, "HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n")
	} else {
		fmt.Fprint(out, "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.docker.raw-stream\r\n")
	}

	// copy headers that were removed as part of hijack
	if err := w.Header().WriteSubset(out, nil); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	fmt.Fprint(out, "\r\n")

	if err := cr.kubernetes.GetLogs(tainr, follow, out); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
}
