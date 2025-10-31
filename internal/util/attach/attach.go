package attach

import (
	"context"
	"io"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// Request is the structure used as argument for RemoteAttach
type Request struct {
	// Client is the kubernetes clientset
	Client kubernetes.Interface
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// Pod is the selected pod for this port forwarding
	Pod v1.Pod
	// Container contains the name of the container in which the cmd should be executed
	Container string
	// Stdin contains a Reader if stdin is required (nil if ignored)
	Stdin io.Reader
	// Stdout contains a Writer if stdout is required (nil if ignored)
	Stdout io.Writer
	// Stderr contains a Writer if stderr is required (nil if ignored)
	Stderr io.Writer
	// TTY will enable interactive tty mode (requires stdin)
	TTY bool
}

// RemoteAttach attaches to an existing container in a pod.
func RemoteAttach(req Request) error {
	r := req.Client.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(req.Pod.Name).
		Namespace(req.Pod.Namespace).
		SubResource("attach")

	r.VersionedParams(&corev1.PodAttachOptions{
		Container: req.Container,
		Stdin:     req.Stdin != nil,
		Stdout:    req.Stdout != nil,
		Stderr:    req.Stderr != nil,
		TTY:       req.TTY,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(req.RestConfig, "POST", r.URL())
	if err != nil {
		return err
	}

	return exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdin:  req.Stdin,
		Stdout: req.Stdout,
		Stderr: req.Stderr,
		Tty:    req.TTY,
	})
}
