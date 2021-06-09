package routes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/server/httputil"
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

	// hmm... how to do this without a running container...
	if !tainr.Running && !tainr.Completed {
		if err := cr.startContainer(tainr); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
	}

	archive, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	if err := cr.kub.CopyToContainer(tainr, archive, path); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "archive copied succesfully to container",
	})
}
