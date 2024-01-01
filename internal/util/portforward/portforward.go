package portforward

// Source: https://github.com/gianarb/kube-port-forward

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/klog"
)

// Request is the structure used as argument for ToPod
type Request struct {
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// Pod is the selected pod for this port forwarding
	Pod v1.Pod
	// LocalPort is the local port that will be selected to expose the PodPort
	LocalPort int
	// PodPort is the target port for the pod
	PodPort int
	// StopCh is the channel used to manage the port forward lifecycle
	StopCh <-chan struct{}
	// ReadyCh communicates when the tunnel is ready to receive traffic
	ReadyCh chan struct{}
}

// ToPod will portforward to given pod.
func ToPod(req Request) error {
	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return err
	}

	logr := NewLogger()
	klog.Infof("start port-forward %d->%d", req.LocalPort, req.PodPort)

	url, err := getURLScheme(req)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, url)
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", req.LocalPort, req.PodPort)}, req.StopCh, req.ReadyCh, logr, logr)
	if err != nil {
		return err
	}

	return fw.ForwardPorts()
}

// getURLScheme will take given request and create a valid url scheme for use
// by the portforward api.
func getURLScheme(req Request) (*url.URL, error) {
	portfw := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", req.Pod.Namespace, req.Pod.Name)

	base, err := url.Parse(req.RestConfig.Host)
	if err != nil {
		return nil, fmt.Errorf("error parsing base URL: %w", err)
	}
	if base.Scheme == "" {
		base.Scheme = "https"
	}

	return &url.URL{Scheme: base.Scheme, Host: base.Host, Path: path.Join(base.Path, portfw)}, nil
}
