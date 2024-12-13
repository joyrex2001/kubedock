package backend

import (
	"io"
	"net"
	"os"
	"regexp"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// toKubernetesValue will create a nice kubernetes string that can be used as a
// key out of given random string.
func (in *instance) toKubernetesKey(v string) string {
	return in.replaceValueWithPatterns(v, "", `^[^A-Za-z0-9]+`, `[^A-Za-z0-9-\./]`, `[-/]*$`)
}

// toKubernetesValue will create a nice kubernetes string that can be used as a
// value out of given random string.
func (in *instance) toKubernetesValue(v string) string {
	return in.replaceValueWithPatterns(v, "", `^[^A-Za-z0-9]+`, `[^A-Za-z0-9-\.]`, `-*$`)
}

// toKubernetesNamewill create a nice kubernetes string that can be used as a
// value out of given random string.
func (in *instance) toKubernetesName(v string) string {
	return in.replaceValueWithPatterns(v, "undef", `^[^A-Za-z0-9]+`, `[^A-Za-z0-9-]`, `-*$`)
}

func (in *instance) replaceValueWithPatterns(v, def string, pt ...string) string {
	for _, exp := range pt {
		re := regexp.MustCompile(exp)
		v = re.ReplaceAllString(v, ``)
		if len(v) > 63 {
			v = v[:63]
		}
	}
	if v == "" {
		v = def
	}
	return v
}

// readFile will read given file and return the contents as []byte. If
// failed, it will return an error.
func (in *instance) readFile(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// MapContainerTCPPorts will map random available ports to the ports
// in the container.
func (in *instance) MapContainerTCPPorts(tainr *types.Container) error {
OUTER:
	for _, pp := range tainr.GetContainerTCPPorts() {
		// skip explicitly bound ports
		for src, dst := range tainr.HostPorts {
			if src > 0 && dst == pp {
				continue OUTER
			}
		}
		addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
		if err != nil {
			return err
		}
		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return err
		}
		tainr.MapPort(l.Addr().(*net.TCPAddr).Port, pp)
		defer l.Close()
	}
	return nil
}
