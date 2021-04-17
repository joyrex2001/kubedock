package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/joyrex2001/kubedock/internal/container"
	rcont "github.com/joyrex2001/kubedock/internal/server/routes/container"
	"github.com/joyrex2001/kubedock/internal/server/routes/images"
	"github.com/joyrex2001/kubedock/internal/server/routes/system"
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

	system.New(router)
	images.New(router)
	rcont.New(router, cf)

	router.Run(port)

	return nil
}
