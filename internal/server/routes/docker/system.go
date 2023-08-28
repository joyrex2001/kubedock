package docker

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/config"
	"github.com/joyrex2001/kubedock/internal/server/filter"
	"github.com/joyrex2001/kubedock/internal/server/routes/common"
)

// Info - get system information.
// https://docs.docker.com/engine/api/v1.41/#operation/SystemInfo
// GET "/info"
func Info(cr *common.ContextRouter, c *gin.Context) {
	labels := []string{}
	for k, v := range config.DefaultLabels {
		labels = append(labels, k+"="+v)
	}
	c.JSON(http.StatusOK, gin.H{
		"ID":              config.ID,
		"Name":            config.Name,
		"ServerVersion":   config.Version,
		"OperatingSystem": config.OS,
		"MemTotal":        0,
		"Labels":          labels,
	})
}

// Version - get version.
// https://docs.docker.com/engine/api/v1.41/#operation/SystemVersion
// GET "/version"
func Version(cr *common.ContextRouter, c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Version":       config.DockerVersion,
		"ApiVersion":    config.DockerAPIVersion,
		"MinAPIVersion": config.DockerAPIVersion,
		"GitCommit":     config.Build,
		"BuildTime":     config.Date,
		"GoVersion":     config.GoVersion,
		"Os":            config.GOOS,
		"Arch":          config.GOARCH,
	})
}

// Ping - dummy endpoint you can use to test if the server is accessible.
// https://docs.docker.com/engine/api/v1.41/#operation/SystemPing
// HEAD "/_ping"
// GET "/_ping"
func Ping(cr *common.ContextRouter, c *gin.Context) {
	w := c.Writer
	w.Header().Set("API-Version", config.DockerAPIVersion)
	c.String(http.StatusOK, "OK")
}

// Events - Stream real-time events from the server.
// https://docs.docker.com/engine/api/v1.41/#tag/System/operation/SystemEvents
// GET "/events"
func Events(cr *common.ContextRouter, c *gin.Context) {
	w := c.Writer
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Flush()

	filtr, err := filter.New(c.Query("filters"))
	if err != nil {
		klog.V(5).Infof("unsupported filter: %s", err)
	}

	enc := json.NewEncoder(w)
	el, id := cr.Events.Subscribe()
	for {
		select {
		case <-c.Request.Context().Done():
			cr.Events.Unsubscribe(id)
			return
		case msg := <-el:
			if filtr.Match(&msg) {
				klog.V(5).Infof("sending message to %s", id)
				enc.Encode(gin.H{
					"id":     msg.ID,
					"Type":   msg.Type,
					"Status": msg.Action,
					"Action": msg.Action,
					"Actor": gin.H{
						"ID": msg.ID,
					},
					"scope":    "local",
					"time":     msg.Time,
					"timeNano": msg.TimeNano,
				})
				w.Flush()
			}
		}
	}
}
