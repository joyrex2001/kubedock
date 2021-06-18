package exec

import (
	"io"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog"
)

// Request is the structure used as argument for RemoteCmd
type Request struct {
	// Clent is the kubernetes clientset
	Client kubernetes.Interface
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// Pod is the selected pod for this port forwarding
	Pod v1.Pod
	// Cmd contains the command to be executed
	Cmd []string
	// Container contains the name of the container in which the cmd should be executed
	Container string
	// Stdin contains a Reader if stdin is required (nil if ignored)
	Stdin io.Reader
	// Stdout contains a Writer if stdout is required (nil if ignored)
	Stdout io.Writer
	// Stderr contains a Writer if stderr is required (nil if ignored)
	Stderr io.Writer
}

// RemoteCmd will execute given exec object in kubernetes.
func RemoteCmd(req Request) error {
	r := req.Client.CoreV1().RESTClient().Post().Resource("pods").
		Name(req.Pod.Name).
		Namespace(req.Pod.Namespace).
		SubResource("exec")

	r.VersionedParams(&corev1.PodExecOptions{
		Container: req.Container,
		Command:   req.Cmd,
		Stdin:     req.Stdin != nil,
		Stdout:    req.Stdout != nil,
		Stderr:    req.Stderr != nil,
		TTY:       false,
	}, scheme.ParameterCodec)

	ex, err := remotecommand.NewSPDYExecutor(req.RestConfig, "POST", r.URL())
	if err != nil {
		return err
	}

	klog.V(3).Infof("exec %s:%v", req.Pod.Name, req.Cmd)

	return ex.Stream(remotecommand.StreamOptions{
		Stdin:  req.Stdin,
		Stdout: req.Stdout,
		Stderr: req.Stderr,
	})
}
