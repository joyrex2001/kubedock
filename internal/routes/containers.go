package routes

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/donk/internal/container"
	"github.com/joyrex2001/donk/internal/kubernetes"
)

// POST "/containers/create"
func ContainerCreate(c *gin.Context) {
	in, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Error(c, err)
		return
	}
	log.Print(string(in))
	tainr := container.New()
	c.JSON(http.StatusCreated, gin.H{
		"Id": tainr.ID,
	})
}

// POST "/containers/:id/start"
func ContainerStart(c *gin.Context) {
	id := c.Param("id")
	tainr, err := container.Load(id)
	if err != nil {
		Error(c, err)
		return
	}
	if err := kubernetes.StartContainer(tainr); err != nil {
		Error(c, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// POST "/containers/:id/stop"
func ContainerStop(c *gin.Context) {
	id := c.Param("id")
	tainr, err := container.Load(id)
	if err != nil {
		Error(c, err)
		return
	}
	if err := kubernetes.StopContainer(tainr); err != nil {
		Error(c, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// GET "/containers/:id/json"
func ContainerInfo(c *gin.Context) {
	id := c.Param("id")
	tainr, err := container.Load(id)
	if err != nil {
		Error(c, err)
		return
	}
	_ = tainr
	c.Writer.WriteHeader(http.StatusNoContent)
}
