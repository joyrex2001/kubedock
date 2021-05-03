package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	kubecli "k8s.io/client-go/kubernetes"

	"github.com/joyrex2001/kubedock/internal/config"
	"github.com/joyrex2001/kubedock/internal/container"
	"github.com/joyrex2001/kubedock/internal/kubernetes"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes"
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

	router := gin.New()
	router.Use(httputil.VersionAliasMiddleware(router))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	if viper.GetBool("generic.logrequest") {
		router.Use(httputil.RequestLoggerMiddleware())
		router.Use(httputil.ResponseLoggerMiddleware())
	}

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

	routes.New(router, cf, kube)
	router.Run(port)

	return nil
}
