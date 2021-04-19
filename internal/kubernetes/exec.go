package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/joyrex2001/kubedock/internal/container"
)

// ExecContainer will execute given exec object in kubernetes.
func (in *instance) ExecContainer(tainr container.Container, exec container.Exec) error {
	pods, err := in.cli.CoreV1().Pods(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubedock=" + tainr.GetID(),
	})
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("no matching pod found")
	}

	// https://stackoverflow.com/questions/43314689/example-of-exec-in-k8ss-pod-by-using-go-client/54317689
	req := in.cli.CoreV1().RESTClient().Post().Resource("pods").Name(pods.Items[0].Name).
		Namespace(pods.Items[0].Namespace).SubResource("exec")
	req.VersionedParams(&corev1.PodExecOptions{
		Command: exec.GetCmd(),
		Stdin:   false,
		Stdout:  false,
		Stderr:  false,
		TTY:     false,
	}, scheme.ParameterCodec)
	ex, err := remotecommand.NewSPDYExecutor(in.cfg, "POST", req.URL())
	if err != nil {
		return err
	}
	err = ex.Stream(remotecommand.StreamOptions{
		// Stdout: stdout,
		// Stderr: stderr,
	})
	if err != nil {
		return err
	}

	return nil
}
