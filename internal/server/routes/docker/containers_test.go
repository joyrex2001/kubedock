package docker

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/routes"
)

func TestGetNetworkSettingsPorts(t *testing.T) {
	tests := []struct {
		tainr  *types.Container
		endp   EndpointConfig
		out    gin.H
		portfw bool
	}{
		{
			tainr: &types.Container{
				HostIP:      "",
				MappedPorts: map[int]int{303: 101},
			},
			out:    gin.H{},
			portfw: true,
		},
		{
			tainr: &types.Container{
				HostIP:      "127.0.0.1",
				MappedPorts: map[int]int{303: 101},
			},
			out:    gin.H{"101/tcp": []map[string]string{{"HostIp": "127.0.0.1", "HostPort": "303"}}},
			portfw: true,
		},
		{
			tainr: &types.Container{
				HostIP:    "127.0.0.1",
				HostPorts: map[int]int{303: 101},
			},
			out:    gin.H{"101/tcp": []map[string]string{{"HostIp": "127.0.0.1", "HostPort": "303"}}},
			portfw: true,
		},
		{
			tainr: &types.Container{
				HostIP:      "127.0.0.1",
				MappedPorts: map[int]int{303: 101},
				HostPorts:   map[int]int{303: 101},
			},
			out:    gin.H{"101/tcp": []map[string]string{{"HostIp": "127.0.0.1", "HostPort": "303"}}},
			portfw: true,
		},
		{
			tainr: &types.Container{
				HostIP:      "127.0.0.1",
				MappedPorts: map[int]int{-303: 303},
			},
			out:    gin.H{},
			portfw: true,
		},
		{
			tainr: &types.Container{
				HostIP:      "127.0.0.1",
				MappedPorts: map[int]int{303: 101},
				HostPorts:   map[int]int{202: 101},
			},
			out: gin.H{"101/tcp": []map[string]string{
				{"HostIp": "127.0.0.1", "HostPort": "202"},
				{"HostIp": "127.0.0.1", "HostPort": "303"},
			}},
			portfw: true,
		},
		{
			tainr: &types.Container{
				HostIP:      "127.0.0.1",
				MappedPorts: map[int]int{303: 101},
			},
			out: gin.H{"101/tcp": []map[string]string{
				{"HostIp": "127.0.0.1", "HostPort": "303"},
			}},
			portfw: false,
		},
	}
	for i, tst := range tests {
		cr := &routes.ContextRouter{Config: routes.Config{PortForward: tst.portfw}}
		res := getNetworkSettingsPorts(cr, tst.tainr)
		if !reflect.DeepEqual(res, tst.out) {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.out, res)
		}
	}
}

func TestGetContainerPorts(t *testing.T) {
	tests := []struct {
		tainr *types.Container
		endp  EndpointConfig
		out   []map[string]interface{}
	}{
		{
			tainr: &types.Container{
				HostIP:      "",
				MappedPorts: map[int]int{303: 101},
			},
			out: []map[string]interface{}{},
		},
		{
			tainr: &types.Container{
				HostIP:      "127.0.0.1",
				MappedPorts: map[int]int{303: 101},
			},
			out: []map[string]interface{}{
				{"IP": "127.0.0.1", "PrivatePort": 101, "PublicPort": 303, "Type": "tcp"},
			},
		},
		{
			tainr: &types.Container{
				HostIP:    "127.0.0.1",
				HostPorts: map[int]int{303: 101},
			},
			out: []map[string]interface{}{
				{"IP": "127.0.0.1", "PrivatePort": 101, "PublicPort": 303, "Type": "tcp"},
			},
		},
		{
			tainr: &types.Container{
				HostIP:      "127.0.0.1",
				MappedPorts: map[int]int{303: 101},
				HostPorts:   map[int]int{303: 101},
			},
			out: []map[string]interface{}{
				{"IP": "127.0.0.1", "PrivatePort": 101, "PublicPort": 303, "Type": "tcp"},
			},
		},
		{
			tainr: &types.Container{
				HostIP:      "127.0.0.1",
				MappedPorts: map[int]int{-303: 303},
			},
			out: []map[string]interface{}{},
		},
		{
			tainr: &types.Container{
				HostIP:      "127.0.0.1",
				MappedPorts: map[int]int{303: 101},
				HostPorts:   map[int]int{202: 101},
			},
			out: []map[string]interface{}{
				{"IP": "127.0.0.1", "PrivatePort": 101, "PublicPort": 202, "Type": "tcp"},
				{"IP": "127.0.0.1", "PrivatePort": 101, "PublicPort": 303, "Type": "tcp"},
			},
		},
	}
	for i, tst := range tests {
		cr := &routes.ContextRouter{Config: routes.Config{PortForward: true}}
		res := getContainerPorts(cr, tst.tainr)
		if !reflect.DeepEqual(res, tst.out) {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.out, res)
		}
	}
}

func TestGetContainerNames(t *testing.T) {
	tests := []struct {
		tainr *types.Container
		out   []string
	}{
		{
			tainr: &types.Container{
				ID:             "12345678",
				ShortID:        "1234",
				Name:           "mrghost",
				NetworkAliases: []string{"mrghost", "metalgear"},
			},
			out: []string{"/mrghost", "/12345678", "/1234", "/metalgear"},
		},
	}
	for i, tst := range tests {
		res := getContainerNames(tst.tainr)
		if !reflect.DeepEqual(res, tst.out) {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.out, res)
		}
	}
}
