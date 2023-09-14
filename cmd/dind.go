package cmd

import (
	"flag"

	"github.com/joyrex2001/kubedock/internal/dind"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog"
)

var dindCmd = &cobra.Command{
	Use:   "dind",
	Short: "Start the kubedock docker-in-docker proxy",
	Run:   startDind,
}

func init() {
	rootCmd.AddCommand(dindCmd)

	dindCmd.PersistentFlags().String("listen-addr", ":2475", "Webserver listen address")
	dindCmd.PersistentFlags().String("unix-socket", "/var/run/docker.sock", "Unix socket to listen to")
	dindCmd.PersistentFlags().String("kubedock-url", "", "Kubedock url to proxy requests to")
	dindCmd.PersistentFlags().StringP("verbosity", "v", "1", "Log verbosity level")

	viper.BindPFlag("dind.socket", dindCmd.PersistentFlags().Lookup("unix-socket"))
	viper.BindPFlag("dind.listen-addr", dindCmd.PersistentFlags().Lookup("listen-addr"))
	viper.BindPFlag("dind.kubedock-url", dindCmd.PersistentFlags().Lookup("kubedock-url"))
	viper.BindPFlag("verbosity", dindCmd.PersistentFlags().Lookup("verbosity"))
}

func startDind(cmd *cobra.Command, args []string) {
	flag.Set("v", viper.GetString("verbosity"))
	dprox := dind.New(viper.GetString("dind.socket"), viper.GetString("dind.listen-addr"), viper.GetString("dind.kubedock-url"))
	if err := dprox.Run(); err != nil {
		klog.Errorf("error running dind proxy: %s", err)
	}
}
