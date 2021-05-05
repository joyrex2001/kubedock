package reaper

import (
	"sync"
	"time"

	"github.com/joyrex2001/kubedock/internal/kubernetes"
	"github.com/joyrex2001/kubedock/internal/model"
	"k8s.io/klog"
)

// reaper is the object handles reaping of resources.
type reaper struct {
	db      *model.Database
	keepMax time.Duration
	kub     kubernetes.Kubernetes
	quit    chan struct{}
}

var instance *reaper
var once sync.Once

// Config is the configuration to be used for the Reaper proces.
type Config struct {
	// KeepMax is the maximum age of resources, older resources are deleted
	KeepMax time.Duration
	// Kubernetes is the kubedock kubernetes helper object
	Kubernetes kubernetes.Kubernetes
}

// New will create return the singleton Reaper instance.
func New(cfg Config) (*reaper, error) {
	var err error
	var db *model.Database
	once.Do(func() {
		instance = &reaper{}
		db, err = model.New()
		instance.db = db
		instance.kub = cfg.Kubernetes
		instance.keepMax = cfg.KeepMax
	})
	return instance, err
}

// Start will start the reaper background process.
func (in *reaper) Start() {
	in.quit = make(chan struct{})
	in.reaper()
}

// Stop will stop the reaper process.
func (in *reaper) Stop() {
	in.quit <- struct{}{}
}

// reaper will reap all lingering resources at a steady interval.
func (in *reaper) reaper() {
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
func (in *reaper) clean() {
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
