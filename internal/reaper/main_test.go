package reaper

import (
	"testing"
)

func TestNew(t *testing.T) {
	in, _ := New(Config{})
	for i := 0; i < 2; i++ {
		_in, _ := New(Config{})
		if _in != in && in != nil {
			t.Errorf("New failed %d - got different instance", i)
		}
	}
}
