package container

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// POST "/containers/:id/exec"
func (cr *containerRouter) ContainerExec(c *gin.Context) {
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
	_, err := cr.factory.Load(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	exec, err := cr.factory.CreateExec(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	exec.SetCmd(in.Cmd)
	exec.SetStderr(in.Stderr)
	exec.SetStdout(in.Stdout)

	c.JSON(http.StatusCreated, gin.H{
		"Id": exec.GetID(),
	})
}

// POST "/exec/:id/start"
func (cr *containerRouter) ExecStart(c *gin.Context) {
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
	exec, err := cr.factory.LoadExec(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	tainr, err := cr.factory.Load(exec.GetContainerID())
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	w := c.Writer
	w.WriteHeader(http.StatusOK)

	if err := cr.kubernetes.ExecContainer(tainr, exec, w); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
}

// GET "/exec/:id/json"
func (cr *containerRouter) ExecInfo(c *gin.Context) {
	id := c.Param("id")

	exec, err := cr.factory.LoadExec(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	_, err = cr.kubernetes.GetExecStatus(exec)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	// TODO: bogus exit code, k8s doesn't seem to support this
	c.JSON(http.StatusOK, gin.H{
		"ID":       id,
		"Running":  false,
		"ExitCode": 0,
	})
}
