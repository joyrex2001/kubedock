package reaper

import (
	"time"

	"k8s.io/klog"
)

// CleanContainers will clean all lingering containers that are
// older than the configured keepMax duration, and stored locally
// in the in memory database.
func (in *Reaper) CleanContainers() error {
	tainrs, err := in.db.GetContainers()
	if err != nil {
		return err
	}
	for _, tainr := range tainrs {
		if tainr.Created.Before(time.Now().Add(-in.keepMax)) {
			klog.V(3).Infof("deleting container: %s", tainr.ID)
			if err := in.kub.DeleteContainer(tainr); err != nil {
				// inform only, if deleting somehow failed, the
				// CleanContainersKubernetes will pick it up anyways
				klog.Warningf("error deleting deployment: %s", err)
			}
			if err := in.db.DeleteContainer(tainr); err != nil {
				return err
			}
		}
	}
	return nil
}

// CleanContainersKubernetes will clean all lingering containers
// that are older than the configured keepMax duration, and stored
// not stored in the local in memory database.
func (in *Reaper) CleanContainersKubernetes() error {
	if err := in.kub.DeleteContainersOlderThan(in.keepMax * 2); err != nil {
		return err
	}
	return in.kub.DeleteServicesOlderThan(in.keepMax * 2)
}
