package reaper

import (
	"time"

	"k8s.io/klog"
)

// CleanExecs will clean all lingering execs that are older than the
// configured keepMax duration.
func (in *Reaper) CleanExecs() error {
	excs, err := in.db.GetExecs()
	if err != nil {
		return err
	}
	for _, exc := range excs {
		if exc.Created.Before(time.Now().Add(-in.keepMax)) {
			klog.V(3).Infof("deleting exec: %s", exc.ID)
			if err := in.db.DeleteExec(exc); err != nil {
				return err
			}
		}
	}
	return nil
}
