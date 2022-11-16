package reverseproxy

import (
	"fmt"
	"io"
	"net"
	"time"

	"k8s.io/klog"
)

// Request is the structure used as argument for Proxy
type Request struct {
	// LocalPort is the local port that will be selected for the reverse proxy
	LocalPort int
	// PodPort is the target port for the reverse proxy
	RemotePort int
	// RemoteIP is the target ip for the reverse proxy
	RemoteIP string
	// StopCh is the channel used to manage the reverse proxy lifecycle
	StopCh <-chan struct{}
}

// Proxy will open a reverse tcp proxy, listening to the provided
// local port and proxies this to the given remote ip and destination port.
// based on: https://gist.github.com/vmihailenco/1380352
func Proxy(req Request) error {
	local := fmt.Sprintf("0.0.0.0:%d", req.LocalPort)
	remote := fmt.Sprintf("%s:%d", req.RemoteIP, req.RemotePort)

	klog.Infof("start reverse-proxy %s->%s", local, remote)

	listener, err := net.Listen("tcp", local)
	if err != nil {
		return err
	}

	done := false
	go func() {
		<-req.StopCh
		klog.Infof("stopped reverse-proxy %s->%s", local, remote)
		done = true
		listener.Close()
	}()

	go func() {
		for {
			if done {
				return
			}
			conn, err := listener.Accept()
			if err != nil {
				if !done {
					klog.Warningf("error accepting connection: %s", err)
				}
				continue
			}
			go func() {
				conn2, err := net.DialTimeout("tcp", remote, time.Second)
				if err != nil {
					klog.Warningf("error dialing remote addr: %s", err)
					conn.Close()
					return
				}
				go io.Copy(conn2, conn)
				io.Copy(conn, conn2)
				conn2.Close()
				conn.Close()
			}()
		}
	}()

	return nil
}
