package backend

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// toKubernetesValue will create a nice kubernetes string that can be used as a
// key out of given random string.
func (in *instance) toKubernetesKey(v string) string {
	return in.replaceValueWithPatterns(v, "", `^[^A-Za-z0-9]+`, `[^A-Za-z0-9-\./]`, `[-/]*$`)
}

// toKubernetesValue will create a nice kubernetes string that can be used as a
// value out of given random string.
func (in *instance) toKubernetesValue(v string) string {
	return in.replaceValueWithPatterns(v, "", `^[^A-Za-z0-9]+`, `[^A-Za-z0-9-\.]`, `-*$`)
}

// toKubernetesNamewill create a nice kubernetes string that can be used as a
// value out of given random string.
func (in *instance) toKubernetesName(v string) string {
	return in.replaceValueWithPatterns(v, "undef", `^[^A-Za-z0-9]+`, `[^A-Za-z0-9-]`, `-*$`)
}

func (in *instance) replaceValueWithPatterns(v, def string, pt ...string) string {
	for _, exp := range pt {
		re := regexp.MustCompile(exp)
		v = re.ReplaceAllString(v, ``)
		if len(v) > 63 {
			v = v[:63]
		}
	}
	if v == "" {
		v = def
	}
	return v
}

// getPodsLabelSelector will return a label selector that can be used to
// uniquely idenitify pods that belong to this deployment.
func (in *instance) getPodsLabelSelector(tainr *types.Container) string {
	return "kubedock.containerid=" + tainr.ShortID
}

// getPods will return a list of pods that are spun up for this deployment.
func (in *instance) getPods(tainr *types.Container) ([]corev1.Pod, error) {
	pods, err := in.cli.CoreV1().Pods(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: in.getPodsLabelSelector(tainr),
	})
	if err != nil {
		return nil, err
	}
	res := []corev1.Pod{}
	for _, p := range pods.Items {
		if p.ObjectMeta.DeletionTimestamp != nil {
			continue
		}
		res = append(res, p)
	}
	return res, nil
}

// readFile will read given file and return the contents as []byte. If
// failed, it will return an error.
func (in *instance) readFile(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

// MapContainerTCPPorts will map random available ports to the ports
// in the container.
func (in *instance) MapContainerTCPPorts(tainr *types.Container) error {
	for _, pp := range tainr.GetContainerTCPPorts() {
		addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
		if err != nil {
			return err
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return err
		}
		tainr.MapPort(l.Addr().(*net.TCPAddr).Port, pp)
		defer l.Close()
	}
	return nil
}
