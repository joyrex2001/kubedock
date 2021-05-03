package internal

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/joyrex2001/kubedock/internal/server"
	"github.com/joyrex2001/kubedock/internal/util/keyval"
)

// Main is the main entry point for starting this service, based the settings
// initiated by cmd.
func Main(cmd *cobra.Command, args []string) {
	kv, err := keyval.New()
	if err != nil {
		log.Fatalf("error initializing internal database: %s", err)
	}

	s := server.New(kv)
	s.Run()
}
