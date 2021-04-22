package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	kubecli "k8s.io/client-go/kubernetes"

	"github.com/joyrex2001/kubedock/internal/config"
	"github.com/joyrex2001/kubedock/internal/container"
	"github.com/joyrex2001/kubedock/internal/kubernetes"
	routes_container "github.com/joyrex2001/kubedock/internal/server/routes/container"
	routes_images "github.com/joyrex2001/kubedock/internal/server/routes/images"
	routes_networks "github.com/joyrex2001/kubedock/internal/server/routes/networks"
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

	cfg, err := config.GetKubernetes()
	if err != nil {
		return err
	}

	cli, err := kubecli.NewForConfig(cfg)
	if err != nil {
		return err
	}

	kube := kubernetes.New(cfg, cli, viper.GetString("kubernetes.namespace"))
	for v := 0; v <= config.DockerMaxAPIMinor; v++ {
		routes_container.New(v, router, cf, kube)
		routes_system.New(v, router)
		routes_images.New(v, router)
		routes_networks.New(v, router)
	}

	router.Run(port)

	return nil
}
