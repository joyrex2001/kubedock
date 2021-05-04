package routes

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// PUT "/containers/:id/archive"
func (cr *Router) PutArchive(c *gin.Context) {
	// TODO: implement noOverwriteDirNonDir
	// TODO: implement copyUIDGID
	id := c.Param("id")
	path := c.Query("path")

	if path == "" {
		httputil.Error(c, http.StatusBadRequest, fmt.Errorf("missing required path parameter"))
		return
	}

	tainr, err := cr.db.LoadContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	// hmm... how to do this without a running container...
	running, _ := cr.kubernetes.IsContainerRunning(tainr)
	if !running {
		if err := cr.kubernetes.StartContainer(tainr); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
	}

	archive, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	if err := cr.kubernetes.CopyToContainer(tainr, archive, path); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "archive copied succesfully to container",
	})
}
