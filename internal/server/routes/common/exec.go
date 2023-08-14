package common

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

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
		"ID":       id,
		"Running":  false,
		"ExitCode": exec.ExitCode,
		"ProcessConfig": gin.H{
			"arguments":  exec.Cmd,
			"entrypoint": "",
		},
	})
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
