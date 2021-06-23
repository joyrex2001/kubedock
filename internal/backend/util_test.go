package backend

import (
	"testing"
)

func TestAsKubernetsName(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{in: "__-abc", out: "abc"},
		{in: "/a/b/c", out: "abc"},
		{
			in:  "StrategicMars",
			out: "StrategicMars",
		},
		{
			in:  "2107007e-b7c8-df23-18fb-6a6f79726578",
			out: "2107007e-b7c8-df23-18fb-6a6f79726578",
		},
		{
			in:  "0123456789012345678901234567890123456789012345678901234567890123456789",
			out: "012345678901234567890123456789012345678901234567890123456789012",
		},
		{
			in:  "StrategicMars-",
			out: "StrategicMars",
		},
		{
			in:  "2107007e-b7c8-df23-18fb-6a6f79726578",
			out: "2107007e-b7c8-df23-18fb-6a6f79726578",
		},
		{
			in:  "",
			out: "undef",
		},
	}

	for i, tst := range tests {
		kub := &instance{}
		out := kub.toKubernetesName(tst.in)
		if out != tst.out {
			t.Errorf("failed test %d - expected %s, but got %s", i, tst.out, out)
		}
	}
}

func TestRandomPort(t *testing.T) {
	m := map[int]int{}
	kub := &instance{}
	for i := 0; i < 100; i++ {
		p := kub.RandomPort()
		if p < 1024 {
			t.Errorf("Invalid random port %d", p)
			break
		}
		if _, ok := m[p]; ok {
			t.Errorf("Random port collision, port %d already provided", p)
			break
		}
		m[p] = p
	}
}
