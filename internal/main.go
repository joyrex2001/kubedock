package internal

import (
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/model"
	"github.com/joyrex2001/kubedock/internal/server"
)

// Main is the main entry point for starting this service, based the settings
// initiated by cmd.
func Main(cmd *cobra.Command, args []string) {
	db, err := model.New()
	if err != nil {
		klog.Fatalf("error instantiating database: %s", err)
	}
	s := server.New(db)
	s.Run()
}
