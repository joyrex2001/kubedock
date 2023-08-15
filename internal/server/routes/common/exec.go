package common

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
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerExec
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/exec/operation/ContainerExecLibpod
// POST "/containers/:id/exec"
// POST "/libpod/containers/:id/exec"
func ContainerExec(cr *ContextRouter, c *gin.Context) {
	in := &ContainerExecRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	if in.Env != nil && len(in.Env) > 0 {
		httputil.Error(c, http.StatusBadRequest, fmt.Errorf("env variables not supported"))
		return
	}

	if !in.Stdout && !in.Stderr {
		in.Stdout = true
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
		TTY:         in.Tty,
		Stderr:      in.Stderr,
		Stdout:      in.Stdout,
		Stdin:       in.Stdin,
	}
	if err := cr.DB.SaveExec(exec); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"Id": exec.ID,
	})
}

// ExecInfo - return low-level information about an exec instance.
// https://docs.docker.com/engine/api/v1.41/#operation/ExecInspect
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/exec/operation/ExecInspectLibpod
// GET "/exec/:id/json"
// GET "/libpod/exec/:id/json"
func ExecInfo(cr *ContextRouter, c *gin.Context) {
	id := c.Param("id")
	exec, err := cr.DB.GetExec(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ID":         id,
		"OpenStderr": exec.Stderr,
		"OpenStdin":  exec.Stdin,
		"OpenStdout": exec.Stdout,
		"Running":    false,
		"ExitCode":   exec.ExitCode,
		"ProcessConfig": gin.H{
			"tty":        exec.TTY,
			"arguments":  exec.Cmd,
			"entrypoint": "",
		},
	})
}

// ExecStart - start an exec instance.
// https://docs.docker.com/engine/api/v1.41/#operation/ExecStart
// POST "/exec/:id/start"
func ExecStart(cr *ContextRouter, c *gin.Context) {
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

// ExecResize - start an exec instance.
// https://docs.docker.com/engine/api/v1.41/#operation/ExecResize
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/exec/operation/ExecResizeLibpod
// POST "/exec/:id/resize"
// POST "/libpod/exec/:id/resize"
func ExecResize(cr *ContextRouter, c *gin.Context) {
	id := c.Param("id")
	_, err := cr.DB.GetExec(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
