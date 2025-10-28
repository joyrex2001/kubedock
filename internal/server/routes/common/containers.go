package common

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/events"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// ContainerStart - start a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerStart
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerStartLibpod
// POST "/containers/:id/start"
// POST "/libpod/containers/:id/start"
func ContainerStart(cr *ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	if !tainr.Running && !tainr.Completed {
		if err := StartContainer(cr, tainr); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
	} else {
		klog.Warningf("container %s already running", id)
	}

	cr.Events.Publish(tainr.ID, events.Container, events.Start)

	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerRestart - restart a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerRestart
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerRestartLibpod
// POST "/containers/:id/restart"
// POST "/libpod/containers/:id/restart"
func ContainerRestart(cr *ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	ts := c.Query("t")
	t, _ := strconv.Atoi(ts)
	if t > 0 {
		time.Sleep(time.Duration(t) * time.Second)
	}

	deleted, err := cr.Backend.WatchDeleteContainer(tainr)
	if err != nil {
		klog.Warningf("error while watching k8s container delete: %s", err)
	}

	if err := cr.Backend.DeleteContainer(tainr); err != nil {
		klog.Warningf("error while deleting k8s container: %s", err)
	}
	tainr.SignalDetach()
	tainr.SignalStop()

	tainr.Running = false
	tainr.Completed = false
	tainr.Stopped = true

	if err := cr.DB.SaveContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	<-deleted

	if err := StartContainer(cr, tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerStop - stop a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerStop
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerStopLibpod
// POST "/containers/:id/stop"
// POST "/libpod/containers/:id/stop"
func ContainerStop(cr *ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	tainr.SignalDetach()
	tainr.SignalStop()

	if !tainr.Stopped && !tainr.Killed {
		if err := cr.Backend.DeleteContainer(tainr); err != nil {
			klog.Warningf("error while deleting k8s container: %s", err)
		}
	}

	tainr.Running = false
	tainr.Completed = false
	tainr.Stopped = true

	if err := cr.DB.SaveContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	cr.Events.Publish(tainr.ID, events.Container, events.Die)

	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerKill - kill a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerKill
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerKillLibpod
// POST "/containers/:id/kill"
// POST "/libpod/containers/:id/kill"
func ContainerKill(cr *ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainerByNameOrID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	signal := strings.ToLower(c.Query("signal"))

	valid := map[string]bool{
		"kil":  true,
		"term": true,
		"quit": true,
		"int":  true,
	}

	if signal != "" && !valid[signal] {
		klog.Infof("ignoring signal %s", signal)
		c.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	tainr.SignalDetach()
	tainr.SignalStop()

	if !tainr.Stopped && !tainr.Killed {
		if err := cr.Backend.DeleteContainer(tainr); err != nil {
			klog.Warningf("error while deleting k8s container: %s", err)
		}
	}

	tainr.Killed = true
	tainr.Running = false
	tainr.Completed = false

	if err := cr.DB.SaveContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	cr.Events.Publish(tainr.ID, events.Container, events.Die)

	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerAttach - attach to a container to read its output or send input.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerAttach
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerAttachLibpod
// POST "/containers/:id/attach"
// POST "/libpod/containers/:id/attach"
func ContainerAttach(cr *ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainerByNameOrID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	// Fallback on settings when container is created
	stdinParam, _ := strconv.ParseBool(c.Query("stdin"))
	stdin := tainr.OpenStdin || stdinParam
	stdout, _ := strconv.ParseBool(c.Query("stdout"))
	stderr, _ := strconv.ParseBool(c.Query("stderr"))
	stream, _ := strconv.ParseBool(c.Query("stream"))
	// TTY is not a query param available in the containerAttachRequest so it is retrieved from the containerCreate req
	tty := tainr.Tty

	if !stream {
		c.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	if !tainr.Running && !tainr.Completed {
		if err := StartContainer(cr, tainr); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
	}

	r := c.Request
	w := c.Writer
	w.WriteHeader(http.StatusOK)

	hj, ok := w.(http.Hijacker)
	if !ok {
		httputil.Error(c, http.StatusInternalServerError, fmt.Errorf("hijacking not supported"))
		return
	}

	conn, _, err := hj.Hijack()
	if err != nil {
		klog.Errorf("error during hijack connection: %s", err)
		return
	}

	// Now conn is both your reader and writer
	in := conn
	out := conn

	defer httputil.CloseStreams(in, out)
	httputil.UpgradeConnection(r, out)

	stop := make(chan struct{}, 1)
	tainr.AddAttachChannel(stop)
	attachDone := make(chan struct{}, 1)

	// Start streaming to/from the container
	go func() {
		defer close(attachDone)
		err := cr.Backend.AttachContainer(
			tainr,
			func() io.Reader {
				if stdin {
					return in
				}
				return nil
			}(),
			func() io.Writer {
				if stdout {
					return out
				}
				return nil
			}(),
			func() io.Writer {
				if stderr {
					// Docker expects stderr merged if TTY is enabled
					if tty {
						return out
					}
					return out // or multiplex if you implement it separately
				}
				return nil
			}(),
			tty,
		)
		if err != nil {
			klog.Errorf("attach error: %v", err)
		}
	}()

	// Wait until container detach or attach completes
	select {
	case <-stop:
		klog.Infof("detach signal received for container %s", tainr.ID)
	case <-attachDone:
		klog.Infof("attach session finished for container %s", tainr.ID)
	}

	tainr.SignalDetach()
	// Cleanup and notify events
	cr.Events.Publish(tainr.ID, events.Container, events.Detach)
}

// ContainerResize - resize the tty for a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerResize
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerResizeLibpod
// POST "/containers/:id/rezise"
// POST "/libpod/containers/:id/rezise"
func ContainerResize(cr *ContextRouter, c *gin.Context) {
	id := c.Param("id")
	_, err := cr.DB.GetContainerByNameOrID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
	return
}

// ContainerRename - rename a container.
// https://docs.docker.com/engine/api/v1.41/#tag/Container/operation/ContainerRename
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerRenameLibpod
// GET "/containers/:id/rename"
// GET "/libpod/containers/:id/rename"
func ContainerRename(cr *ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainerByNameOrID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	name := c.Query("name")
	if _, err := cr.DB.GetContainerByName(name); err == nil {
		httputil.Error(c, http.StatusConflict, fmt.Errorf("name `%s` already in used", name))
		return
	}
	tainr.Name = name
	if err := cr.DB.SaveContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}
