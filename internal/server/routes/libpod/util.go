package libpod

import (
	"strings"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

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
