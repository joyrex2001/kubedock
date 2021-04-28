package container

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/container"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// POST "/containers/create"
func (cr *containerRouter) ContainerCreate(c *gin.Context) {
	in := &ContainerCreateRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	tainr, err := cr.factory.Create()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	tainr.Name = in.Name
	tainr.Image = in.Image
	tainr.Cmd = in.Cmd
	tainr.Env = in.Env
	tainr.ExposedPorts = in.ExposedPorts
	tainr.Labels = in.Labels
	tainr.Update()

	c.JSON(http.StatusCreated, gin.H{
		"Id": tainr.ID,
	})
}

// POST "/containers/:id/start"
func (cr *containerRouter) ContainerStart(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.factory.Load(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	if err := cr.kubernetes.StartContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// DELETE "/containers/:id"
func (cr *containerRouter) ContainerDelete(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.factory.Load(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	if err := cr.kubernetes.DeleteContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	if err := tainr.Delete(); err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// GET "/containers/:id/json"
func (cr *containerRouter) ContainerInfo(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.factory.Load(id)
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
func (cr *containerRouter) getNetworkSettingsPorts(tainr *container.Container) gin.H {
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
