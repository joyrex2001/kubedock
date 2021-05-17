package backend

import (
	"regexp"
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
