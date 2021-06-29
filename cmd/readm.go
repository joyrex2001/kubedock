package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/joyrex2001/kubedock/internal/util/md2text"
)

var README string
var LICENSE string

func init() {
	rootCmd.AddCommand(readmeCmd)
	readmeCmd.AddCommand(licenseCmd)
}

var readmeCmd = &cobra.Command{
	Use:   "readme",
	Short: "Display project readme",
	Run: func(cmd *cobra.Command, args []string) {
		text := md2text.ToText(README)
		fmt.Println(md2text.Wrap(text, 80))
	},
}

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Display license",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(LICENSE)
	},
}
