package dind

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"
)

// Dind is the docker-in-docker proxy server.
type Dind struct {
	kuburl string
	sock   string
	port   string
}

// New will instantiate a Dind object.
func New(sock, port, kuburl string) *Dind {
	return &Dind{
		kuburl: kuburl,
		sock:   sock,
		port:   port,
	}
}

// proxy forwards the request to the configured kubedock endpoint.
func (d *Dind) proxy(c *gin.Context) {
	remote, err := url.Parse(d.kuburl)
	if err != nil {
		klog.Errorf("error parsing kubedock url `%s`: %s", d.kuburl, err)
		return
	}

	path := c.Param("proxyPath")

	if path == "/shutdown" {
		klog.Infof("exit signal received...")
		os.Exit(0)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = path
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// Run will initialize the http api server and start the proxy.
func (d *Dind) Run() error {
	if !klog.V(2) {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.Any("/*proxyPath", d.proxy)

	if d.port != "" {
		go func() {
			klog.Infof("start listening on %s", d.port)
			if err := r.Run(d.port); err != nil {
				klog.Fatalf("failed starting webserver on port %s", d.port)
			}
		}()
	}

	klog.Infof("start listening on %s", d.sock)
	if err := r.RunUnix(d.sock); err != nil {
		return err
	}

	return nil
}
