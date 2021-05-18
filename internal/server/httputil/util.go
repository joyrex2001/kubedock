package httputil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"k8s.io/klog"
)

// Error will return an error response in json.
func Error(c *gin.Context, status int, err error) {
	klog.Errorf("error during request[%d]: %s", status, err)
	c.JSON(status, gin.H{
		"message": err.Error(),
	})
}

// NotImplemented will return a not implented response.
func NotImplemented(c *gin.Context) {
	c.Writer.WriteHeader(http.StatusNotImplemented)
}

// HijackConnection interrupts the http response writer to get the
// underlying connection and operate with it.
func HijackConnection(w http.ResponseWriter) (io.ReadCloser, io.Writer, error) {
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return nil, nil, err
	}
	// Flush the options to make sure the client sets the raw mode
	_, _ = conn.Write([]byte{})
	return conn, conn, nil
}

// UpgradeConnection will upgrade the Hijacked connection.
func UpgradeConnection(r *http.Request, out io.Writer) {
	if _, ok := r.Header["Upgrade"]; ok {
		fmt.Fprint(out, "HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n")
	} else {
		fmt.Fprint(out, "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.docker.raw-stream\r\n")
	}
	fmt.Fprint(out, "\r\n")
}

// CloseStreams ensures that a list for http streams are properly closed.
func CloseStreams(streams ...interface{}) {
	for _, stream := range streams {
		if tcpc, ok := stream.(interface {
			CloseWrite() error
		}); ok {
			_ = tcpc.CloseWrite()
		} else if closer, ok := stream.(io.Closer); ok {
			_ = closer.Close()
		}
	}
}

// RequestLoggerMiddleware is a gin-gonic middleware that will log the
// raw request.
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := ioutil.ReadAll(tee)
		c.Request.Body = ioutil.NopCloser(&buf)
		klog.V(5).Infof("Request Headers: %#v", c.Request.Header)
		klog.V(4).Infof("Request Body: %s", string(body))
		c.Next()
	}
}

// reponseWriter is the writer interface used by the ResponseLoggerMiddleware
type reponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write is the writer implementation used by the ResponseLoggerMiddleware
func (w reponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// ResponseLoggerMiddleware is a gin-gonic middleware that will the raw response.
func ResponseLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w := &reponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = w
		c.Next()
		klog.V(4).Infof("Response Body: %s", w.body.String())
	}
}

// VersionAliasMiddleware is a gin-gonic middleware that will remove /v1.xx
// from the url path (ignoring versioned apis).
func VersionAliasMiddleware(router *gin.Engine) gin.HandlerFunc {
	re := regexp.MustCompile(`^/v1.[0-9]+`)
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/v1.") {
			c.Request.URL.Path = re.ReplaceAllString(c.Request.URL.Path, ``)
			router.HandleContext(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
