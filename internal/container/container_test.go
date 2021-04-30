package container

import (
	"reflect"
	"sort"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
)

func TestGetKubernetesName(t *testing.T) {
	tests := []struct {
		in  *Container
		out string
	}{
		{
			in:  &Container{Name: "StrategicMars"},
			out: "StrategicMars",
		},
		{
			in:  &Container{Name: "", ID: "2107007e-b7c8-df23-18fb-6a6f79726578"},
			out: "2107007e-b7c8-df23-18fb-6a6f79726578",
		},
		{
			in:  &Container{Name: "0123456789012345678901234567890123456789012345678901234567890123456789"},
			out: "012345678901234567890123456789012345678901234567890123456789012",
		},
	}
	for i, tst := range tests {
		res := tst.in.GetKubernetesName()
		if res != tst.out {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.out, res)
		}
	}
}

func TestGetEnvVar(t *testing.T) {
	tests := []struct {
		in  *Container
		out []corev1.EnvVar
	}{
		{
			in: &Container{Env: []string{
				"rc743=Penguin Adventure",
				"rc768=Space Manbow",
			}},
			out: []corev1.EnvVar{
				{Name: "rc743", Value: "Penguin Adventure"},
				{Name: "rc768", Value: "Space Manbow"},
			},
		},
	}
	for i, tst := range tests {
		res := tst.in.GetEnvVar()
		if !reflect.DeepEqual(res, tst.out) {
			t.Errorf("failed test %d - expected %v, but got %v", i, tst.out, res)
		}
	}
}

func TestMapPort(t *testing.T) {
	in := &Container{}
	if in.MappedPorts != nil {
		t.Errorf("mapped ports to be expect nil when no mapping done")
	}
	in.MapPort(808, 1808)
	in.MapPort(909, 1909)
	if !reflect.DeepEqual(in.MappedPorts, map[int]int{808: 1808, 909: 1909}) {
		t.Errorf("port mapping failed")
	}
}

func TestGetContainerTCPPorts(t *testing.T) {
	tests := []struct {
		in  *Container
		out []int
	}{
		{
			in: &Container{ExposedPorts: map[string]interface{}{
				"303/tcp":   0,
				"606/udp":   0,
				"tr808/tcp": 0,
				"909/tcp":   0,
			}},
			out: []int{303, 909},
		},
	}
	for i, tst := range tests {
		res := tst.in.GetContainerTCPPorts()
		sort.Ints(res)
		if !reflect.DeepEqual(res, tst.out) {
			t.Errorf("failed test %d - expected %v, but got %v", i, tst.out, res)
		}
	}
}

func TestStop(t *testing.T) {
	tainr := &Container{}
	res := 0
	stop := make(chan struct{}, 1)
	done := make(chan struct{}, 1)
	tainr.AddStopChannel(stop)
	go func(res *int, in chan struct{}) {
		<-in
		*res = 1
		done <- struct{}{}
	}(&res, stop)
	tainr.SignalStop()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
	}
	if res != 1 {
		t.Errorf("failed stop channels")
	}
}
