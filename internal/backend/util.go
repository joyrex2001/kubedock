package backend

import (
	"context"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// asKubernetesName will create a nice kubernetes name out of given random string.
func (in *instance) toKubernetesName(nm string) string {
	for _, exp := range []string{`^[^A-Za-z0-9]+`, `[^A-Za-z0-9-]`, `-*$`} {
		re := regexp.MustCompile(exp)
		nm = re.ReplaceAllString(nm, ``)
		if len(nm) > 63 {
			nm = nm[:63]
		}
	}
	if nm == "" {
		nm = "undef"
	}
	return nm
}

// getPodsLabelSelector will return a label selector that can be used to
// uniquely idenitify pods that belong to this deployment.
func (in *instance) getPodsLabelSelector(tainr *types.Container) string {
	return "kubedock=" + tainr.ShortID
}

// getPods will return a list of pods that are spun up for this deployment.
func (in *instance) getPods(tainr *types.Container) ([]corev1.Pod, error) {
	pods, err := in.cli.CoreV1().Pods(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: in.getPodsLabelSelector(tainr),
	})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
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

// RandomPort will return a random port number.
func (in *instance) RandomPort() int {
	min := 32012
	max := 64319
	p := min
	for i := 0; i < 10; i++ {
		p = (rand.Intn(max-min) + min)
		if _, ok := in.randomPorts[p]; !ok {
			return p
		}
		in.randomPorts[p] = p
	}
	return p
}
