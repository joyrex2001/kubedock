package common

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// ImageList - list Images. Stubbed, not relevant on k8s.
// https://docs.docker.com/engine/api/v1.41/#operation/ImageList
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/images/operation/ImageListLibpod
// GET "/images/json"
// GET "/libpod/images/json"
func ImageList(cr *ContextRouter, c *gin.Context) {
	imgs, err := cr.DB.GetImages()
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
func ImageJSON(cr *ContextRouter, c *gin.Context) {
	id := strings.TrimSuffix(c.Param("image")+c.Param("json"), "/json")
	img, err := cr.DB.GetImageByNameOrID(id)
	if err != nil {
		img = &types.Image{Name: id}
		if cr.Config.Inspector {
			pts, err := cr.Backend.GetImageExposedPorts(id)
			if err != nil {
				httputil.Error(c, http.StatusInternalServerError, err)
				return
			}
			img.ExposedPorts = pts
		}
		if err := cr.DB.SaveImage(img); err != nil {
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
