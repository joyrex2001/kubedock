package backend

import (
	"fmt"
	"testing"
)

func TestParseExecResponse(t *testing.T) {
	tests := []struct {
		in  error
		cod int
		suc bool
	}{
		{nil, 0, true},
		{fmt.Errorf("some generic error"), 0, false},
		{fmt.Errorf("command terminated with exit code 2"), 2, true},
	}

	for i, tst := range tests {
		kub := &instance{}
		cod, err := kub.parseExecResponse(tst.in)
		if cod != tst.cod {
			t.Errorf("failed test %d - expected %d, but got %d", i, tst.cod, cod)
		}
		if err != nil && tst.suc {
			t.Errorf("failed test %d - unexpected error: %s", i, err)
		}
		if err == nil && !tst.suc {
			t.Errorf("failed test %d - expected error, but succeeded instead", i)
		}
	}
}
