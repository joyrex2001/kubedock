package routes

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ContainerCreate(c *gin.Context) {
	in, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Error(c, err)
	}
	log.Print(string(in))
	c.JSON(http.StatusCreated, gin.H{
		"Id": "1234-5678",
	})
}

func ContainerStart(c *gin.Context) {
	// id := ps.ByName("id")
	// log.Print(string(id))
	c.Writer.WriteHeader(http.StatusNoContent)
}
