package kubernetes

import (
	"fmt"
	"io"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/joyrex2001/kubedock/internal/container"
)

// CopyToContainer will copy given (tar) archive to given path of the container.
func (in *instance) CopyToContainer(tainr *container.Container, archive []byte, path string) error {
	pods, err := in.GetPods(tainr)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("no matching pod found")
	}

	if path != "/" && strings.HasSuffix(string(path[len(path)-1]), "/") {
		path = path[:len(path)-1]
	}

	reader, writer := io.Pipe()
	go func() {
		writer.Write(archive)
		writer.Close()
	}()

	req := in.cli.CoreV1().RESTClient().Post().Resource("pods").
		Name(pods[0].Name).
		Namespace(pods[0].Namespace).
		SubResource("exec")
	req.VersionedParams(&corev1.PodExecOptions{
		Command: []string{"tar", "-xf", "-", "-C", path},
		Stdin:   true,
		Stdout:  false,
		Stderr:  true,
		TTY:     false,
	}, scheme.ParameterCodec)
	ex, err := remotecommand.NewSPDYExecutor(in.cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	return ex.Stream(remotecommand.StreamOptions{
		Stdin:  reader,
		Stderr: os.Stderr,
	})
}
