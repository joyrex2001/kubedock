package dind

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
)

// Dind is the docker-in-docker proxy server.
type Dind struct {
	kuburl string
	sock   string
}

// New will instantiate a Dind object.
func New(sock, kuburl string) *Dind {
	return &Dind{
		kuburl: kuburl,
		sock:   sock,
	}
}

// shutDownHandler will watch the path where the docker socket resides (in the
// background). It will terminates the daemon (exit) when a file called 'shutdown'
// is created/remove/touched.
func (d *Dind) shutdownHandler() error {
	path := filepath.Dir(d.sock)
	shutdown := "shutdown"

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := watcher.Add(path); err != nil {
		return err
	}

	klog.Infof("watching %s/%s for activity", path, shutdown)

	go func() {
		for event := range watcher.Events {
			if strings.HasSuffix(event.Name, shutdown) {
				klog.Infof("exit signal received...")
				os.Exit(0)
			}
		}
	}()

	return nil
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
	if !klog.V(2) {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := d.shutdownHandler(); err != nil {
		return err
	}

	r := gin.Default()

	r.Any("/*proxyPath", d.proxy)

	klog.Infof("start listening on %s", d.sock)
	if err := r.RunUnix(d.sock); err != nil {
		return err
	}

	return nil
}
