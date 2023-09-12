package dind

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"
)

// Dind is the docker-in-docker proxy server.
type Dind struct {
	kuburl string
	sock   string
}

// New will instantiate a Dind object.
func New(sock string, kuburl string) *Dind {
	return &Dind{
		kuburl: kuburl,
		sock:   sock,
	}
}

// proxy forwards the request to the configured kubedock endpoint.
func (d *Dind) proxy(c *gin.Context) {
	remote, err := url.Parse(d.kuburl)
	if err != nil {
		klog.Errorf("error parsing kubedock url `%s`: %s", d.kuburl, err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("proxyPath")
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// Run will initialize the http api server and start the proxy.
func (d *Dind) Run() error {
	r := gin.Default()

	r.Any("/*proxyPath", d.proxy)

	d.exitHandler()

	if err := r.RunUnix(d.sock); err != nil {
		return err
	}

	return nil
}

// exitHandler will remove the created socket.
func (d *Dind) exitHandler() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		if err := os.Remove(d.sock); err != nil {
			klog.Errorf("error removing socket: %s", err)
		}
		os.Exit(0)
	}()
}
