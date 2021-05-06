package backend

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestIsContainerRunning(t *testing.T) {
	tests := []struct {
		in  *types.Container
		kub *instance
		out bool
		suc bool
	}{
		{
			kub: &instance{
				namespace: "default",
				cli: fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f1spirit",
						Namespace: "default",
					},
					Status: appsv1.DeploymentStatus{
						ReadyReplicas: 0,
					},
				}),
			},
			in:  &types.Container{ID: "rc752", Name: "f1spirit"},
			out: false,
			suc: true,
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
			in:  &types.Container{ID: "rc752", Name: "f1spirit"},
			out: true,
			suc: true,
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
	}
}
