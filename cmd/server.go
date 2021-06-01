package cmd

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal"
	"github.com/joyrex2001/kubedock/internal/config"
)

var cfgFile string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the kubedock api server",
	Run: func(cmd *cobra.Command, args []string) {
		flag.Set("v", viper.GetString("verbosity"))
		internal.Main()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	klog.InitFlags(nil)
	// pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// for now; let's just add everything to rootCmd, as there is only
	// one 'real' cmd (this 'server') it is probably more intuitive to
	// the user to see all available args immediately.
	rootCmd.PersistentFlags().String("listen-addr", ":8999", "Webserver listen address")
	rootCmd.PersistentFlags().String("unix-socket", "", "Unix socket to listen to (instead of port)")
	rootCmd.PersistentFlags().Bool("tls-enable", false, "Enable TLS on api server")
	rootCmd.PersistentFlags().String("tls-key-file", "", "TLS keyfile")
	rootCmd.PersistentFlags().String("tls-cert-file", "", "TLS certificate file")
	rootCmd.PersistentFlags().StringP("namespace", "n", getContextNamespace(), "Namespace in which containers should be orchestrated")
	rootCmd.PersistentFlags().String("initimage", config.Image, "Image to use as initcontainer for volume setup")
	rootCmd.PersistentFlags().BoolP("inspector", "i", false, "Enable image inspect to fetch container port config from a registry")
	rootCmd.PersistentFlags().DurationP("timeout", "t", 1*time.Minute, "Container creating timeout")
	rootCmd.PersistentFlags().DurationP("reapmax", "r", 15*time.Minute, "Reap all resources older than this time")
	rootCmd.PersistentFlags().Bool("lock", false, "Lock namespace for this instance")
	rootCmd.PersistentFlags().Duration("lock-timeout", 15*time.Minute, "Max time trying to acquire namespace lock")
	rootCmd.PersistentFlags().StringP("verbosity", "v", "1", "Log verbosity level")
	rootCmd.PersistentFlags().BoolP("prune-start", "P", false, "Prune all existing kubedock resources before starting")

	viper.BindPFlag("server.listen-addr", rootCmd.PersistentFlags().Lookup("listen-addr"))
	viper.BindPFlag("server.socket", rootCmd.PersistentFlags().Lookup("unix-socket"))
	viper.BindPFlag("server.tls-enable", rootCmd.PersistentFlags().Lookup("tls-enable"))
	viper.BindPFlag("server.tls-cert-file", rootCmd.PersistentFlags().Lookup("tls-cert-file"))
	viper.BindPFlag("server.tls-key-file", rootCmd.PersistentFlags().Lookup("tls-key-file"))
	viper.BindPFlag("kubernetes.namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("kubernetes.initimage", rootCmd.PersistentFlags().Lookup("initimage"))
	viper.BindPFlag("kubernetes.timeout", rootCmd.PersistentFlags().Lookup("timeout"))
	viper.BindPFlag("registry.inspector", rootCmd.PersistentFlags().Lookup("inspector"))
	viper.BindPFlag("reaper.reapmax", rootCmd.PersistentFlags().Lookup("reapmax"))
	viper.BindPFlag("lock.enabled", rootCmd.PersistentFlags().Lookup("lock"))
	viper.BindPFlag("lock.timeout", rootCmd.PersistentFlags().Lookup("lock-timeout"))
	viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity"))
	viper.BindPFlag("prune-start", rootCmd.PersistentFlags().Lookup("prune-start"))

	viper.BindEnv("server.listen-addr", "SERVER_LISTEN_ADDR")
	viper.BindEnv("server.tls-enable", "SERVER_TLS_ENABLE")
	viper.BindEnv("server.tls-cert-file", "SERVER_TLS_CERT_FILE")
	viper.BindEnv("server.tls-key-file", "SERVER_TLS_KEY_FILE")
	viper.BindEnv("kubernetes.namespace", "NAMESPACE")
	viper.BindEnv("kubernetes.initimage", "INIT_IMAGE")
	viper.BindEnv("kubernetes.timeout", "TIME_OUT")
	viper.BindEnv("reaper.reapmax", "REAPER_REAPMAX")

	// kubeconfig
	if home := homeDir(); home != "" {
		rootCmd.PersistentFlags().String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		rootCmd.PersistentFlags().String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	viper.BindPFlag("kubernetes.kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
}

// homeDir returns the current home directory of the user.
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// getContextNamespace will return the namespace that is set in the current
// kubeconfig context, and returns 'default' if none is set.
func getContextNamespace() string {
	res := "default"
	rul := clientcmd.NewDefaultClientConfigLoadingRules()
	if rul == nil {
		return res
	}
	cfg, err := rul.Load()
	if err != nil {
		return res
	}
	ctx := cfg.Contexts[cfg.CurrentContext]
	if ctx != nil && ctx.Namespace != "" {
		res = ctx.Namespace
	}
	return res
}
