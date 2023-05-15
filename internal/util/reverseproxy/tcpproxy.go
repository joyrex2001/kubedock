package reverseproxy

import (
	"fmt"
	"io"
	"net"
	"time"

	"k8s.io/klog"
)

const RATE = 5                        // number of tries per second for retry scenarios
const INITIAL_CONNECT_TRY_TIMEOUT = 5 // number of seconds to try to wait before actually listening

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

	// this is a workaround to make sure that healthchecks based on log output, rather than
	// end-to-end connectivity, have a bit more slack setting up this connectivity; this
	// fixes liquibase read-timeouts whe using quarkus + postgres + liquibase.
	waitUntilRemoteAcceptsConnection(remote, INITIAL_CONNECT_TRY_TIMEOUT)

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
		for try := 0; try < req.MaxRetry*RATE && !done; try++ {
			if remoteAcceptsConnection(remote) {
				klog.V(3).Infof("proxying from 127.0.0.1:%s -> %s", local, remote)
				break
			} else {
				time.Sleep(time.Second / RATE)
			}
		}
		for !done {
			conn, err := listener.Accept()
			klog.V(3).Infof("accepted connection for %s to %s", local, remote)
			if err != nil {
				if !done {
					klog.Errorf("error accepting connection: %s", err)
				}
				continue
			}
			go handleConnection(conn, local, remote, req.MaxRetry)
		}
		return
	}()

	return nil
}

// handleConnection will proxy a single connection towards the given endpoint. If the initial
// connection fails, it will retry with a maximum of 30 tries (equal to 30 seconds). It will
// close the given connection when returned.
func handleConnection(conn net.Conn, local, remote string, maxRetry int) {
	var err error
	var conn2 net.Conn
	for try := 0; try < maxRetry*RATE; try++ {
		conn2, err = net.DialTimeout("tcp", remote, time.Second/RATE)
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

// waitUntilAcceptConnection will wait until the given remote is accepting connections,
// if given timeout seconds is passed, it will return a timeout error.
func waitUntilRemoteAcceptsConnection(remote string, timeout int) error {
	for try := 0; try < timeout*RATE; try++ {
		if !remoteAcceptsConnection(remote) {
			time.Sleep(time.Second / RATE)
			continue
		} else {
			return nil
		}
	}
	return fmt.Errorf("timeout connecting to %s", remote)
}

// remoteAcceptsConnection will check if the given remote is accepting connections.
func remoteAcceptsConnection(remote string) bool {
	conn, err := net.DialTimeout("tcp", remote, time.Second/RATE)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
