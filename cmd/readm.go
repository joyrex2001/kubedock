package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/joyrex2001/kubedock/internal/util/md2text"
)

var README string
var CONFIG string
var LICENSE string

func init() {
	rootCmd.AddCommand(readmeCmd)
	readmeCmd.AddCommand(licenseCmd)
	readmeCmd.AddCommand(configCmd)
}

var readmeCmd = &cobra.Command{
	Use:   "readme",
	Short: "Display project readme",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(md2text.ToText(README, 80))
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display project configuration reference",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(md2text.ToText(CONFIG, 80))
	},
}

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Display project license",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(LICENSE)
	},
}
