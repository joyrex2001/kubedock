package routes

import (
	"reflect"
	"testing"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/types/docker"
)

func TestAddNetworkAliases(t *testing.T) {
	tests := []struct {
		tainr  *types.Container
		endp   docker.EndpointConfig
		out    []string
		portfw bool
	}{
		{
			tainr:  &types.Container{},
			endp:   docker.EndpointConfig{Aliases: []string{"tb303"}},
			out:    []string{"tb303"},
			portfw: true,
		},
		{
			tainr:  &types.Container{NetworkAliases: []string{"tb303"}},
			endp:   docker.EndpointConfig{},
			out:    []string{"tb303"},
			portfw: true,
		},
		{
			tainr:  &types.Container{NetworkAliases: []string{"tb303"}},
			endp:   docker.EndpointConfig{Aliases: []string{"tb303"}},
			out:    []string{"tb303"},
			portfw: true,
		},
		{
			tainr:  &types.Container{NetworkAliases: []string{"tb303", "tr909"}},
			endp:   docker.EndpointConfig{Aliases: []string{"tb303"}},
			out:    []string{"tb303", "tr909"},
			portfw: true,
		},
		{
			tainr:  &types.Container{NetworkAliases: []string{"tb303"}},
			endp:   docker.EndpointConfig{Aliases: []string{"tb303", "tr909"}},
			out:    []string{"tb303", "tr909"},
			portfw: true,
		},
	}

	for i, tst := range tests {
		routr := &Router{cfg: Config{PortForward: tst.portfw}}
		routr.addNetworkAliases(tst.tainr, tst.endp)
		if !reflect.DeepEqual(tst.tainr.NetworkAliases, tst.out) {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.out, tst.tainr.NetworkAliases)
		}
	}
}
