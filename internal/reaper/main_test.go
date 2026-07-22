package reaper

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/model/types"
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

func TestCleanDisabled(t *testing.T) {
	kub, _ := backend.New(backend.Config{
		Client:    fake.NewSimpleClientset(),
		Namespace: viper.GetString("kubernetes.namespace"),
		InitImage: viper.GetString("kubernetes.initimage"),
	})
	rp, _ := New(Config{
		KeepMax: 0,
		Backend: kub,
	})
	orig := rp.keepMax
	rp.keepMax = 0
	defer func() { rp.keepMax = orig }()

	tainr := &types.Container{}
	rp.db.SaveContainer(tainr)
	defer rp.db.DeleteContainer(tainr)

	time.Sleep(100 * time.Millisecond)
	rp.clean()
	if tainrs, err := rp.db.GetContainers(); err != nil {
		t.Errorf("unexpected error while retrieving containers: %s", err)
	} else if len(tainrs) != 1 {
		t.Errorf("expected old container to survive with reaping disabled, got %d", len(tainrs))
	}

	rp.keepMax = 20 * time.Millisecond
	rp.clean()
	if tainrs, err := rp.db.GetContainers(); err != nil {
		t.Errorf("unexpected error while retrieving containers: %s", err)
	} else if len(tainrs) != 0 {
		t.Errorf("expected old container reaped once enabled, got %d", len(tainrs))
	}
}
