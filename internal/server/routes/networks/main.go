package networks

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// imagesRouter is the object that facilitate all image API endpoints.
type networksRouter struct {
}

// New will instantiate a imagesRouter object.
func New(version int, router *gin.Engine) *networksRouter {
	vprefix := ""
	if version != 0 {
		vprefix = fmt.Sprintf("/v1.%d", version)
	}
	ir := &networksRouter{}
	ir.initRoutes(vprefix, router)
	return ir
}

// initRoutes will add all suported routes.
func (nr *networksRouter) initRoutes(version string, router *gin.Engine) {
	router.GET(version+"/networks", nr.NetworksList)
	router.POST(version+"/networks/create", nr.NetworksCreate)
	router.GET(version+"/networks/reaper_default", httputil.NotImplemented)
}
