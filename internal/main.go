package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/joyrex2001/kubedock/internal/routes"
)

// Main is the main entry point for starting this service, based the settings
// initiated by cmd.
func Main(cmd *cobra.Command, args []string) {
	// https://docs.docker.com/engine/api/v1.18/
	// https://docs.docker.com/engine/api/v1.41/
	// https://github.com/moby/moby

	if !viper.GetBool("generic.verbose") {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.GET("/info", routes.Info)
	router.GET("/version", routes.Version)
	router.GET("/healthz", routes.Healthz)
	router.GET("/images/json", routes.ImageList)
	router.POST("/images/create", routes.ImageCreate)
	router.GET("/images/:image/*json", routes.ImageJson)
	router.POST("/containers/create", routes.ContainerCreate)
	router.POST("/containers/:id/start", routes.ContainerStart)
	router.GET("/containers/:id/logs", NotImplemented)
	router.GET("/containers/:id/json", routes.ContainerInfo)
	router.POST("/containers/:id/stop", NotImplemented)
	router.POST("/containers/:id/kill", NotImplemented)
	router.DELETE("/containers/:id", routes.ContainerDelete)
	router.POST("/containers/:id/exec", routes.ContainerExec)
	router.POST("/exec/:id/start", routes.ExecStart)
	router.GET("/exec/:id/json", routes.ExecInfo)

	router.GET("/containers/json", NotImplemented)
	router.GET("/containers/:id/top", NotImplemented)
	router.GET("/containers/:id/changes", NotImplemented)
	router.GET("/containers/:id/export", NotImplemented)
	router.GET("/containers/:id/stats", NotImplemented)
	router.POST("/containers/:id/resize", NotImplemented)
	router.POST("/containers/:id/restart", NotImplemented)
	router.POST("/containers/:id/update", NotImplemented)
	router.POST("/containers/:id/rename", NotImplemented)
	router.POST("/containers/:id/pause", NotImplemented)
	router.POST("/containers/:id/unpause", NotImplemented)
	router.POST("/containers/:id/attach", NotImplemented)
	router.GET("/containers/:id/attach/ws", NotImplemented)
	router.POST("/containers/:id/wait", NotImplemented)
	router.HEAD("/containers/:id/archive", NotImplemented)
	router.GET("/containers/:id/archive", NotImplemented)
	router.PUT("/containers/:id/archive", NotImplemented)
	router.POST("/containers/prune", NotImplemented)

	router.Run(viper.GetString("server.listen-addr"))
}

// NotImplemented will return a not implented response.
func NotImplemented(c *gin.Context) {
	c.Writer.WriteHeader(http.StatusNotImplemented)
}
