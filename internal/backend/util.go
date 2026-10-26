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
	return in.replaceValueWithPatterns(v, "", `^[^A-Za-z0-9]+`, `[^A-Za-z0-9-\./]`, `[-_\./]*$`)
}

// toKubernetesValue will create a nice kubernetes string that can be used as a
// value out of given random string.
func (in *instance) toKubernetesValue(v string) string {
	return in.replaceValueWithPatterns(v, "", `^[^A-Za-z0-9]+`, `[^A-Za-z0-9-\.]`, `[-_\.]*$`)
}

// toKubernetesNamewill create a nice kubernetes string that can be used as a
// value out of given random string.
func (in *instance) toKubernetesName(v string) string {
	return in.replaceValueWithPatterns(v, "undef", `^[^A-Za-z0-9]+`, `[^A-Za-z0-9-]`, `[-_\.]*$`)
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
		// skip already mapped ports (for idempotency)
		for _, dst := range tainr.MappedPorts {
			if dst == pp {
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
		port := l.Addr().(*net.TCPAddr).Port
		tainr.MapPort(port, pp)
		if in.holdPorts {
			in.holdListener(port, l)
		} else {
			defer l.Close()
		}
	}
	return nil
}

// holdListener keeps given listener open until the reverse proxy for its
// port takes it over, so the port can't be claimed in the meantime.
func (in *instance) holdListener(port int, l net.Listener) {
	in.listenersMu.Lock()
	defer in.listenersMu.Unlock()
	if in.listeners == nil {
		in.listeners = map[int]net.Listener{}
	}
	in.listeners[port] = l
}

// takeListener returns the held listener for given port and stops holding
// it, or nil if no listener is held for the port.
func (in *instance) takeListener(port int) net.Listener {
	in.listenersMu.Lock()
	defer in.listenersMu.Unlock()
	l := in.listeners[port]
	delete(in.listeners, port)
	return l
}

// releaseListeners closes the listeners that are still held for the mapped
// ports of given container.
func (in *instance) releaseListeners(tainr *types.Container) {
	for src := range tainr.MappedPorts {
		if l := in.takeListener(src); l != nil {
			l.Close()
		}
	}
}
