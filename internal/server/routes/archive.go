package routes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/util/tar"
)

// PutArchive - extract an archive of files or folders to a directory in a container.
// https://docs.docker.com/engine/api/v1.41/#operation/PutContainerArchive
// PUT "/containers/:id/archive"
func (cr *Router) PutArchive(c *gin.Context) {
	id := c.Param("id")

	path := c.Query("path")
	if path == "" {
		httputil.Error(c, http.StatusBadRequest, fmt.Errorf("missing required path parameter"))
		return
	}

	ovw, _ := strconv.ParseBool(c.Query("noOverwriteDirNonDir"))
	if ovw {
		klog.Warning("noOverwriteDirNonDir is not supported, ignoring setting.")
	}

	cgid, _ := strconv.ParseBool(c.Query("copyUIDGID"))
	if cgid {
		klog.Warning("copyUIDGID is not supported, ignoring setting.")
	}

	tainr, err := cr.db.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	archive, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	if !tainr.Running && !tainr.Completed && cr.cfg.PreArchive && tar.IsSingleFileArchive(&archive) {
		tainr.PreArchives = append(tainr.PreArchives, types.PreArchive{Path: path, Archive: &archive})
		klog.V(2).Infof("adding prearchive: %v", tainr.PreArchives)
		if err := cr.db.SaveContainer(tainr); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "planned archive to be copied to container",
		})
		return
	}

	if !tainr.Running && !tainr.Completed && !cr.cfg.PreArchive {
		if err := cr.startContainer(tainr); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
	}

	if err := cr.kub.CopyToContainer(tainr, archive, path); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "archive copied succesfully to container",
	})
}

// GetArchive - get a tar archive of a resource in the filesystem of container id.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerArchive
// GET "/containers/:id/archive"
func (cr *Router) GetArchive(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.db.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	path := c.Query("path")
	if path == "" {
		httputil.Error(c, http.StatusBadRequest, fmt.Errorf("missing required path parameter"))
		return
	}

	dat, err := cr.kub.CopyFromContainer(tainr, path)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Header().Set("Content-Type", "application/x-tar")
	// {"name": "","size": 0,"mode": 0,"mtime": "2021-01-01T20:00:00Z","linkTarget": ""}
	c.Writer.Header().Set("X-Docker-Container-Path-Stat", "eyJuYW1lIjogIiIsInNpemUiOiAwLCJtb2RlIjogMCwibXRpbWUiOiAiMjAyMS0wMS0wMVQyMDowMDowMFoiLCJsaW5rVGFyZ2V0IjogIiJ9Cg==")
	c.Writer.Write(dat)
}
