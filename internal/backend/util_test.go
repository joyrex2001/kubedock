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
