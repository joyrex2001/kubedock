package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/joyrex2001/kubedock/internal/config"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display kubedock version details",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("-------------------------------------------------\n")
		fmt.Printf("kubedock\n")
		fmt.Printf("-------------------------------------------------\n")
		fmt.Printf("version: %s\n", config.Version)
		fmt.Printf("date:    %s\n", config.Date)
		fmt.Printf("build:   %s\n", config.Build)
		fmt.Printf("-------------------------------------------------\n")
	},
}
