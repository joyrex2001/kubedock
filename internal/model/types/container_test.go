package types

import (
	"reflect"
	"sort"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
)

func TestGetEnvVar(t *testing.T) {
	tests := []struct {
		in  *Container
		out []corev1.EnvVar
	}{
		{
			in: &Container{Env: []string{
				"rc738",
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
				"sh101":     0,
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
	if len(tainr.StopChannels) != 0 {
		t.Errorf("expected stop channels to be erased")
	}
}

func TestVolumes(t *testing.T) {
	tests := []struct {
		in  *Container
		out map[string]string
		vol bool
	}{
		{
			in: &Container{Binds: []string{
				"/tmp/code:/usr/wbass2/code:ro",
				"/tmp/config:/etc/wbass2:ro",
			}},
			out: map[string]string{
				"/usr/wbass2/code": "/tmp/code",
				"/etc/wbass2":      "/tmp/config",
			},
			vol: true,
		},
		{
			in:  &Container{Binds: []string{}},
			out: map[string]string{},
			vol: false,
		},
	}
	for i, tst := range tests {
		res := tst.in.GetVolumes()
		if !reflect.DeepEqual(res, tst.out) {
			t.Errorf("failed test %d - expected %v, but got %v", i, tst.out, res)
		}
		if tst.in.HasVolumes() != tst.vol {
			t.Errorf("failed test %d - expected %t, but got %t", i, tst.in.HasVolumes(), tst.vol)
		}
	}
}

func TestConnectNetwork(t *testing.T) {
	var err error
	in := &Container{}
	in.ConnectNetwork("1234")
	if in.Networks == nil {
		t.Errorf("networks to be expect populated when container is connected")
	}
	if _, ok := in.Networks["1234"]; !ok {
		t.Errorf("network 1234 expected to be connected")
	}
	err = in.DisconnectNetwork("1234")
	if err != nil {
		t.Errorf("unexpected error on delete %s", err)
	}
	if _, ok := in.Networks["1234"]; ok {
		t.Errorf("network 1234 expected to be disconnected")
	}
	err = in.DisconnectNetwork("1234")
	if err == nil {
		t.Errorf("expected error on delete non existing network, but got none")
	}
	err = in.DisconnectNetwork("bridge")
	if err == nil {
		t.Errorf("expected error on delete bridge, but got none")
	}
}
