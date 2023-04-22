package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/config"
)

// Info - get system information.
// https://docs.docker.com/engine/api/v1.41/#operation/SystemInfo
// GET "/info"
func (cr *Router) Info(c *gin.Context) {
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
func (cr *Router) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Version":    config.DockerVersion,
		"ApiVersion": config.DockerAPIVersion,
		"GitCommit":  config.Build,
		"BuildTime":  config.Date,
		"GoVersion":  config.GoVersion,
		"Os":         config.GOOS,
		"Arch":       config.GOARCH,
	})
}

// Ping - dummy endpoint you can use to test if the server is accessible.
// https://docs.docker.com/engine/api/v1.41/#operation/SystemPing
// HEAD "/_ping"
// GET "/_ping"
func (cr *Router) Ping(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// Events - Stream real-time events from the server.
// https://docs.docker.com/engine/api/v1.41/#tag/System/operation/SystemEvents
// GET "/events"
func (cr *Router) Events(c *gin.Context) {
	w := c.Writer
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Flush()
	enc := json.NewEncoder(w)
	el, id := cr.events.Subscribe()
	for {
		select {
		case <-c.Request.Context().Done():
			cr.events.Unsubscribe(id)
			return
		case msg := <-el:
			// TODO: add (id) filter support
			klog.Info("sending message")
			enc.Encode(gin.H{
				"id":     msg.ID,
				"Type":   msg.Type,
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
