package libpod

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/events"
	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/filter"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes/common"
)

// ContainerCreate - create a container.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerCreateLibpod
// POST "/libpod/containers/create"
func ContainerCreate(cr *common.ContextRouter, c *gin.Context) {
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
	in.Labels[types.LabelServiceAccount] = cr.Config.ServiceAccount

	tainr := &types.Container{
		Name:         in.Name,
		Image:        in.Image,
		Entrypoint:   in.Entrypoint,
		User:         in.User,
		Cmd:          in.Command,
		Env:          in.Env,
		Binds:        []string{},
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

	for _, mount := range in.Mounts {
		tainr.Binds = append(tainr.Binds, mount.Source+":"+mount.Destination)
	}

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

// ContainerWait - Block until a container stops, then returns the exit code.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerWaitLibpod
// POST "/libpod/containers/:id/wait"
func ContainerWait(cr *common.ContextRouter, c *gin.Context) {
	id := c.Param("id")
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			tainr, err := cr.DB.GetContainer(id)
			if err == nil {
				common.UpdateContainerStatus(cr, tainr)
			}
			if err != nil || tainr.Stopped || tainr.Killed || tainr.Completed {
				c.Data(http.StatusOK, "application/json", []byte("0"))
				return
			}
		}
	}
}

// ContainerDelete - remove a container.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerDeleteLibpod
// DELETE "/libpod/containers/:id"
func ContainerDelete(cr *common.ContextRouter, c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{
			"cause":    err,
			"message":  "",
			"response": http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, []gin.H{})
}

// ContainerExists - Check if container exists.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerInspectLibpod
// GET "/libpod/containers/:id/exists"
func ContainerExists(cr *common.ContextRouter, c *gin.Context) {
	id := c.Param("id")
	_, err := cr.DB.GetContainer(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// ContainerInfo - return low-level information about a container.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/containers/operation/ContainerInspectLibpod
// GET "/libpod/containers/:id/json"
func ContainerInfo(cr *common.ContextRouter, c *gin.Context) {
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
func ContainerList(cr *common.ContextRouter, c *gin.Context) {
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

// getContainerInfo will return a gin.H containing the details of the
// given container.
func getContainerInfo(cr *common.ContextRouter, tainr *types.Container, detail bool) gin.H {
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
		"Name":  tainr.Name,
		"Image": tainr.Image,
		"NetworkSettings": gin.H{
			"Networks": netdtl,
			"Ports":    getNetworkSettingsPorts(cr, tainr),
		},
		"HostConfig": gin.H{
			"PortBindings": getNetworkSettingsPorts(cr, tainr),
		},
		"Names": getContainerNames(tainr),
	}
	common.UpdateContainerStatus(cr, tainr)
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
		names = append(names, tainr.Name)
	}
	names = append(names, tainr.ID)
	names = append(names, tainr.ShortID)
	for _, alias := range tainr.NetworkAliases {
		if alias != tainr.Name {
			names = append(names, alias)
		}
	}
	return names
}

// getNetworkSettingsPorts will return the available ports of the container
// as a gin.H json structure to be used in container details.
func getNetworkSettingsPorts(cr *common.ContextRouter, tainr *types.Container) gin.H {
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

// getAvailablePorts will return all ports that are currently available on
// the running container.
func getAvailablePorts(cr *common.ContextRouter, tainr *types.Container) map[int][]int {
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

// addNetworkAliases will add the networkaliases as defined in the provided
// NetworksProperty to the container.
func addNetworkAliases(tainr *types.Container, networks map[string]NetworksProperty) {
	aliases := []string{}
	done := map[string]string{tainr.ShortID: tainr.ShortID}
	for _, netwp := range networks {
		for _, a := range netwp.Aliases {
			if _, ok := done[a]; !ok {
				alias := strings.ToLower(a)
				aliases = append(aliases, alias)
				done[alias] = alias
			}
		}
	}
	tainr.NetworkAliases = aliases
}
