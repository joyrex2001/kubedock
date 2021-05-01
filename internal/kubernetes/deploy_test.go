package kubernetes

import (
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/joyrex2001/kubedock/internal/container"
)

func TestGetContainerPorts(t *testing.T) {
	tests := []struct {
		in  *container.Container
		out []corev1.ContainerPort
	}{
		{
			in: &container.Container{ExposedPorts: map[string]interface{}{
				"909/tcp": 0,
			}},
			out: []corev1.ContainerPort{
				{ContainerPort: 909, Name: "kd-tcp-909", Protocol: corev1.ProtocolTCP},
			},
		},
	}
	kub := &instance{}
	for i, tst := range tests {
		res := kub.getContainerPorts(tst.in)
		if !reflect.DeepEqual(res, tst.out) {
			t.Errorf("failed test %d - expected %v, but got %v", i, tst.out, res)
		}
	}
}

func TestWaitReadyState(t *testing.T) {
	tests := []struct {
		in  *container.Container
		kub *instance
		out bool
	}{
		{
			kub: &instance{
				namespace: "default",
				cli:       fake.NewSimpleClientset(),
			},
			in: &container.Container{
				Name: "f1spirit",
				ExposedPorts: map[string]interface{}{
					"909/tcp": 0,
				}},
			out: true,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
					},
				}),
			},
			in: &container.Container{
				Name: "f1spirit",
				ExposedPorts: map[string]interface{}{
					"909/tcp": 0,
				}},
			out: true,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
					},
					Status: appsv1.DeploymentStatus{
						ReadyReplicas: 1,
					},
				}),
			},
			in: &container.Container{
				ID:   "rc752",
				Name: "f1spirit",
				ExposedPorts: map[string]interface{}{
					"909/tcp": 0,
				}},
			out: false,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "rc752"},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
					},
				}, &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
					},
				}),
			},
			in: &container.Container{
				ID:   "rc752",
				Name: "f1spirit",
				ExposedPorts: map[string]interface{}{
					"909/tcp": 0,
				}},
			out: true,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "rc752"},
					},
					Status: corev1.PodStatus{
						ContainerStatuses: []corev1.ContainerStatus{
							{RestartCount: 1},
						},
					},
				}, &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
					},
				}),
			},
			in: &container.Container{
				ID:   "rc752",
				Name: "f1spirit",
				ExposedPorts: map[string]interface{}{
					"909/tcp": 0,
				}},
			out: true,
		},
	}

	for i, tst := range tests {
		res := tst.kub.waitReadyState(tst.in, 1)
		if (res != nil && !tst.out) || (res == nil && tst.out) {
			t.Errorf("failed test %d - unexpected return value %s", i, res)
		}
	}
}
