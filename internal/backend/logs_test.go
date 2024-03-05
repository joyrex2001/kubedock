package backend

import (
	"io"
	"testing"

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
			in:  &types.Container{ID: "rc752", ShortID: "tb303", Name: "f1spirit"},
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

	count := uint64(100)
	logOpts := LogOptions{TailLines: &count}
	for i, tst := range tests {
		r, w := io.Pipe()
		stop := make(chan struct{}, 1)
		res := tst.kub.GetLogs(tst.in, &logOpts, stop, w)
		if (res != nil && !tst.out) || (res == nil && tst.out) {
			t.Errorf("failed test %d - unexpected return value %s", i, res)
		}
		r.Close()
		w.Close()
	}
}
