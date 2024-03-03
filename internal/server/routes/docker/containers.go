package docker

import (
	"encoding/json"
	"fmt"
	"net/http"
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
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerCreate
// POST "/containers/create"
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

	if _, ok := in.Labels[types.LabelRunasUser]; !ok && cr.Config.RunasUser != "" {
		in.Labels[types.LabelRunasUser] = cr.Config.RunasUser
	}
	if in.User != "" {
		// The User defined in HTTP request takes precedence over the cli and label.
		in.Labels[types.LabelRunasUser] = in.User
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
	if _, ok := in.Labels[types.LabelActiveDeadlineSeconds]; !ok && cr.Config.ActiveDeadlineSeconds >= 0 {
		in.Labels[types.LabelActiveDeadlineSeconds] = fmt.Sprintf("%d", cr.Config.ActiveDeadlineSeconds)
	}
	if in.HostConfig.Memory != 0 {
		in.Labels[types.LabelRequestMemory] = fmt.Sprintf("%d", in.HostConfig.Memory)
	}
	if in.HostConfig.NanoCpus != 0 {
		in.Labels[types.LabelRequestCPU] = fmt.Sprintf("%dn", in.HostConfig.NanoCpus)
	}
	in.Labels[types.LabelServiceAccount] = cr.Config.ServiceAccount

	mounts := []types.Mount{}
	for _, m := range in.HostConfig.Mounts {
		if m.Type != "bind" {
			klog.Infof("mount '%s:%s' with type '%s' not supported, ignoring", m.Source, m.Target, m.Type)
			continue
		}
		mounts = append(mounts, types.Mount{
			Type:     m.Type,
			Source:   m.Source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		})
	}

	tainr := &types.Container{
		Name:         in.Name,
		Image:        in.Image,
		Entrypoint:   in.Entrypoint,
		Cmd:          in.Cmd,
		Env:          in.Env,
		ExposedPorts: in.ExposedPorts,
		ImagePorts:   map[string]interface{}{},
		Labels:       in.Labels,
		Binds:        in.HostConfig.Binds,
		Mounts:       mounts,
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

// ContainerWait - Block until a container stops, then returns the exit code.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerWait
// POST "/containers/:id/wait"
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
				c.JSON(http.StatusOK, gin.H{"StatusCode": 0})
				return
			}
		}
	}
}

// ContainerDelete - remove a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerDelete
// DELETE "/containers/:id"
func ContainerDelete(cr *common.ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainerByNameOrID(id)
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

// ContainerInfo - return low-level information about a container.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerInspect
// GET "/containers/:id/json"
func ContainerInfo(cr *common.ContextRouter, c *gin.Context) {
	id := c.Param("id")
	tainr, err := cr.DB.GetContainerByNameOrID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, getContainerInfo(cr, tainr, true))
}

// ContainerList - returns a list of containers.
// https://docs.docker.com/engine/api/v1.41/#operation/ContainerList
// GET "/containers/json"
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
	mounts := []gin.H{}
	for _, m := range tainr.Mounts {
		mounts = append(mounts, gin.H{
			"Source":   m.Source,
			"Target":   m.Target,
			"Type":     m.Type,
			"ReadOnly": m.ReadOnly,
		})
	}
	names := getContainerNames(tainr)
	res := gin.H{
		"Id":    tainr.ID,
		"Name":  names[0],
		"Image": tainr.Image,
		"Names": names,
		"NetworkSettings": gin.H{
			"IPAddress": "127.0.0.1",
			"Networks":  netdtl,
			"Ports":     getNetworkSettingsPorts(cr, tainr),
		},
		"HostConfig": gin.H{
			"NetworkMode": "bridge",
			"LogConfig": gin.H{
				"Type":   "json-file",
				"Config": gin.H{},
			},
			"Mounts": mounts,
		},
	}
	if detail {
		common.UpdateContainerStatus(cr, tainr)
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
			"Image":        tainr.Image,
			"Labels":       tainr.Labels,
			"Env":          tainr.Env,
			"Cmd":          tainr.Cmd,
			"Hostname":     "localhost",
			"ExposedPorts": getConfigExposedPorts(cr, tainr),
			"Tty":          false,
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

// getConfigExposedPorts will return the available ports of the container
// as a gin.H json structure to be used in container config details.
func getConfigExposedPorts(cr *common.ContextRouter, tainr *types.Container) gin.H {
	res := gin.H{}
	if tainr.HostIP == "" {
		return res
	}
	for dst := range getAvailablePorts(cr, tainr) {
		res[fmt.Sprintf("%d/tcp", dst)] = gin.H{}
	}
	return res
}

// getContainerPorts will return the available ports of the container as
// a gin.H json structure to be used in container list.
func getContainerPorts(cr *common.ContextRouter, tainr *types.Container) []map[string]interface{} {
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
