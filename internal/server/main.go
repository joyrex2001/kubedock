package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/joyrex2001/kubedock/internal/server/routes/container"
	"github.com/joyrex2001/kubedock/internal/server/routes/images"
	"github.com/joyrex2001/kubedock/internal/server/routes/system"
)

// Server is the API server.
type Server struct {
}

// New will instantiate a Server object.
func New() *Server {
	return &Server{}
}

// Run will initialize the http api server and configure all available
// routers.
func (s *Server) Run(port string) {
	// https://docs.docker.com/engine/api/v1.18/
	// https://docs.docker.com/engine/api/v1.41/
	// https://github.com/moby/moby

	if !viper.GetBool("generic.verbose") {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	system.New(router)
	images.New(router)
	container.New(router)

	router.Run(port)
}
