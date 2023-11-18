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

// ContainerFromPod will return a corev1.Container that is based on the first
// configured container in the given pod, which can be used as a template
// for to be created containers. If no containers are present in the pod,
// it will return an empty corev1.Container object instead.
func ContainerFromPod(pod *corev1.Pod) corev1.Container {
	container := corev1.Container{}
	if len(pod.Spec.Containers) > 0 {
		container = pod.Spec.Containers[0]
	}
	return container
}
