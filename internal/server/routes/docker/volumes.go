package docker

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/server/routes/common"
)

// VolumesPrune - Delete unused volumes.
// https://docs.docker.com/engine/api/v1.41/#operation/VolumePrune
// POST "/volumes/prune"
func VolumesPrune(cr *common.ContextRouter, c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"VolumesDeleted": []string{},
		"SpaceReclaimed": 0,
	})
}
