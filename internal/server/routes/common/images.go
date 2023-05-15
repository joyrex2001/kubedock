package common

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

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
