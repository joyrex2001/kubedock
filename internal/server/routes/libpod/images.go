package libpod

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/events"
	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// ImageList - list Images. Stubbed, not relevant on k8s.
// https://docs.docker.com/engine/api/v1.41/#operation/ImageList
// GET "/images/json"
func (cr *Router) ImageList(c *gin.Context) {
	imgs, err := cr.db.GetImages()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	res := []gin.H{}
	for _, img := range imgs {
		name := img.Name
		if !strings.Contains(name, ":") {
			name = name + ":latest"
		}
		res = append(res, gin.H{"ID": img.ID, "Size": 0, "Created": img.Created.Unix(), "RepoTags": []string{name}})
	}
	c.JSON(http.StatusOK, res)
}

// ImagePull - pull one or more images from a container registry.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/images/operation/ImagePullLibpod
// POST "/libpod/images/pull"
func (cr *Router) ImagePull(c *gin.Context) {
	from := c.Query("reference")
	img := &types.Image{Name: from}
	if cr.cfg.Inspector {
		pts, err := cr.kub.GetImageExposedPorts(from)
		if err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
		img.ExposedPorts = pts
	}
	if err := cr.db.SaveImage(img); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	cr.events.Publish(from, events.Image, events.Pull)

	c.JSON(http.StatusOK, gin.H{
		"Id": img.ID,
	})
}
