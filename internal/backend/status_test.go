package backend

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestGetContainerStatusExitCode(t *testing.T) {
	tests := []struct {
		term  *corev1.ContainerStateTerminated
		state DeployState
		code  int
		err   bool
	}{
		{term: &corev1.ContainerStateTerminated{Reason: "Completed"}, state: DeployCompleted, code: 0},
		{term: &corev1.ContainerStateTerminated{Reason: "Error", ExitCode: 3}, state: DeployCompleted, code: 3},
		{term: &corev1.ContainerStateTerminated{Reason: "OOMKilled", ExitCode: 137}, state: DeployCompleted, code: 137},
		{term: &corev1.ContainerStateTerminated{Reason: "StartError", ExitCode: 128}, state: DeployFailed, code: 128, err: true},
	}
	for i, tst := range tests {
		kub := &instance{
			namespace: "default",
			cli: fake.NewSimpleClientset(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "kubedock-f1spirit-tr909", Namespace: "default"},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Name: "main", State: corev1.ContainerState{Terminated: tst.term}},
					},
				},
			}),
		}
		tainr := &types.Container{ID: "rc752", ShortID: "tr909", Name: "f1spirit"}
		state, err := kub.GetContainerStatus(tainr)
		if (err != nil) != tst.err {
			t.Errorf("failed test %d - unexpected error %v", i, err)
		}
		if state != tst.state {
			t.Errorf("failed test %d - expected state %d, but got %d", i, tst.state, state)
		}
		if tainr.ExitCode != tst.code {
			t.Errorf("failed test %d - expected exit code %d, but got %d", i, tst.code, tainr.ExitCode)
		}
	}
}
