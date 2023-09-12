package podtemplate

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// PodFromFile will read a given file with pod definition and returns a corev1.Pod
// accordingly.
func PodFromFile(file string) (*corev1.Pod, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	stream, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	obj, gvk, err := decode(stream, nil, nil)
	if err != nil {
		return nil, err
	}
	if gvk.Kind == "Pod" {
		return obj.(*corev1.Pod), nil
	}
	return nil, fmt.Errorf("invalid podtemplate: %s", file)
}
