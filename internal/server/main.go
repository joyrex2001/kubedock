package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes"
)

// Server is the API server.
type Server struct {
	kub backend.Backend
}

// New will instantiate a Server object.
func New(kub backend.Backend) *Server {
	return &Server{kub: kub}
}

// Run will initialize the http api server and configure all available
// routers.
func (s *Server) Run() error {
	if !klog.V(2) {
		gin.SetMode(gin.ReleaseMode)
	}

	router := s.getGinEngine()

	socket := viper.GetString("server.socket")
	if socket == "" {
		port := viper.GetString("server.listen-addr")
		if viper.GetBool("server.enable-tls") {
			cert := viper.GetString("server.cert-file")
			key := viper.GetString("server.key-file")
			router.RunTLS(port, cert, key)
		} else {
			router.Run(port)
		}
	} else {
		router.RunUnix(socket)
	}

	return nil
}

// getGinEngine will return a gin.Engine router and configure the
// appropriate middleware.
func (s *Server) getGinEngine() *gin.Engine {
	router := gin.New()
	router.Use(httputil.VersionAliasMiddleware(router))
	router.Use(gin.Logger())
	router.Use(httputil.RequestLoggerMiddleware())
	router.Use(httputil.ResponseLoggerMiddleware())
	router.Use(gin.Recovery())
	routes.New(router, s.kub)
	return router
}
