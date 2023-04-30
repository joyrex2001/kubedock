package docker

import (
	"reflect"
	"testing"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestAddNetworkAliases(t *testing.T) {
	tests := []struct {
		tainr  *types.Container
		endp   EndpointConfig
		out    []string
		portfw bool
	}{
		{
			tainr: &types.Container{},
			endp:  EndpointConfig{Aliases: []string{"tb303"}},
			out:   []string{"tb303"},
		},
		{
			tainr: &types.Container{NetworkAliases: []string{"tb303"}},
			endp:  EndpointConfig{},
			out:   []string{"tb303"},
		},
		{
			tainr: &types.Container{NetworkAliases: []string{"tb303"}},
			endp:  EndpointConfig{Aliases: []string{"tb303"}},
			out:   []string{"tb303"},
		},
		{
			tainr: &types.Container{NetworkAliases: []string{"tb303", "tr909"}},
			endp:  EndpointConfig{Aliases: []string{"tb303"}},
			out:   []string{"tb303", "tr909"},
		},
		{
			tainr: &types.Container{NetworkAliases: []string{"tb303"}},
			endp:  EndpointConfig{Aliases: []string{"tb303", "tr909"}},
			out:   []string{"tb303", "tr909"},
		},
	}

	for i, tst := range tests {
		addNetworkAliases(tst.tainr, tst.endp)
		if !reflect.DeepEqual(tst.tainr.NetworkAliases, tst.out) {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.out, tst.tainr.NetworkAliases)
		}
	}
}
