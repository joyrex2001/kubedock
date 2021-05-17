package backend

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestWaitReadyState(t *testing.T) {
	tests := []struct {
		in  *types.Container
		kub *instance
		out bool
	}{
		{
			kub: &instance{
				namespace: "default",
				cli:       fake.NewSimpleClientset(),
			},
			in:  &types.Container{Name: "f1spirit"},
			out: true,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tb303",
						Namespace: "default",
					},
				}),
			},
			in:  &types.Container{Name: "f1spirit", ShortID: "tb303"},
			out: true,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tb303",
						Namespace: "default",
					},
					Status: appsv1.DeploymentStatus{
						ReadyReplicas: 1,
					},
				}),
			},
			in:  &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			out: false,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "tr909"},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
					},
				}, &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tr909",
						Namespace: "default",
					},
				}),
			},
			in:  &types.Container{ID: "rc752", ShortID: "tr909", Name: "f1spirit"},
			out: true,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "tr808"},
					},
					Status: corev1.PodStatus{
						ContainerStatuses: []corev1.ContainerStatus{
							{RestartCount: 1},
						},
					},
				}, &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tr808",
						Namespace: "default",
					},
				}),
			},
			in:  &types.Container{ID: "rc752", ShortID: "tr808", Name: "f1spirit"},
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

func TestWaitInitContainerRunning(t *testing.T) {
	tests := []struct {
		in   *types.Container
		name string
		kub  *instance
		out  bool
	}{
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
						InitContainerStatuses: []corev1.ContainerStatus{
							{Name: "setup", State: corev1.ContainerState{Running: nil}},
						},
					},
				}),
			},
			name: "setup",
			in:   &types.Container{ID: "rc752", ShortID: "tr808", Name: "f1spirit"},
			out:  true,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tb303",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "tb303"},
					},
					Status: corev1.PodStatus{
						InitContainerStatuses: []corev1.ContainerStatus{
							{Name: "setup", State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}},
						},
					},
				}),
			},
			name: "setup",
			in:   &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			out:  false,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tr606",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "tr606"},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
					},
				}),
			},
			name: "setup",
			in:   &types.Container{ID: "rc752", ShortID: "tr606", Name: "f1spirit"},
			out:  true,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tr606",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "tr606"},
					},
					Status: corev1.PodStatus{
						InitContainerStatuses: []corev1.ContainerStatus{
							{Name: "setup"},
						},
					},
				}),
			},
			name: "main",
			in:   &types.Container{ID: "rc752", ShortID: "tr606", Name: "f1spirit"},
			out:  true,
		},
	}

	for i, tst := range tests {
		res := tst.kub.waitInitContainerRunning(tst.in, tst.name, 1)
		if (res != nil && !tst.out) || (res == nil && tst.out) {
			t.Errorf("failed test %d - unexpected return value %s", i, res)
		}
	}
}

func TestAddVolumes(t *testing.T) {
	tests := []struct {
		in    *types.Container
		count int
	}{
		{in: &types.Container{}, count: 0},
		{in: &types.Container{Binds: []string{"/local:/remote:rw"}}, count: 1},
	}

	for i, tst := range tests {
		dep := &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{}},
					},
				},
			},
		}
		kub := &instance{}
		kub.addVolumes(tst.in, dep)
		count := len(dep.Spec.Template.Spec.Volumes)
		if count != tst.count {
			t.Errorf("failed test %d - expected %d initContainers, but got %d", i, tst.count, count)
		}
	}
}

func TestContainerPorts(t *testing.T) {
	tests := []struct {
		in    *types.Container
		count int
	}{
		{in: &types.Container{}, count: 0},
		{in: &types.Container{ExposedPorts: map[string]interface{}{"909/tcp": 0}}, count: 1},
	}

	for i, tst := range tests {
		kub := &instance{}
		count := len(kub.getContainerPorts(tst.in))
		if count != tst.count {
			t.Errorf("failed test %d - expected %d container ports, but got %d", i, tst.count, count)
		}
	}
}

func TestGetLabels(t *testing.T) {
	tests := []struct {
		in    *types.Container
		count int
	}{
		{in: &types.Container{}, count: 3},
		{in: &types.Container{Labels: map[string]string{"computer": "msx"}}, count: 3},
	}

	for i, tst := range tests {
		kub := &instance{}
		count := len(kub.getLabels(tst.in))
		if count != tst.count {
			t.Errorf("failed test %d - expected %d labels, but got %d", i, tst.count, count)
		}
	}
}

func TestGetAnnotations(t *testing.T) {
	tests := []struct {
		in    *types.Container
		count int
	}{
		{in: &types.Container{}, count: 1},
		{in: &types.Container{Labels: map[string]string{"computer": "msx"}}, count: 2},
	}

	for i, tst := range tests {
		kub := &instance{}
		count := len(kub.getAnnotations(tst.in))
		if count != tst.count {
			t.Errorf("failed test %d - expected %d labels, but got %d", i, tst.count, count)
		}
	}
}
