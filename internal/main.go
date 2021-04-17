package internal

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/joyrex2001/kubedock/internal/server"
)

// Main is the main entry point for starting this service, based the settings
// initiated by cmd.
func Main(cmd *cobra.Command, args []string) {
	s := server.New()
	s.Run(viper.GetString("server.listen-addr"))
}
