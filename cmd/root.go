package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal"
	"github.com/joyrex2001/kubedock/internal/config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "kubedock",
	Short: "kubedock is a docker api to orchestrate containers on kubernetes.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		flag.Set("v", viper.GetString("verbosity"))
		internal.Main()
	},
}

func init() {
	klog.InitFlags(nil)
	// pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	rootCmd.PersistentFlags().String("listen-addr", ":8080", "Webserver listen address")
	rootCmd.PersistentFlags().String("socket", "", "Unix socket to listen to (instead of port)")
	rootCmd.PersistentFlags().Bool("tls-enable", false, "Enable TLS on api server")
	rootCmd.PersistentFlags().String("tls-key-file", "", "TLS keyfile")
	rootCmd.PersistentFlags().String("tls-cert-file", "", "TLS certificate file")
	rootCmd.PersistentFlags().String("namespace", "default", "Namespace in which containers should be orchestrated")
	rootCmd.PersistentFlags().String("initimage", config.Image, "Image to use as initcontainer for volume setup")
	rootCmd.PersistentFlags().Duration("keepmax", 5*time.Minute, "Reap all resources older than this time")
	rootCmd.PersistentFlags().StringP("verbosity", "v", "1", "Log verbosity level")

	viper.BindPFlag("server.listen-addr", rootCmd.PersistentFlags().Lookup("listen-addr"))
	viper.BindPFlag("server.socket", rootCmd.PersistentFlags().Lookup("socket"))
	viper.BindPFlag("server.tls-enable", rootCmd.PersistentFlags().Lookup("tls-enable"))
	viper.BindPFlag("server.tls-cert-file", rootCmd.PersistentFlags().Lookup("tls-cert-file"))
	viper.BindPFlag("server.tls-key-file", rootCmd.PersistentFlags().Lookup("tls-key-file"))
	viper.BindPFlag("kubernetes.namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("kubernetes.initimage", rootCmd.PersistentFlags().Lookup("initimage"))
	viper.BindPFlag("reaper.keepmax", rootCmd.PersistentFlags().Lookup("keepmax"))
	viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity"))

	viper.BindEnv("server.listen-addr", "SERVER_LISTEN_ADDR")
	viper.BindEnv("server.socket", "SERVER_SOCKET")
	viper.BindEnv("server.tls-enable", "SERVER_TLS_ENABLE")
	viper.BindEnv("server.tls-cert-file", "SERVER_TLS_CERT_FILE")
	viper.BindEnv("server.tls-key-file", "SERVER_TLS_KEY_FILE")
	viper.BindEnv("kubernetes.namespace", "NAMESPACE")
	viper.BindEnv("kubernetes.initimage", "INIT_IMAGE")
	viper.BindEnv("reaper.keepmax", "REAPER_KEEPMAX")

	// kubeconfig
	if home := homeDir(); home != "" {
		rootCmd.PersistentFlags().String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		rootCmd.PersistentFlags().String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	viper.BindPFlag("kubernetes.kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
