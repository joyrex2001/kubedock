package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/container"
	"github.com/joyrex2001/kubedock/internal/kubernetes"
)

type ContainerCreateRequest struct {
	Name         string                 `json:"name"`
	Image        string                 `json:"image"`
	ExposedPorts map[string]interface{} `json:"ExposedPorts"`
	Labels       map[string]string      `json:"Labels"`
}

// POST "/containers/create"
func ContainerCreate(c *gin.Context) {
	in := &ContainerCreateRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		Error(c, http.StatusInternalServerError, err)
		return
	}
	ctainr := container.New(in.Name, in.Image, in.ExposedPorts, in.Labels)
	c.JSON(http.StatusCreated, gin.H{
		"Id": ctainr.ID,
	})
}

// POST "/containers/:id/start"
func ContainerStart(c *gin.Context) {
	id := c.Param("id")
	ctainr, err := container.Load(id)
	if err != nil {
		Error(c, http.StatusNotFound, err)
		return
	}
	if err := kubernetes.StartContainer(ctainr); err != nil {
		Error(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// POST "/containers/:id/stop"
func ContainerStop(c *gin.Context) {
	id := c.Param("id")
	ctainr, err := container.Load(id)
	if err != nil {
		Error(c, http.StatusNotFound, err)
		return
	}
	if err := kubernetes.StopContainer(ctainr); err != nil {
		Error(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// GET "/containers/:id/json"
func ContainerInfo(c *gin.Context) {
	id := c.Param("id")
	tainr, err := container.Load(id)
	if err != nil {
		Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Id": id,
		"Config": gin.H{
			"Image":  tainr.Image,
			"Labels": tainr.Labels,
		},
		"Image": tainr.Image,
		"State": gin.H{
			"Health": gin.H{
				"Status": "healthy",
			},
			"Running": true,
			"Status":  "running",
		},
	})
}
