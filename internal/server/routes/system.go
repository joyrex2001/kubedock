package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/config"
)

// GET "/healthz"
func (cr *Router) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"timestamp": time.Now().Unix(),
	})
}

// GET "/info"
func (cr *Router) Info(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ID":              config.ID,
		"Name":            config.Name,
		"ServerVersion":   config.Version,
		"OperatingSystem": config.OS,
		"MemTotal":        0,
	})
}

// GET "/version"
func (cr *Router) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Version":    config.DockerVersion,
		"ApiVersion": config.DockerAPIVersion,
		"GitCommit":  config.Build,
		"BuildTime":  config.Date,
		"GoVersion":  config.GoVersion,
		"Os":         config.GOOS,
		"Arch":       config.GOARCH,
	})
}

// GET "/_ping"
func (cr *Router) Ping(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
