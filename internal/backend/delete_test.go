package backend

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestDeleteContainerKubedockID(t *testing.T) {
	tests := []struct {
		in  *types.Container
		kub *instance
		ins int
	}{
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tb303",
						Namespace: "default",
					},
				}),
			},
			in:  &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			ins: 1,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tb303",
						Namespace: "default",
						Labels:    map[string]string{"kubedock.containerid": "tb303", "kubedock.id": "6502"},
					},
				}),
			},
			in:  &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			ins: 1,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tb303",
						Namespace: "default",
						Labels:    map[string]string{"kubedock.containerid": "tb303", "kubedock.id": "z80"},
					},
				}),
			},
			in:  &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			ins: 0,
		},
	}

	for i, tst := range tests {
		if err := tst.kub.DeleteWithKubedockID("z80"); err != nil {
			t.Errorf("failed test %d - unexpected error  %s", i, err)
		}
		pods, _ := tst.kub.cli.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
		cnt := len(pods.Items)
		if cnt != tst.ins {
			t.Errorf("failed delete instances test %d - expected %d remaining deployments but got %d", i, tst.ins, cnt)
		}
	}
}

func TestDeleteContainers(t *testing.T) {
	tests := []struct {
		in  *types.Container
		kub *instance
		cnt int
	}{
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tb303",
						Namespace: "default",
					},
				}),
			},
			in:  &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			cnt: 1,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tb303",
						Namespace: "default",
						Labels:    map[string]string{"kubedock.containerid": "tb303", "kubedock.id": "6502"},
					},
				}),
			},
			in:  &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			cnt: 0,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tb303",
						Namespace: "default",
						Labels:    map[string]string{"kubedock.containerid": "tb303", "kubedock.id": "z80"},
					},
				}),
			},
			in:  &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			cnt: 0,
		},
	}

	for i, tst := range tests {
		if err := tst.kub.DeleteContainer(tst.in); err != nil {
			t.Errorf("failed test %d - unexpected error  %s", i, err)
		}
		pods, _ := tst.kub.cli.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
		cnt := len(pods.Items)
		if cnt != tst.cnt {
			t.Errorf("failed test %d - expected %d remaining deployments but got %d", i, tst.cnt, cnt)
		}
	}
}

func TestDeleteContainerKubedock(t *testing.T) {
	tests := []struct {
		in  *types.Container
		kub *instance
		all int
	}{
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tb303",
						Namespace: "default",
					},
				}),
			},
			in:  &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			all: 1,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tb303",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "true", "kubedock.id": "6502"},
					},
				}),
			},
			in:  &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			all: 0,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tb303",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "true", "kubedock.id": "z80"},
					},
				}),
			},
			in:  &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			all: 0,
		},
	}

	for i, tst := range tests {
		if err := tst.kub.DeleteAll(); err != nil {
			t.Errorf("failed test %d - unexpected error  %s", i, err)
		}
		pods, _ := tst.kub.cli.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
		cnt := len(pods.Items)
		if cnt != tst.all {
			t.Errorf("failed delete all test %d - expected %d remaining deployments but got %d", i, tst.all, cnt)
		}
	}
}

func TestDeleteServices(t *testing.T) {
	tests := []struct {
		id  string
		kub *instance
		cnt int
	}{
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tr909",
						Namespace: "default",
					},
				}),
			},
			id:  "tb303",
			cnt: 1,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tr909",
						Namespace: "default",
						Labels:    map[string]string{"kubedock.containerid": "tb303"},
					},
				}),
			},
			id:  "tb303",
			cnt: 0,
		},
	}

	for i, tst := range tests {
		if err := tst.kub.deleteServices("kubedock.containerid=" + tst.id); err != nil {
			t.Errorf("failed test %d - unexpected error  %s", i, err)
		}
		svcs, _ := tst.kub.cli.CoreV1().Services("default").List(context.TODO(), metav1.ListOptions{})
		cnt := len(svcs.Items)
		if cnt != tst.cnt {
			t.Errorf("failed test %d - expected %d remaining services but got %d", i, tst.cnt, cnt)
		}
	}
}
func TestDeleteContainersOlderThan(t *testing.T) {
	tests := []struct {
		cnt int
		kub *instance
	}{
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
					},
				}),
			},
			cnt: 1,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "true"},
					},
				}),
			},
			cnt: 0,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "f1spirit",
						Namespace:         "default",
						Labels:            map[string]string{"kubedock": "true"},
						DeletionTimestamp: &metav1.Time{},
					},
				}),
			},
			cnt: 1,
		},
	}

	for i, tst := range tests {
		tst.kub.DeleteContainersOlderThan(100 * time.Millisecond)
		pods, _ := tst.kub.cli.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
		cnt := len(pods.Items)
		if cnt != tst.cnt {
			t.Errorf("failed test %d - expected %d remaining deployments but got %d", i, tst.cnt, cnt)
		}
	}
}

func TestDeletePodsOlderThan(t *testing.T) {
	tests := []struct {
		cnt int
		kub *instance
	}{
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
					},
				}),
			},
			cnt: 1,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "true"},
					},
				}),
			},
			cnt: 0,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "f1spirit",
						Namespace:         "default",
						Labels:            map[string]string{"kubedock": "true"},
						DeletionTimestamp: &metav1.Time{},
					},
				}),
			},
			cnt: 1,
		},
	}

	for i, tst := range tests {
		tst.kub.DeletePodsOlderThan(100 * time.Millisecond)
		pods, _ := tst.kub.cli.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
		cnt := len(pods.Items)
		if cnt != tst.cnt {
			t.Errorf("failed test %d - expected %d remaining deployments but got %d", i, tst.cnt, cnt)
		}
	}
}

func TestServiceContainersOlderThan(t *testing.T) {
	tests := []struct {
		cnt int
		kub *instance
	}{
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
					},
				}),
			},
			cnt: 1,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "true"},
					},
				}),
			},
			cnt: 0,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "f1spirit",
						Namespace:         "default",
						Labels:            map[string]string{"kubedock": "true"},
						DeletionTimestamp: &metav1.Time{},
					},
				}),
			},
			cnt: 1,
		},
	}

	for i, tst := range tests {
		tst.kub.DeleteServicesOlderThan(100 * time.Millisecond)
		svcs, _ := tst.kub.cli.CoreV1().Services("default").List(context.TODO(), metav1.ListOptions{})
		cnt := len(svcs.Items)
		if cnt != tst.cnt {
			t.Errorf("failed test %d - expected %d remaining services but got %d", i, tst.cnt, cnt)
		}
	}
}

func TestDeleteConfigMapsOlderThan(t *testing.T) {
	tests := []struct {
		cnt int
		kub *instance
	}{
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
					},
				}),
			},
			cnt: 1,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
						Labels:    map[string]string{"kubedock": "true"},
					},
				}),
			},
			cnt: 0,
		},
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "f1spirit",
						Namespace:         "default",
						Labels:            map[string]string{"kubedock": "true"},
						DeletionTimestamp: &metav1.Time{},
					},
				}),
			},
			cnt: 1,
		},
	}

	for i, tst := range tests {
		tst.kub.DeleteConfigMapsOlderThan(100 * time.Millisecond)
		cms, _ := tst.kub.cli.CoreV1().ConfigMaps("default").List(context.TODO(), metav1.ListOptions{})
		cnt := len(cms.Items)
		if cnt != tst.cnt {
			t.Errorf("failed test %d - expected %d remaining configmaps but got %d", i, tst.cnt, cnt)
		}
	}
}

func TestWatchDeleteContainer(t *testing.T) {
	kub := &instance{
		namespace: "default",
		cli: fake.NewSimpleClientset(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "f1spirit",
				Namespace: "default",
				Labels:    map[string]string{"kubedock.containerid": "303"},
			},
		}),
	}

	tainr := &types.Container{ShortID: "303"}
	timeout := time.Millisecond * 200

	start := time.Now()
	delch, err := kub.WatchDeleteContainer(tainr, timeout)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	err = kub.DeleteContainer(tainr)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	<-delch
	if time.Since(start) >= timeout {
		t.Errorf("unexpected timeout")
	}

	start = time.Now()
	delch, err = kub.WatchDeleteContainer(tainr, timeout)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	err = kub.DeleteContainer(tainr)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	<-delch
	if time.Since(start) < timeout {
		t.Errorf("expected timeout, but no timeout occurred")
	}
}
