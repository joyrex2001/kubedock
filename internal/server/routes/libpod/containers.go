package libpod

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/events"
	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/filter"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes"
)

// ContainerCreate - create a container.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerCreateLibpod
// POST "/libpod/containers/create"
func ContainerCreate(cr *routes.ContextRouter, c *gin.Context) {
	in := &ContainerCreateRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	if in.Name == "" {
		in.Name = c.Query("name")
	}

	if in.Labels == nil {
		in.Labels = map[string]string{}
	}

	if in.User == "" && cr.Config.RunasUser != "" {
		in.User = cr.Config.RunasUser
	}

	if _, ok := in.Labels[types.LabelRequestCPU]; !ok && cr.Config.RequestCPU != "" {
		in.Labels[types.LabelRequestCPU] = cr.Config.RequestCPU
	}
	if _, ok := in.Labels[types.LabelRequestMemory]; !ok && cr.Config.RequestMemory != "" {
		in.Labels[types.LabelRequestMemory] = cr.Config.RequestMemory
	}
	if _, ok := in.Labels[types.LabelPullPolicy]; !ok && cr.Config.PullPolicy != "" {
		in.Labels[types.LabelPullPolicy] = cr.Config.PullPolicy
	}
	if _, ok := in.Labels[types.LabelDeployAsJob]; !ok && cr.Config.DeployAsJob {
		in.Labels[types.LabelDeployAsJob] = "true"
	}
	in.Labels[types.LabelServiceAccount] = cr.Config.ServiceAccount

	tainr := &types.Container{
		Name:         in.Name,
		Image:        in.Image,
		Entrypoint:   in.Entrypoint,
		User:         in.User,
		Cmd:          in.Command,
		Env:          in.Env,
		ExposedPorts: map[string]interface{}{},
		ImagePorts:   map[string]interface{}{},
		Labels:       in.Labels,
	}

	if img, err := cr.DB.GetImageByNameOrID(in.Image); err != nil {
		klog.Warningf("unable to fetch image details: %s", err)
	} else {
		for pp := range img.ExposedPorts {
			tainr.ImagePorts[pp] = pp
		}
	}

	for _, mapping := range in.PortMappings {
		src := fmt.Sprintf("%d", mapping.HostPort)
		dst := fmt.Sprintf("%d", mapping.ContainerPort)
		if err := tainr.AddHostPort(src, dst); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
		tainr.ExposedPorts[dst] = src
	}

	addNetworkAliases(tainr, in.Network)

	netw, err := cr.DB.GetNetworkByName("bridge")
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	tainr.ConnectNetwork(netw.ID)

	if err := cr.DB.SaveContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	cr.Events.Publish(tainr.ID, events.Container, events.Create)

	c.JSON(http.StatusCreated, gin.H{
		"Id": tainr.ID,
	})
}

// ContainerStart - start a container.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerStartLibpod
// POST "/libpod/containers/:id/start"
func ContainerStart(cr *routes.ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	if !tainr.Running && !tainr.Completed {
		if err := startContainer(cr, tainr); err != nil {
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
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerRestartLibpod
// POST "/libpod/containers/:id/restart"
func ContainerRestart(cr *routes.ContextRouter, c *gin.Context) {
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

	time.Sleep(time.Second)
	if err := startContainer(cr, tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerStop - stop a container.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerStopLibpod
// POST "/libpod/containers/:id/stop"
func ContainerStop(cr *routes.ContextRouter, c *gin.Context) {
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
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerKillLibpod
// POST "/libpod/containers/:id/kill"
func ContainerKill(cr *routes.ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	signal := strings.ToLower(c.Query("signal"))
	if strings.Contains(signal, "int") {
		tainr.SignalDetach()
		if err := cr.DB.SaveContainer(tainr); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
		c.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	if signal != "" && !strings.Contains(signal, "kil") && !strings.Contains(signal, "term") && !strings.Contains(signal, "quit") {
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

// ContainerDelete - remove a container.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerDeleteLibpod
// DELETE "/libpod/containers/:id"
func ContainerDelete(cr *routes.ContextRouter, c *gin.Context) {
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
		cr.Events.Publish(tainr.ID, events.Container, events.Die)
	}

	if err := cr.DB.DeleteContainer(tainr); err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerAttach - attach to a container to read its output or send input.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerAttachLibpod
// POST "/libpod/containers/:id/attach"
func ContainerAttach(cr *routes.ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	stdin, _ := strconv.ParseBool(c.Query("stdin"))
	if stdin {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	}
	stdout, _ := strconv.ParseBool(c.Query("stdout"))
	stderr, _ := strconv.ParseBool(c.Query("stderr"))
	if !stdout || !stderr {
		klog.Warningf("Ignoring stdout/stderr filtering")
	}

	if !tainr.Running && !tainr.Completed {
		if err := startContainer(cr, tainr); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
	}

	stream, _ := strconv.ParseBool(c.Query("stream"))
	if !stream {
		c.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	r := c.Request
	w := c.Writer
	w.WriteHeader(http.StatusOK)

	in, out, err := httputil.HijackConnection(w)
	if err != nil {
		klog.Errorf("error during hijack connection: %s", err)
		return
	}
	defer httputil.CloseStreams(in, out)
	httputil.UpgradeConnection(r, out)

	stop := make(chan struct{}, 1)
	tainr.AddAttachChannel(stop)

	if err := cr.Backend.GetLogs(tainr, true, 100, stop, out); err != nil {
		klog.Errorf("error retrieving logs: %s", err)
		return
	}
}

// ContainerWait - Block until a container stops, then returns the exit code.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerWaitLibpod
// POST "/libpod/containers/:id/wait"
func ContainerWait(cr *routes.ContextRouter, c *gin.Context) {
	id := c.Param("id")
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		tainr, err := cr.DB.GetContainer(id)
		updateContainerStatus(cr, tainr)
		if err != nil || tainr.Stopped || tainr.Killed || tainr.Completed {
			c.JSON(http.StatusOK, gin.H{"StatusCode": 0})
			return
		}
	}
}

// ContainerInfo - return low-level information about a container.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerInspectLibpod
// GET "/libpod/containers/:id/json"
func ContainerInfo(cr *routes.ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, getContainerInfo(cr, tainr, true))
}

// ContainerList - returns a list of containers.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerListLibpod
// GET "/libpod/containers/json"
func ContainerList(cr *routes.ContextRouter, c *gin.Context) {
	filtr, err := filter.New(c.Query("filters"))
	if err != nil {
		klog.V(5).Infof("unsupported filter: %s", err)
	}

	tainrs, err := cr.DB.GetContainers()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	res := []gin.H{}
	for _, tainr := range tainrs {
		if filtr.Match(tainr) {
			res = append(res, getContainerInfo(cr, tainr, false))
		}
	}
	c.JSON(http.StatusOK, res)
}

// ContainerRename - rename a container.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerRenameLibpod
// GET "/libpod/containers/:id/rename"
func ContainerRename(cr *routes.ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainer(id)
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

// getContainerInfo will return a gin.H containing the details of the
// given container.
func getContainerInfo(cr *routes.ContextRouter, tainr *types.Container, detail bool) gin.H {
	errstr := ""
	netws, err := cr.DB.GetNetworksByIDs(tainr.Networks)
	if err != nil {
		errstr += err.Error()
	}
	netdtl := gin.H{}
	for _, netw := range netws {
		netdtl[netw.Name] = gin.H{
			"NetworkID": netw.ID,
			"Aliases":   tainr.NetworkAliases,
			"IPAddress": "127.0.0.1",
		}
	}
	res := gin.H{
		"Id":    tainr.ID,
		"Name":  "/" + tainr.Name,
		"Image": tainr.Image,
		"Names": getContainerNames(tainr),
	}
	updateContainerStatus(cr, tainr)
	if detail {
		res["State"] = gin.H{
			"Health": gin.H{
				"Status": tainr.StatusString(),
			},
			"Running":    tainr.Running,
			"Status":     tainr.StateString(),
			"Paused":     false,
			"Restarting": false,
			"OOMKilled":  false,
			"Dead":       tainr.Failed,
			"StartedAt":  tainr.Created.Format("2006-01-02T15:04:05Z"),
			"FinishedAt": tainr.Finished.Format("2006-01-02T15:04:05Z"),
			"ExitCode":   0,
			"Error":      errstr,
		}
		res["Config"] = gin.H{
			"Image":  tainr.Image,
			"Labels": tainr.Labels,
			"Env":    tainr.Env,
			"Cmd":    tainr.Cmd,
			"Tty":    false,
		}
	} else {
		res["Created"] = tainr.Created.Format("2006-01-02T15:04:05Z")
		res["Labels"] = tainr.Labels
		res["State"] = tainr.StatusString()
		res["Status"] = tainr.StateString()
	}
	return res
}

// getContainerNames will list of possible names to identify the container.
func getContainerNames(tainr *types.Container) []string {
	names := []string{}
	if tainr.Name != "" {
		names = append(names, "/"+tainr.Name)
	}
	names = append(names, "/"+tainr.ID)
	names = append(names, "/"+tainr.ShortID)
	for _, alias := range tainr.NetworkAliases {
		if alias != tainr.Name {
			names = append(names, "/"+alias)
		}
	}
	return names
}
