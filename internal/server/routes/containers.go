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

// ContainerCreate - create a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerCreate
// POST "/containers/create"
func (cr *Router) ContainerCreate(c *gin.Context) {
	in := &ContainerCreateRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	tainr := &types.Container{
		Name:         in.Name,
		Image:        in.Image,
		Cmd:          in.Cmd,
		Env:          in.Env,
		ExposedPorts: in.ExposedPorts,
		Labels:       in.Labels,
		Binds:        in.HostConfig.Binds,
	}
	if err := cr.db.SaveContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"Id": tainr.ID,
	})
}

// ContainerStart - start a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerStart
// POST "/containers/:id/start"
func (cr *Router) ContainerStart(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.db.LoadContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	running, _ := cr.kubernetes.IsContainerRunning(tainr)
	if !running {
		if err := cr.kubernetes.StartContainer(tainr); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
	} else {
		klog.Warningf("container %s already running", id)
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerDelete - remove a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerDelete
// DELETE "/containers/:id"
func (cr *Router) ContainerDelete(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.db.LoadContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	tainr.SignalStop()
	if err := cr.kubernetes.DeleteContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	if err := cr.db.DeleteContainer(tainr); err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerInfo - return low-level information about a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerInspect
// GET "/containers/:id/json"
func (cr *Router) ContainerInfo(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.db.LoadContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	status, err := cr.kubernetes.GetContainerStatus(tainr)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Id":    id,
		"Image": tainr.Image,
		"Config": gin.H{
			"Image":  tainr.Image,
			"Labels": tainr.Labels,
			"Env":    tainr.Env,
			"Cmd":    tainr.Cmd,
		},
		"NetworkSettings": gin.H{
			"Networks": gin.H{
				"bridge": gin.H{},
			},
			"Ports": cr.getNetworkSettingsPorts(tainr),
		},
		"HostConfig": gin.H{
			"NetworkMode": "host",
		},
		"State": gin.H{
			"Health": gin.H{
				"Status": status["Status"],
			},
			"Running":    status["Running"] == "running",
			"Status":     status["Running"],
			"Paused":     false,
			"Restarting": false,
			"OOMKilled":  false,
			"Dead":       false,
			"StartedAt":  "2021-01-01T00:00:00Z",
			"FinishedAt": "0001-01-01T00:00:00Z",
			"ExitCode":   0,
			"Error":      "",
		},
	})
}

// getNetworkSettingsPorts will return the mapped ports of the container
// as k8s ports structure to be used in network settings.
func (cr *Router) getNetworkSettingsPorts(tainr *types.Container) gin.H {
	res := gin.H{}
	for dst, src := range tainr.MappedPorts {
		p := fmt.Sprintf("%d/tcp", dst)
		res[p] = []gin.H{
			{
				"HostIp":   "localhost",
				"HostPort": fmt.Sprintf("%d", src),
			},
		}
	}
	return res
}
