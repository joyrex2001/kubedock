package portforward

import (
	"testing"
)

func TestRandomPort(t *testing.T) {
	m := map[int]int{}
	for i := 0; i < 100; i++ {
		p := RandomPort()
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
