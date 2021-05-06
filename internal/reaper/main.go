package reaper

import (
	"sync"
	"time"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/model"
	"k8s.io/klog"
)

// Reaper is the object handles reaping of resources.
type Reaper struct {
	db      *model.Database
	keepMax time.Duration
	kub     backend.Backend
	quit    chan struct{}
}

var instance *Reaper
var once sync.Once

// Config is the configuration to be used for the Reaper proces.
type Config struct {
	// KeepMax is the maximum age of resources, older resources are deleted.
	KeepMax time.Duration
	// Backend is the kubedock backend object.
	Backend backend.Backend
}

// New will create return the singleton Reaper instance.
func New(cfg Config) (*Reaper, error) {
	var err error
	var db *model.Database
	once.Do(func() {
		instance = &Reaper{}
		db, err = model.New()
		instance.db = db
		instance.kub = cfg.Backend
		instance.keepMax = cfg.KeepMax
	})
	return instance, err
}

// Start will start the reaper background process.
func (in *Reaper) Start() {
	in.quit = make(chan struct{})
	in.runloop()
}

// Stop will stop the reaper process.
func (in *Reaper) Stop() {
	in.quit <- struct{}{}
}

// runloop will reap all lingering resources at a steady interval.
func (in *Reaper) runloop() {
	go func() {
		for {
			tmr := time.NewTimer(time.Minute)
			select {
			case <-in.quit:
				return
			case <-tmr.C:
				klog.V(2).Info("start cleaning lingering objects...")
				in.clean()
				klog.V(2).Info("finished cleaning lingering objects...")
			}
		}
	}()
}

// clean will run all cleaners.
func (in *Reaper) clean() {
	if err := in.CleanExecs(); err != nil {
		klog.Errorf("error cleaning execs: %s", err)
	}
	if err := in.CleanContainers(); err != nil {
		klog.Errorf("error cleaning containers: %s", err)
	}
	if err := in.CleanContainersKubernetes(); err != nil {
		klog.Errorf("error cleaning k8s containers: %s", err)
	}
}
