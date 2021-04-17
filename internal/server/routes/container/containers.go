package container

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// POST "/containers/create"
func (cr *containerRouter) ContainerCreate(c *gin.Context) {
	in := &ContainerCreateRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	ctainr, err := cr.factory.Create()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	ctainr.SetName(in.Name)
	ctainr.SetImage(in.Image)
	ctainr.SetCmd(in.Cmd)
	ctainr.SetEnv(in.Env)
	ctainr.SetExposedPorts(in.ExposedPorts)
	ctainr.SetLabels(in.Labels)
	ctainr.Update()

	c.JSON(http.StatusCreated, gin.H{
		"Id": ctainr.GetID(),
	})
}

// POST "/containers/:id/start"
func (cr *containerRouter) ContainerStart(c *gin.Context) {
	id := c.Param("id")
	ctainr, err := cr.factory.Load(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	_ = ctainr
	// if err := kubernetes.StartContainer(ctainr); err != nil {
	// 	httputil.Error(c, http.StatusInternalServerError, err)
	// 	return
	// }
	c.Writer.WriteHeader(http.StatusNoContent)
}

// DELETE "/containers/:id"
func (cr *containerRouter) ContainerDelete(c *gin.Context) {
	id := c.Param("id")
	ctainr, err := cr.factory.Load(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	_ = ctainr
	// if err := kubernetes.DeleteContainer(ctainr); err != nil {
	// 	httputil.Error(c, http.StatusInternalServerError, err)
	// 	return
	// }
	c.Writer.WriteHeader(http.StatusNoContent)
}

// GET "/containers/:id/json"
func (cr *containerRouter) ContainerInfo(c *gin.Context) {
	id := c.Param("id")
	ctainr, err := cr.factory.Load(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Id": id,
		"Config": gin.H{
			"Image":  ctainr.GetImage(),
			"Labels": ctainr.GetLabels(),
			"Env":    ctainr.GetEnv(),
			"Cmd":    ctainr.GetCmd(),
		},
		"NetworkSettings": gin.H{
			"Ports": gin.H{
				"9000/tcp": []gin.H{
					{
						"HostIp":   "localhost",
						"HostPort": "8080",
					},
				},
			},
		},
		"Image": ctainr.GetImage(),
		"State": gin.H{
			"Health": gin.H{
				"Status": "healthy",
			},
			"Running": true,
			"Status":  "running",
		},
	})
}
