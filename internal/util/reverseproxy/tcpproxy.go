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
	// MaxRetry is the maximum number of retries (equals to seconds) upon error
	// and initial connection.
	MaxRetry int
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
		try := 0
		accept := false
		for {
			if done {
				return
			}
			if !accept && try < req.MaxRetry {
				// wait until the target endpoint is also accepting connections before
				// accepting proxied connections.
				var conn net.Conn
				if conn, err = net.DialTimeout("tcp", remote, time.Second); err != nil {
					try++
					time.Sleep(time.Second)
					continue
				}
				klog.V(3).Infof("proxying from 127.0.0.1:%s -> %s", local, remote)
				accept = true
				conn.Close()
			}
			conn, err := listener.Accept()
			if err != nil {
				if !done {
					klog.Errorf("error accepting connection: %s", err)
				}
				continue
			}
			go handleConnection(conn, local, remote, req.MaxRetry)
		}
	}()

	return nil
}

// handleConnection will proxy a single connection towards the given endpoint. If the initial
// connection fails, it will retry with a maximum of 30 tries (equal to 30 seconds). It will
// close the given connection when returned.
func handleConnection(conn net.Conn, local, remote string, maxRetry int) {
	var err error
	var conn2 net.Conn
	for try := 0; try < maxRetry; try++ {
		conn2, err = net.DialTimeout("tcp", remote, time.Second)
		if err == nil {
			klog.V(3).Infof("handling connection for %s", local)
			go io.Copy(conn2, conn)
			io.Copy(conn, conn2)
			conn2.Close()
			conn.Close()
			return
		}
		klog.Warningf("error dialing %s: %s (attempt: %d)", remote, err, try)
	}
	klog.Errorf("error dialing %s: max retry attempts reached", remote)
	conn.Close()
}
