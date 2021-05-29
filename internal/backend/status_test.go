package backend

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestContainerStatus(t *testing.T) {
	tests := []struct {
		in     *types.Container
		kub    *instance
		out    bool
		suc    bool
		state  string
		status string
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
						ReadyReplicas: 0,
					},
				}),
			},
			in:     &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			out:    false,
			suc:    true,
			state:  "Created",
			status: "unhealthy",
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
						ReadyReplicas: 0,
					},
				}),
			},
			in:     &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit", Killed: true},
			out:    false,
			suc:    true,
			state:  "Dead",
			status: "unhealthy",
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
						ReadyReplicas: 0,
					},
				}),
			},
			in:     &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit", Stopped: true},
			out:    false,
			suc:    true,
			state:  "Dead",
			status: "unhealthy",
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
			in:     &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
			out:    true,
			suc:    true,
			state:  "Up",
			status: "healthy",
		},
	}

	for i, tst := range tests {
		res, err := tst.kub.IsContainerRunning(tst.in)
		if !tst.suc && err == nil {
			t.Errorf("failed test %d - expected error", i)
		}
		if tst.suc && err != nil {
			t.Errorf("failed test %d - unexpected error %s", i, err)
		}
		if tst.suc && tst.out != res {
			t.Errorf("failed test %d - expected %t, but got %t", i, tst.out, res)
		}
		stat, err := tst.kub.GetContainerStatus(tst.in)
		if !tst.suc && err == nil {
			t.Errorf("failed test status %d - expected error", i)
		}
		if tst.suc && err != nil {
			t.Errorf("failed test status %d - unexpected error %s", i, err)
		}
		if stat.StateString() != tst.state {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.state, stat.StateString())
		}
		if stat.StatusString() != tst.status {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.status, stat.StatusString())
		}
	}
}
