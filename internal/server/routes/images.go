package routes

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

// ImageJSON - return low-level information about an image.
// https://docs.docker.com/engine/api/v1.41/#operation/ImageInspect
// GET "/images/:image/json"
func (cr *Router) ImageJSON(c *gin.Context) {
	id := strings.TrimSuffix(c.Param("image")+c.Param("json"), "/json")
	img, err := cr.db.GetImageByNameOrID(id)
	if err != nil {
		img = &types.Image{Name: id}
		if cr.cfg.Inspector {
			pts, err := cr.kub.GetImageExposedPorts(id)
			if err != nil {
				httputil.Error(c, http.StatusInternalServerError, err)
				return
			}
			img.ExposedPorts = pts
		}
		if err := cr.db.SaveImage(img); err != nil {
			httputil.Error(c, http.StatusNotFound, err)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"Id":      img.Name,
		"Created": img.Created.Format("2006-01-02T15:04:05Z"),
		"Size":    0,
		"ContainerConfig": gin.H{
			"Image": img.Name,
		},
	})
}

// ImageCreate - create an image.
// https://docs.docker.com/engine/api/v1.41/#operation/ImageCreate
// POST "/images/create"
func (cr *Router) ImageCreate(c *gin.Context) {
	from := c.Query("fromImage")
	tag := c.Query("tag")
	if tag != "" {
		from = from + ":" + tag
	}
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
		"status": "Download complete",
	})
}
