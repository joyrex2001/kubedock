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
		res = append(res, gin.H{"Name": netw.Name, "ID": netw.ID, "Driver": "host", "Scope": "local"})
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
	c.JSON(http.StatusOK, gin.H{"Name": netw.Name, "ID": netw.ID, "Driver": "host", "Scope": "local"})
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

	if netw.Name == "bridge" || netw.Name == "none" || netw.Name == "host" {
		httputil.Error(c, http.StatusForbidden, fmt.Errorf("%s is a pre-defined network and cannot be removed", netw.Name))
		return
	}

	tainrs, err := nr.db.GetContainers()
	if err == nil {
		for _, tainr := range tainrs {
			if _, ok := tainr.Networks[netw.ID]; ok {
				httputil.Error(c, http.StatusForbidden, fmt.Errorf("cannot delete network, containers attachd"))
				return
			}
		}
	} else {
		klog.Errorf("error retrieving containers: %s", err)
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
	if err := nr.db.SaveContainer(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	nr.addNetworkAliases(tainr, in.EndpointConfig)
	if err := nr.kub.CreateServices(tainr); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
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
