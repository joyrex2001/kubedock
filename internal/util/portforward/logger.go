package portforward

import (
	"io"

	"k8s.io/klog"
)

type logger struct {
	io.Writer
	out io.Writer
}

// NewLogger will return a new logger instance.
func NewLogger() io.Writer {
	return &logger{}
}

// Write will write the log using klog.
func (w *logger) Write(p []byte) (int, error) {
	klog.V(3).Infof(string(p))
	return len(p), nil
}
