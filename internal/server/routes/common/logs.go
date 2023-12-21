package common

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// ContainerLogs - get container logs.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerLogs
// POST "/containers/:id/logs"
func ContainerLogs(cr *ContextRouter, c *gin.Context) {
	id := c.Param("id")
	// TODO: implement until

	tainr, err := cr.DB.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	if !tainr.Running && !tainr.Completed {
		httputil.Error(c, http.StatusNotFound, fmt.Errorf("container %s is not running", tainr.ShortID))
		return
	}

	r := c.Request
	w := c.Writer
	w.WriteHeader(http.StatusOK)

	follow, _ := strconv.ParseBool(c.Query("follow"))
	tailLines, _ := parseInt64(c.Query("tail"))
	sinceTime, _ := parseUnix(c.Query("since"))
	timestamps, _ := strconv.ParseBool(c.Query("timestamps"))

	logOpts := backend.LogOptions{
		Follow:     follow,
		SinceTime:  sinceTime,
		Timestamps: timestamps,
		TailLines:  tailLines,
	}

	if !follow {
		stop := make(chan struct{}, 1)
		if err := cr.Backend.GetLogs(tainr, &logOpts, stop, w); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
		close(stop)
		return
	}

	in, out, err := httputil.HijackConnection(w)
	if err != nil {
		klog.Errorf("error during hijack connection: %s", err)
		return
	}
	defer httputil.CloseStreams(in, out)
	httputil.UpgradeConnection(r, out)

	stop := make(chan struct{}, 1)
	tainr.AddStopChannel(stop)

	if err := cr.Backend.GetLogs(tainr, &logOpts, stop, out); err != nil {
		klog.V(3).Infof("error retrieving logs: %s", err)
		return
	}
}

// Parses the input expecting an int64 number as a string.
func parseInt64(input string) (*int64, error) {
	num, err := strconv.ParseInt(input, 10, 32)
	if err != nil {
		return nil, err
	}
	return &num, nil
}

// Parses the input expecting a string representing number of seconds since the Epoch.
func parseUnix(input string) (*time.Time, error) {
	num, err := strconv.ParseInt(input, 10, 32)
	if err != nil {
		return nil, err
	}
	result := time.Unix(num, 0)
	return &result, nil
}
