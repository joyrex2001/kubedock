package docker

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
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerCreate
// POST "/containers/create"
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

	// The User defined in HTTP request takes precedence over the CLI flag
	// if not null.
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
	if in.HostConfig.Memory != 0 {
		in.Labels[types.LabelRequestMemory] = fmt.Sprintf("%d", in.HostConfig.Memory)
	}
	if in.HostConfig.NanoCpus != 0 {
		in.Labels[types.LabelRequestCPU] = fmt.Sprintf("%dn", in.HostConfig.NanoCpus)
	}
	in.Labels[types.LabelServiceAccount] = cr.Config.ServiceAccount

	tainr := &types.Container{
		Name:         in.Name,
		Image:        in.Image,
		Entrypoint:   in.Entrypoint,
		User:         in.User,
		Cmd:          in.Cmd,
		Env:          in.Env,
		ExposedPorts: in.ExposedPorts,
		ImagePorts:   map[string]interface{}{},
		Labels:       in.Labels,
		Binds:        in.HostConfig.Binds,
		PreArchives:  []types.PreArchive{},
	}

	if img, err := cr.DB.GetImageByNameOrID(in.Image); err != nil {
		klog.Warningf("unable to fetch image details: %s", err)
	} else {
		for pp := range img.ExposedPorts {
			tainr.ImagePorts[pp] = pp
		}
	}

	for dst, ports := range in.HostConfig.PortBindings {
		for _, src := range ports {
			if err := tainr.AddHostPort(src.HostPort, dst); err != nil {
				httputil.Error(c, http.StatusInternalServerError, err)
				return
			}
		}
	}

	for _, endp := range in.NetworkConfig.EndpointsConfig {
		addNetworkAliases(tainr, endp)
		if endp.NetworkID != "" {
			netw, err := cr.DB.GetNetworkByNameOrID(endp.NetworkID)
			if err != nil {
				httputil.Error(c, http.StatusInternalServerError, err)
				return
			}
			tainr.ConnectNetwork(netw.ID)
		}
	}

	if len(tainr.Networks) == 0 {
		netw, err := cr.DB.GetNetworkByName("bridge")
		if err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
		tainr.ConnectNetwork(netw.ID)
	}

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
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerStart
// POST "/containers/:id/start"
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
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerRestart
// POST "/containers/:id/restart"
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
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerStop
// POST "/containers/:id/stop"
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
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerKill
// POST "/containers/:id/kill"
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
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerDelete
// DELETE "/containers/:id"
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
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerAttach
// POST "/containers/:id/attach"
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
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerWait
// POST "/containers/:id/wait"
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

// ContainerResize - resize the tty for a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerResize
// POST "/containers/:id/rezise"
func ContainerResize(cr *routes.ContextRouter, c *gin.Context) {
	id := c.Param("id")
	_, err := cr.DB.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
	return
}

// ContainerInfo - return low-level information about a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerInspect
// GET "/containers/:id/json"
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
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerList
// GET "/containers/json"
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
// https://docs.docker.com/engine/api/v1.41/#tag/Container/operation/ContainerRename
// GET "/containers/:id/rename"
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
		"NetworkSettings": gin.H{
			"Networks": netdtl,
			"Ports":    getNetworkSettingsPorts(cr, tainr),
		},
		"HostConfig": gin.H{
			"NetworkMode": "bridge",
			"LogConfig": gin.H{
				"Type":   "json-file",
				"Config": gin.H{},
			},
		},
	}
	if detail {
		updateContainerStatus(cr, tainr)
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
		res["Created"] = tainr.Created.Format("2006-01-02T15:04:05Z")
	} else {
		res["Labels"] = tainr.Labels
		res["State"] = tainr.StatusString()
		res["Status"] = tainr.StateString()
		res["Created"] = tainr.Created.Unix()
		res["Ports"] = getContainerPorts(cr, tainr)
	}
	return res
}

// getNetworkSettingsPorts will return the available ports of the container
// as a gin.H json structure to be used in container details.
func getNetworkSettingsPorts(cr *routes.ContextRouter, tainr *types.Container) gin.H {
	ports := getAvailablePorts(cr, tainr)
	res := gin.H{}
	if tainr.HostIP == "" {
		return res
	}
	for dst, prts := range ports {
		pp := []map[string]string{}
		done := map[int]int{}
		for _, src := range prts {
			if _, ok := done[src]; ok {
				continue
			}
			pp = append(pp, map[string]string{
				"HostIp":   tainr.HostIP,
				"HostPort": fmt.Sprintf("%d", src),
			})
			done[src] = 1
		}
		res[fmt.Sprintf("%d/tcp", dst)] = pp
	}
	return res
}

// getContainerPorts will return the available ports of the container as
// a gin.H json structure to be used in container list.
func getContainerPorts(cr *routes.ContextRouter, tainr *types.Container) []map[string]interface{} {
	ports := getAvailablePorts(cr, tainr)
	res := []map[string]interface{}{}
	if tainr.HostIP == "" {
		return res
	}
	for dst, prts := range ports {
		done := map[int]int{}
		for _, src := range prts {
			if _, ok := done[src]; ok {
				continue
			}
			pp := map[string]interface{}{
				"IP":          tainr.HostIP,
				"PrivatePort": dst,
				"Type":        "tcp",
			}
			if src > 0 {
				pp["PublicPort"] = src
			}
			res = append(res, pp)
			done[src] = 1
		}
	}
	return res
}

// getAvailablePorts will return all ports that are currently available on
// the running container.
func getAvailablePorts(cr *routes.ContextRouter, tainr *types.Container) map[int][]int {
	ports := map[int][]int{}
	add := func(prts map[int]int) {
		for src, dst := range prts {
			if src < 0 {
				continue
			}
			if _, ok := ports[dst]; !ok {
				ports[dst] = []int{}
			}
			ports[dst] = append(ports[dst], src)
		}
	}
	if cr.Config.PortForward || cr.Config.ReverseProxy {
		add(tainr.HostPorts)
		add(tainr.MappedPorts)
	} else {
		add(tainr.GetServicePorts())
	}
	return ports
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
