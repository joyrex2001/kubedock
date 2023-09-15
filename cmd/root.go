package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/klog"
)

var rootCmd = &cobra.Command{
	Use:   "kubedock",
	Short: "Kubedock is a docker api implementation that orchestrate containers on kubernetes.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	klog.InitFlags(nil)
	// pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}
