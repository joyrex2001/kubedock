package reaper

import (
	"testing"
	"time"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestCleanExecs(t *testing.T) {
	rp, _ := New(Config{})
	execReapMax = 20 * time.Millisecond
	rp.db.SaveExec(&types.Exec{})
	if err := rp.CleanExecs(); err != nil {
		t.Errorf("unexpected error while cleaning execs: %s", err)
	}
	if excs, err := rp.db.GetExecs(); err != nil {
		t.Errorf("unexpected error while retrieving execs: %s", err)
	} else {
		if len(excs) != 1 {
			t.Errorf("expected 1 exec, but got %d", len(excs))
		}
	}
	time.Sleep(100 * time.Millisecond)
	if err := rp.CleanExecs(); err != nil {
		t.Errorf("unexpected error while cleaning execs: %s", err)
	}
	if excs, err := rp.db.GetExecs(); err != nil {
		t.Errorf("unexpected error while retrieving execs: %s", err)
	} else {
		if len(excs) != 0 {
			t.Errorf("expected 0 exec, but got %d", len(excs))
		}
	}
}
