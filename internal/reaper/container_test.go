package reaper

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestCleanContainers(t *testing.T) {
	kub := backend.New(backend.Config{
		Client:    fake.NewSimpleClientset(),
		Namespace: viper.GetString("kubernetes.namespace"),
		InitImage: viper.GetString("kubernetes.initimage"),
	})
	rp, _ := New(Config{
		KeepMax: 20 * time.Millisecond,
		Backend: kub,
	})
	rp.db.SaveContainer(&types.Container{})
	if err := rp.CleanContainers(); err != nil {
		t.Errorf("unexpected error while cleaning containers: %s", err)
	}
	if excs, err := rp.db.GetContainers(); err != nil {
		t.Errorf("unexpected error while retrieving containers: %s", err)
	} else {
		if len(excs) != 1 {
			t.Errorf("expected 1 container, but got %d", len(excs))
		}
	}
	time.Sleep(100 * time.Millisecond)
	if err := rp.CleanContainers(); err != nil {
		t.Errorf("unexpected error while cleaning containers: %s", err)
	}
	if excs, err := rp.db.GetContainers(); err != nil {
		t.Errorf("unexpected error while retrieving containers: %s", err)
	} else {
		if len(excs) != 0 {
			t.Errorf("expected 0 container, but got %d", len(excs))
		}
	}
}
