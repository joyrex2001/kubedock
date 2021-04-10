package routes

import (
	"encoding/json"
	"log"
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

type ContainerExecRequest struct {
	Cmd []string `json:"Cmd"`
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
		"NetworkSettings": gin.H{
			"Ports": gin.H{
				"9000/tcp": []gin.H{
					{
						"HostIp":   "127.0.0.1",
						"HostPort": "55000",
					},
				},
			},
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

// POST "/containers/:id/start"
func ContainerExec(c *gin.Context) {
	in := &ContainerExecRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		Error(c, http.StatusInternalServerError, err)
		return
	}
	id := c.Param("id")
	ctainr, err := container.Load(id)
	if err != nil {
		Error(c, http.StatusNotFound, err)
		return
	}
	log.Printf("cmd = %v", in.Cmd)
	if err := kubernetes.StartContainer(ctainr); err != nil {
		Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"Id": ctainr.ID,
	})
}
