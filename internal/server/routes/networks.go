package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// NetworksList - list networks.
// https://docs.docker.com/engine/api/v1.41/#operation/NetworkList
// GET "/networks"
func (nr *Router) NetworksList(c *gin.Context) {
	netws, err := nr.db.GetNetworks()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	res := []gin.H{}
	for _, netw := range netws {
		tainrs := nr.getContainersInNetwork(netw)
		res = append(res, gin.H{
			"Name":       netw.Name,
			"ID":         netw.ID,
			"Driver":     "bridge",
			"Scope":      "local",
			"Attachable": true,
			"Containers": tainrs,
		})
	}
	c.JSON(http.StatusOK, res)
}

// NetworksInfo - inspect a network.
// https://docs.docker.com/engine/api/v1.41/#operation/NetworkInspect
// GET "/network/:id"
func (nr *Router) NetworksInfo(c *gin.Context) {
	id := c.Param("id")
	netw, err := nr.db.GetNetworkByNameOrID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	tainrs := nr.getContainersInNetwork(netw)
	c.JSON(http.StatusOK, gin.H{
		"Name":       netw.Name,
		"ID":         netw.ID,
		"Driver":     "bridge",
		"Scope":      "local",
		"Attachable": true,
		"Containers": tainrs,
	})
}

// NetworksCreate - create a network.
// https://docs.docker.com/engine/api/v1.41/#operation/NetworkCreate
// POST "/networks/create"
func (nr *Router) NetworksCreate(c *gin.Context) {
	in := &NetworkCreateRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	netw := &types.Network{
		Name: in.Name,
	}
	if err := nr.db.SaveNetwork(netw); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"Id": netw.ID,
	})
}

// NetworksDelete - remove a network.
// https://docs.docker.com/engine/api/v1.41/#operation/NetworkDelete
// DELETE "/networks/:id"
func (nr *Router) NetworksDelete(c *gin.Context) {
	id := c.Param("id")
	netw, err := nr.db.GetNetworkByNameOrID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	if netw.IsPredefined() {
		httputil.Error(c, http.StatusForbidden, fmt.Errorf("%s is a pre-defined network and cannot be removed", netw.Name))
		return
	}

	if len(nr.getContainersInNetwork(netw)) != 0 {
		httputil.Error(c, http.StatusForbidden, fmt.Errorf("cannot delete network, containers attachd"))
		return
	}

	if err := nr.db.DeleteNetwork(netw); err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// NetworksConnect - connect a container to a network.
// https://docs.docker.com/engine/api/v1.41/#operation/NetworkConnect
// POST "/networks/:id/connect"
func (nr *Router) NetworksConnect(c *gin.Context) {
	in := &NetworkConnectRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	id := c.Param("id")
	netw, err := nr.db.GetNetworkByNameOrID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	tainr, err := nr.db.GetContainer(in.Container)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}

	tainr.ConnectNetwork(netw.ID)
	n := len(tainr.NetworkAliases)
	nr.addNetworkAliases(tainr, in.EndpointConfig)

	if tainr.Running && n != len(tainr.NetworkAliases) {
		klog.Warningf("adding networkaliases to a running container, will not create new services...")
	}
	if err := nr.db.SaveContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"ID": netw.ID,
	})
}

// NetworksDisconnect - connect a container to a network.
// https://docs.docker.com/engine/api/v1.41/#operation/NetworkDisconnect
// POST "/networks/:id/disconnect"
func (nr *Router) NetworksDisconnect(c *gin.Context) {
	in := &NetworkDisconnectRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	id := c.Param("id")
	_, err := nr.db.GetNetworkByNameOrID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	tainr, err := nr.db.GetContainer(in.Container)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	if err := tainr.DisconnectNetwork(id); err != nil {
		httputil.Error(c, http.StatusNotFound, err)
		return
	}
	if err := nr.db.SaveContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

// NetworksPrune - delete unused networks.
// https://docs.docker.com/engine/api/v1.41/#operation/NetworkPrune
// POST "/networks/prune"
func (nr *Router) NetworksPrune(c *gin.Context) {
	netws, err := nr.db.GetNetworks()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}

	names := []string{}
	for _, netw := range netws {
		if netw.IsPredefined() || len(nr.getContainersInNetwork(netw)) != 0 {
			continue
		}
		if err := nr.db.DeleteNetwork(netw); err != nil {
			httputil.Error(c, http.StatusNotFound, err)
			return
		}
		names = append(names, netw.Name)
	}

	c.JSON(http.StatusCreated, gin.H{
		"NetworksDeleted": names,
	})
}

// getContainersInNetwork will return an array of containers in an array
// of gin.H structs, containing the details of the container.
func (nr *Router) getContainersInNetwork(netw *types.Network) map[string]gin.H {
	res := map[string]gin.H{}
	tainrs, err := nr.db.GetContainers()
	if err == nil {
		for _, tainr := range tainrs {
			if _, ok := tainr.Networks[netw.ID]; ok {
				res[tainr.ID] = gin.H{
					"Name": tainr.Name,
				}
			}
		}
	} else {
		klog.Errorf("error retrieving containers: %s", err)
	}
	return res
}
