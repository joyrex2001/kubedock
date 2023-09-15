package docker

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/events"
	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes/common"
)

// ImageCreate - create an image.
// https://docs.docker.com/engine/api/v1.41/#operation/ImageCreate
// POST "/images/create"
func ImageCreate(cr *common.ContextRouter, c *gin.Context) {
	from := c.Query("fromImage")
	tag := c.Query("tag")
	if tag != "" {
		from = from + ":" + tag
	}
	img := &types.Image{Name: from}
	if cr.Config.Inspector {
		pts, err := cr.Backend.GetImageExposedPorts(from)
		if err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
		img.ExposedPorts = pts
	}
	if err := cr.DB.SaveImage(img); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	cr.Events.Publish(from, events.Image, events.Pull)

	c.JSON(http.StatusOK, gin.H{
		"status": "Download complete",
	})
}

// ImagesPrune - Delete unused images.
// https://docs.docker.com/engine/api/v1.41/#operation/ImagePrune
// POST "/images/prune"
func ImagesPrune(cr *common.ContextRouter, c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"ImagesDeleted":  []string{},
		"SpaceReclaimed": 0,
	})
}
