package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mitchellh/go-wordwrap"

	"github.com/spf13/cobra"
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
		text := README
		// good enough md to text conversion
		re1 := regexp.MustCompile("(?m)^((#+) +(.*$))")
		sub := re1.FindAllStringSubmatch(text, -1)
		for i := range sub {
			ch := ""
			n := len(sub[i][3])
			switch len(sub[i][2]) {
			case 1:
				ch = "\n" + strings.Repeat("=", n)
			case 2:
				ch = "\n" + strings.Repeat("-", n)
			}
			text = strings.ReplaceAll(text, sub[i][1], sub[i][3]+ch)
		}

		re2 := regexp.MustCompile("(?m)\n```.*$")
		re3 := regexp.MustCompile(`(?m)\(http[^\)]*\)`)
		text = re2.ReplaceAllString(text, "")
		text = re3.ReplaceAllString(text, "")

		fmt.Println(wordwrap.WrapString(text, 80))
	},
}

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Display license",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(LICENSE)
	},
}
