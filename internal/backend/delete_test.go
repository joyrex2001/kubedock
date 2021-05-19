package backend

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestDeleteContainer(t *testing.T) {
	tests := []struct {
		in  *types.Container
		kub *instance
		out bool
	}{
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
	}

	for i, tst := range tests {
		res := tst.kub.DeleteContainer(tst.in)
		if (res != nil && !tst.out) || (res == nil && tst.out) {
			t.Errorf("failed test %d - unexpected return value %s", i, res)
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
		if err := tst.kub.deleteServices(tst.id); err != nil {
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
				cli: fake.NewSimpleClientset(&appsv1.Deployment{
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
				cli: fake.NewSimpleClientset(&appsv1.Deployment{
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
				cli: fake.NewSimpleClientset(&appsv1.Deployment{
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
		deps, _ := tst.kub.cli.AppsV1().Deployments("default").List(context.TODO(), metav1.ListOptions{})
		cnt := len(deps.Items)
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
