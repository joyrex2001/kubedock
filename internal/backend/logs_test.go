package backend

import (
	"io"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestGetLogs(t *testing.T) {
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
			in:  &types.Container{ID: "rc752", Name: "f1spirit"},
			out: true,
		},
		// {
		// 	kub: &instance{
		// 		namespace: "default",
		// 		cli: fake.NewSimpleClientset(&corev1.Pod{
		// 			ObjectMeta: metav1.ObjectMeta{
		// 				Name:      "f1spirit",
		// 				Namespace: "default",
		// 				Labels:    map[string]string{"kubedock": "rc752"},
		// 			},
		// 		}),
		// 	},
		// 	in:  &types.Container{ID: "rc752", Name: "f1spirit"},
		// 	out: false,
		// },
	}

	for i, tst := range tests {
		r, w := io.Pipe()
		res := tst.kub.GetLogs(tst.in, false, 100, w)
		if (res != nil && !tst.out) || (res == nil && tst.out) {
			t.Errorf("failed test %d - unexpected return value %s", i, res)
		}
		r.Close()
		w.Close()
	}
}

func TestGetFirstPodName(t *testing.T) {
	tests := []struct {
		in  *types.Container
		kub *instance
		out string
		suc bool
	}{
		{
			kub: &instance{
				namespace: "default",
				cli:       fake.NewSimpleClientset(),
			},
			in:  &types.Container{ID: "rc752", Name: "f1spirit"},
			out: "",
			suc: false,
		},
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
			in:  &types.Container{ID: "rc752", Name: "f1spirit"},
			out: "",
			suc: false,
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
				}),
			},
			in:  &types.Container{ID: "rc752", Name: "f1spirit"},
			out: "f1spirit",
			suc: true,
		},
	}

	for i, tst := range tests {
		res, err := tst.kub.getFirstPodName(tst.in)
		if !tst.suc && err == nil {
			t.Errorf("failed test %d - expected error", i)
		}
		if tst.suc && err != nil {
			t.Errorf("failed test %d - unexpected error %s", i, err)
		}
		if tst.suc && tst.out != res {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.out, res)
		}
	}
}
