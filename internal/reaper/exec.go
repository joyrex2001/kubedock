package reaper

import (
	"time"

	"k8s.io/klog"
)

var execReapMax = 5 * time.Minute

// CleanExecs will clean all lingering execs that are older than 5 minutes.
func (in *Reaper) CleanExecs() error {
	excs, err := in.db.GetExecs()
	if err != nil {
		return err
	}
	for _, exc := range excs {
		if exc.Created.Before(time.Now().Add(-execReapMax)) {
			klog.V(3).Infof("deleting exec: %s", exc.ID)
			if err := in.db.DeleteExec(exc); err != nil {
				return err
			}
		}
	}
	return nil
}
