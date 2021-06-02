package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/filter"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// ContainerCreate - create a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerCreate
// POST "/containers/create"
func (cr *Router) ContainerCreate(c *gin.Context) {
	in := &ContainerCreateRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	if in.Name == "" {
		in.Name = c.Query("name")
	}

	tainr := &types.Container{
		Name:         in.Name,
		Image:        in.Image,
		Cmd:          in.Cmd,
		Env:          in.Env,
		ExposedPorts: in.ExposedPorts,
		ImagePorts:   map[string]interface{}{},
		Labels:       in.Labels,
		Binds:        in.HostConfig.Binds,
	}

	if img, err := cr.db.GetImageByNameOrID(in.Image); err != nil {
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
		cr.addNetworkAliases(tainr, endp)
	}

	netw, err := cr.db.GetNetworkByName("bridge")
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	tainr.ConnectNetwork(netw.ID)

	if err := cr.db.SaveContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"Id": tainr.ID,
	})
}

// ContainerStart - start a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerStart
// POST "/containers/:id/start"
func (cr *Router) ContainerStart(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.db.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	running, _ := cr.kub.IsContainerRunning(tainr)
	if !running {
		if err := cr.kub.StartContainer(tainr); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
		tainr.Stopped = false
		tainr.Killed = false
		if err := cr.db.SaveContainer(tainr); err != nil {
			httputil.Error(c, http.StatusInternalServerError, err)
			return
		}
	} else {
		klog.Warningf("container %s already running", id)
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerStop - stop a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerStop
// POST "/containers/:id/stop"
func (cr *Router) ContainerStop(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.db.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	tainr.SignalStop()
	if !tainr.Stopped && !tainr.Killed {
		if err := cr.kub.DeleteContainer(tainr); err != nil {
			klog.Warningf("error while deleting k8s container: %s", err)
		}
	}
	tainr.Stopped = true
	if err := cr.db.SaveContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerKill - kill a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerKill
// POST "/containers/:id/kill"
func (cr *Router) ContainerKill(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.db.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	// signal := strings.ToLower(c.Param("signal"))
	tainr.SignalStop()
	if !tainr.Stopped && !tainr.Killed {
		if err := cr.kub.DeleteContainer(tainr); err != nil {
			klog.Warningf("error while deleting k8s container: %s", err)
		}
	}
	tainr.Killed = true
	if err := cr.db.SaveContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerDelete - remove a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerDelete
// DELETE "/containers/:id"
func (cr *Router) ContainerDelete(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.db.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	tainr.SignalStop()
	if !tainr.Stopped && !tainr.Killed {
		if err := cr.kub.DeleteContainer(tainr); err != nil {
			klog.Warningf("error while deleting k8s container: %s", err)
		}
	}
	if err := cr.db.DeleteContainer(tainr); err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerAttach - attach to a container to read its output or send input.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerAttach
// POST "/containers/:id/attach"
func (cr *Router) ContainerAttach(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.db.GetContainer(id)
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

	running, _ := cr.kub.IsContainerRunning(tainr)
	if !running {
		if err := cr.kub.StartContainer(tainr); err != nil {
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

	if err := cr.kub.GetLogs(tainr, true, 100, out); err != nil {
		klog.Errorf("error retrieving logs: %s", err)
		return
	}
}

// ContainerWait - Block until a container stops, then returns the exit code.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerWait
// POST "/containers/:id/wait"
func (cr *Router) ContainerWait(c *gin.Context) {
	id := c.Param("id")
	for {
		tmr := time.NewTimer(time.Second)
		select {
		case <-tmr.C:
			tainr, err := cr.db.GetContainer(id)
			if err != nil || tainr.Stopped || tainr.Killed {
				c.JSON(http.StatusOK, gin.H{"StatusCode": 0})
				return
			}
		}
	}
}

// ContainerInfo - return low-level information about a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerInspect
// GET "/containers/:id/json"
func (cr *Router) ContainerInfo(c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.db.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, cr.getContainerInfo(tainr, true))
}

// ContainerList - returns a list of containers.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerList
// GET "/containers/json"
func (cr *Router) ContainerList(c *gin.Context) {
	filtr, err := filter.New(c.Query("filters"))
	if err != nil {
		klog.V(5).Infof("unsupported filter: %s", err)
	}
	tainrs, err := cr.db.GetContainers()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	res := []gin.H{}
	for _, tainr := range tainrs {
		if filtr.Match(tainr) {
			res = append(res, cr.getContainerInfo(tainr, false))
		}
	}
	c.JSON(http.StatusOK, res)
}

// getContainerInfo will return a gin.H containing the details of the
// given container.
func (cr *Router) getContainerInfo(tainr *types.Container, detail bool) gin.H {
	errstr := ""
	status, err := cr.kub.GetContainerStatus(tainr)
	if err != nil {
		errstr += err.Error()
	}
	netws, err := cr.db.GetNetworksByIDs(tainr.Networks)
	if err != nil {
		errstr += err.Error()
	}
	netdtl := gin.H{}
	for _, netw := range netws {
		netdtl[netw.Name] = gin.H{"NetworkID": netw.ID, "IPAddress": "127.0.0.1"}
	}
	res := gin.H{
		"Id":    tainr.ID,
		"Name":  "/" + tainr.Name,
		"Image": tainr.Image,
		"Names": cr.getContainerNames(tainr),
		"NetworkSettings": gin.H{
			"Networks": netdtl,
			"Ports":    cr.getNetworkSettingsPorts(tainr),
		},
		"HostConfig": gin.H{
			"NetworkMode": "host",
			"LogConfig": gin.H{
				"Type":   "json-file",
				"Config": gin.H{},
			},
		},
	}
	if detail {
		res["State"] = gin.H{
			"Health": gin.H{
				"Status": status.StatusString(),
			},
			"Running":    status.Replicas > 0,
			"Status":     status.StateString(),
			"Paused":     false,
			"Restarting": false,
			"OOMKilled":  false,
			"Dead":       status.Replicas == 0,
			"StartedAt":  tainr.Created.Format("2006-01-02T15:04:05Z"),
			"FinishedAt": "0001-01-01T00:00:00Z",
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
		res["State"] = status.StatusString()
		res["Status"] = status.StateString()
		res["Created"] = tainr.Created.Unix()
		res["Ports"] = cr.getContainerPorts(tainr)
	}
	return res
}

// getNetworkSettingsPorts will return the available ports of the container
// as a gin.H json structure to be used in container details.
func (cr *Router) getNetworkSettingsPorts(tainr *types.Container) gin.H {
	ports := cr.getAvailablePorts(tainr)
	res := gin.H{}
	for dst, prts := range ports {
		pp := []map[string]string{}
		done := map[int]int{}
		for _, src := range prts {
			if _, ok := done[src]; ok {
				continue
			}
			pp = append(pp, map[string]string{
				"HostIp":   "localhost",
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
func (cr *Router) getContainerPorts(tainr *types.Container) []map[string]interface{} {
	ports := cr.getAvailablePorts(tainr)
	res := []map[string]interface{}{}
	for dst, prts := range ports {
		done := map[int]int{}
		for _, src := range prts {
			if _, ok := done[src]; ok {
				continue
			}
			pp := map[string]interface{}{
				"IP":          "localhost",
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
func (cr *Router) getAvailablePorts(tainr *types.Container) map[int][]int {
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
	add(tainr.HostPorts)
	add(tainr.MappedPorts)
	return ports
}

// getContainerNames will list of possible names to identify the container.
func (cr *Router) getContainerNames(tainr *types.Container) []string {
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
