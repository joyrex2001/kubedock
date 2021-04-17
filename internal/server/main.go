package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/joyrex2001/kubedock/internal/container"
	"github.com/joyrex2001/kubedock/internal/kubernetes"
	routes_container "github.com/joyrex2001/kubedock/internal/server/routes/container"
	routes_images "github.com/joyrex2001/kubedock/internal/server/routes/images"
	routes_system "github.com/joyrex2001/kubedock/internal/server/routes/system"
	"github.com/joyrex2001/kubedock/internal/util/keyval"
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
func (s *Server) Run(port string) error {
	if !viper.GetBool("generic.verbose") {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	kv, err := keyval.New()
	if err != nil {
		return err
	}

	cf := container.NewFactory(kv)
	kube := kubernetes.New()

	routes_container.New(router, cf, kube)
	routes_system.New(router)
	routes_images.New(router)

	router.Run(port)

	return nil
}
