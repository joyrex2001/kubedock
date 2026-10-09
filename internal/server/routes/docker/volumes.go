package docker

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/filter"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes/common"
)

// VolumesList - list volumes.
// https://docs.docker.com/engine/api/v1.44/#tag/Volume/operation/VolumeList
// GET "/volumes"
func VolumesList(cr *common.ContextRouter, c *gin.Context) {
	vols, err := cr.DB.GetVolumes()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	filtr, err := filter.New(c.Query("filters"))
	if err != nil {
		klog.V(5).Infof("unsupported filter: %s", err)
	}
	res := []gin.H{}
	for _, vol := range vols {
		if filtr.Match(vol) {
			res = append(res, getVolumeResponse(vol))
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"Volumes":  res,
		"Warnings": []string{},
	})
}

// VolumesInfo - inspect a volume.
// https://docs.docker.com/engine/api/v1.44/#tag/Volume/operation/VolumeInspect
// GET "/volumes/:name"
func VolumesInfo(cr *common.ContextRouter, c *gin.Context) {
	vol, err := cr.DB.GetVolume(c.Param("name"))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, getVolumeResponse(vol))
}

// VolumesCreate - create a volume. Creating a volume that already exists
// returns the existing volume, as docker does.
// https://docs.docker.com/engine/api/v1.44/#tag/Volume/operation/VolumeCreate
// POST "/volumes/create"
func VolumesCreate(cr *common.ContextRouter, c *gin.Context) {
	in := &VolumeCreateRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		httputil.Error(c, http.StatusBadRequest, err)
		return
	}
	if in.Name != "" {
		if vol, err := cr.DB.GetVolume(in.Name); err == nil {
			c.JSON(http.StatusCreated, getVolumeResponse(vol))
			return
		}
	}
	vol := &types.Volume{
		Name:   in.Name,
		Labels: in.Labels,
	}
	if err := cr.DB.SaveVolume(vol); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, getVolumeResponse(vol))
}

// VolumesDelete - remove a volume.
// https://docs.docker.com/engine/api/v1.44/#tag/Volume/operation/VolumeDelete
// DELETE "/volumes/:name"
func VolumesDelete(cr *common.ContextRouter, c *gin.Context) {
	vol, err := cr.DB.GetVolume(c.Param("name"))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	if err := cr.DB.DeleteVolume(vol); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// VolumesPrune - Delete unused volumes. Volumes are never backed by
// storage, so every volume matching the filters counts as unused.
// https://docs.docker.com/engine/api/v1.44/#tag/Volume/operation/VolumePrune
// POST "/volumes/prune"
func VolumesPrune(cr *common.ContextRouter, c *gin.Context) {
	vols, err := cr.DB.GetVolumes()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	filtr, err := filter.New(c.Query("filters"))
	if err != nil {
		klog.V(5).Infof("unsupported filter: %s", err)
	}
	deleted := []string{}
	for _, vol := range vols {
		if !filtr.Match(vol) {
			continue
		}
		if err := cr.DB.DeleteVolume(vol); err != nil {
			klog.Warningf("error deleting volume %s: %s", vol.Name, err)
			continue
		}
		deleted = append(deleted, vol.Name)
	}
	c.JSON(http.StatusOK, gin.H{
		"VolumesDeleted": deleted,
		"SpaceReclaimed": 0,
	})
}

// getVolumeResponse will return the docker api representation of a volume.
func getVolumeResponse(vol *types.Volume) gin.H {
	labels := vol.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	return gin.H{
		"Name":       vol.Name,
		"Driver":     "local",
		"Mountpoint": "",
		"CreatedAt":  vol.Created.Format(time.RFC3339),
		"Labels":     labels,
		"Scope":      "local",
		"Options":    map[string]string{},
	}
}
