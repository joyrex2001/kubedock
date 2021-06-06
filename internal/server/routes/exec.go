package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// ContainerExec - create an exec instance.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerInspect
// POST "/containers/:id/exec"
func (cr *Router) ContainerExec(c *gin.Context) {
	in := &ContainerExecRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	if in.Env != nil {
		httputil.Error(c, http.StatusBadRequest, fmt.Errorf("env variables not supported"))
		return
	}

	if in.Stdin {
		httputil.Error(c, http.StatusBadRequest, fmt.Errorf("stdin not supported"))
		return
	}

	if in.Tty {
		httputil.Error(c, http.StatusBadRequest, fmt.Errorf("tty not supported"))
		return
	}

	id := c.Param("id")
	_, err := cr.db.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	exec := &types.Exec{
		ContainerID: id,
		Cmd:         in.Cmd,
		Stderr:      in.Stderr,
		Stdout:      in.Stdout,
	}
	if err := cr.db.SaveExec(exec); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"Id": exec.ID,
	})
}

// ExecStart - start an exec instance.
// https://docs.docker.com/engine/api/v1.41/#operation/ExecStart
// POST "/exec/:id/start"
func (cr *Router) ExecStart(c *gin.Context) {
	req := &ExecStartRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	if req.Detach {
		httputil.Error(c, http.StatusBadRequest, fmt.Errorf("detached mode not supported"))
		return
	}

	id := c.Param("id")
	exec, err := cr.db.GetExec(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	tainr, err := cr.db.GetContainer(exec.ContainerID)
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

	code, err := cr.kub.ExecContainer(tainr, exec, out)
	if err != nil {
		klog.Errorf("error during exec: %s", err)
		return
	}
	exec.ExitCode = code
	if err := cr.db.SaveExec(exec); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
}

// ExecInfo - return low-level information about an exec instance.
// https://docs.docker.com/engine/api/v1.41/#operation/ExecInspect
// GET "/exec/:id/json"
func (cr *Router) ExecInfo(c *gin.Context) {
	id := c.Param("id")
	exec, err := cr.db.GetExec(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	// TODO: bogus exit code, k8s doesn't seem to support this
	c.JSON(http.StatusOK, gin.H{
		"ID":       id,
		"Running":  false,
		"ExitCode": exec.ExitCode,
	})
}
