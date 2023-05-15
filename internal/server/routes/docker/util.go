package docker

import (
	"strings"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// addNetworkAliases will add the networkaliases as defined in the provided
// EndpointConfig to the container.
func addNetworkAliases(tainr *types.Container, endp EndpointConfig) {
	aliases := []string{}
	done := map[string]string{tainr.ShortID: tainr.ShortID}
	for _, l := range [][]string{tainr.NetworkAliases, endp.Aliases} {
		for _, a := range l {
			if _, ok := done[a]; !ok {
				alias := strings.ToLower(a)
				aliases = append(aliases, alias)
				done[alias] = alias
			}
		}
	}
	tainr.NetworkAliases = aliases
}
