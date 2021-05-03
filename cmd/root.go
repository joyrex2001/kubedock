package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "kubedock",
	Short: "kubedock is a docker on kubernetes service.",
	Long:  ``,
	Run:   internal.Main,
}

func init() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().String("listen-addr", ":8080", "Webserver listen address")
	rootCmd.PersistentFlags().Bool("enable-tls", false, "Enable TLS on admin webserver")
	rootCmd.PersistentFlags().String("key-file", "", "TLS keyfile")
	rootCmd.PersistentFlags().String("cert-file", "", "TLS certificate file")
	rootCmd.PersistentFlags().StringP("socket", "s", "", "Unix socket to listen to (instead of port)")
	rootCmd.PersistentFlags().String("namespace", "default", "Namespace in which containers should be orchestrated")
	viper.BindPFlag("server.listen-addr", rootCmd.PersistentFlags().Lookup("listen-addr"))
	viper.BindPFlag("server.socket", rootCmd.PersistentFlags().Lookup("socket"))
	viper.BindPFlag("server.enable-tls", rootCmd.PersistentFlags().Lookup("enable-tls"))
	viper.BindPFlag("server.cert-file", rootCmd.PersistentFlags().Lookup("cert-file"))
	viper.BindPFlag("server.key-file", rootCmd.PersistentFlags().Lookup("key-file"))
	viper.BindPFlag("kubernetes.namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindEnv("server.listen-addr", "SERVER_LISTEN_ADDR")
	viper.BindEnv("server.socket", "SERVER_SOCKET")
	viper.BindEnv("server.enable-tls", "SERVER_ENABLE_TLS")
	viper.BindEnv("server.cert-file", "SERVER_CERT_FILE")
	viper.BindEnv("server.key-file", "SERVER_KEY_FILE")
	viper.BindEnv("kubernetes.namespace", "NAMESPACE")

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

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		// fmt.Printf("not using config file: %s\n", err)
	} else {
		fmt.Printf("using config: %s\n", viper.ConfigFileUsed())
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
