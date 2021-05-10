package backend

import (
	"regexp"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// asKubernetesName will create a nice kubernetes name out of given random string.
func (in *instance) toKubernetesName(nm string) string {
	for _, exp := range []string{`^[^A-Za-z0-9]+`, `[^A-Za-z0-9-]`, `-*$`} {
		re := regexp.MustCompile(exp)
		nm = re.ReplaceAllString(nm, ``)
		if len(nm) > 63 {
			nm = nm[:63]
		}
	}
	if nm == "" {
		nm = "undef"
	}
	return nm
}

// GetKubernetesName will return the a k8s compatible name of the container.
func (in *instance) getContainerName(tainr *types.Container) string {
	n := in.toKubernetesName(tainr.Name)
	if n != "undef" {
		return n
	}
	return in.toKubernetesName(tainr.ID)
}
