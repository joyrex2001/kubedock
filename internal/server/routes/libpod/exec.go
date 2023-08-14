package libpod

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes/common"
)

// ExecStart - start an exec instance.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/exec/operation/ExecStartLibpod
// POST "/libpod/exec/:id/start"
func ExecStart(cr *common.ContextRouter, c *gin.Context) {
	req := &common.ExecStartRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	if req.Detach {
		httputil.Error(c, http.StatusBadRequest, fmt.Errorf("detached mode not supported"))
		return
	}

	id := c.Param("id")
	exec, err := cr.DB.GetExec(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	if exec.TTY {
		httputil.Error(c, http.StatusBadRequest, fmt.Errorf("tty mode not supported"))
		return
	}

	tainr, err := cr.DB.GetContainer(exec.ContainerID)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	r := c.Request
	w := c.Writer
	w.WriteHeader(http.StatusOK)

	in, out, err := httputil.HijackConnection(w)
	if err != nil {
		klog.Errorf("error during hijack connection: %s", err)
		return
	}
	defer httputil.CloseStreams(in, out)
	httputil.UpgradeConnection(r, out)

	code, err := cr.Backend.ExecContainer(tainr, exec, in, out)
	if err != nil {
		klog.Errorf("error during exec: %s", err)
		return
	}
	exec.ExitCode = code
	if err := cr.DB.SaveExec(exec); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
}
