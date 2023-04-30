package docker

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes"
)

// ContainerExec - create an exec instance.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerInspect
// POST "/containers/:id/exec"
func ContainerExec(cr *routes.ContextRouter, c *gin.Context) {
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
	_, err := cr.DB.GetContainer(id)
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
	if err := cr.DB.SaveExec(exec); err != nil {
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
func ExecStart(cr *routes.ContextRouter, c *gin.Context) {
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
	exec, err := cr.DB.GetExec(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
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

	code, err := cr.Backend.ExecContainer(tainr, exec, out)
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

// ExecInfo - return low-level information about an exec instance.
// https://docs.docker.com/engine/api/v1.41/#operation/ExecInspect
// GET "/exec/:id/json"
func ExecInfo(cr *routes.ContextRouter, c *gin.Context) {
	id := c.Param("id")
	exec, err := cr.DB.GetExec(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ID":       id,
		"Running":  false,
		"ExitCode": exec.ExitCode,
		"ProcessConfig": gin.H{
			"arguments":  exec.Cmd,
			"entrypoint": "",
		},
	})
}
