package routes

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// GET "/healthz"
func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"timestamp": time.Now().Unix(),
	})
}

// GET "/info"
func Info(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ID":              "donk-909",
		"Name":            "donk",
		"ServerVersion":   "donk-909",
		"OperatingSystem": "kubernetes",
		"MemTotal":        0,
	})
}

// GET "/version"
func Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Version":    "1.0",
		"ApiVersion": "1.0",
		"GoVersion":  runtime.Version(),
		"Os":         runtime.GOOS,
		"Arch":       runtime.GOARCH,
	})
}
