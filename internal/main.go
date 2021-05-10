package internal

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/config"
	"github.com/joyrex2001/kubedock/internal/reaper"
	"github.com/joyrex2001/kubedock/internal/server"
)

// Main is the main entry point for starting this service, based the settings
// initiated by cmd.
func Main(cmd *cobra.Command, args []string) {
	kub, err := getBackend()
	if err != nil {
		klog.Fatalf("error instantiating kubernetes: %s", err)
	}

	rpr, err := reaper.New(reaper.Config{
		KeepMax: viper.GetDuration("reaper.keepmax"),
		Backend: kub,
	})
	if err != nil {
		klog.Fatalf("error instantiating reaper: %s", err)
	}
	rpr.Start()

	svr := server.New(kub)
	if err := svr.Run(); err != nil {
		klog.Fatalf("error instantiating server: %s", err)
	}
}

// getBackend will instantiate a the kubedock kubernetes object.
func getBackend() (backend.Backend, error) {
	cfg, err := config.GetKubernetes()
	if err != nil {
		return nil, err
	}
	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	kub := backend.New(backend.Config{
		Client:     cli,
		RestConfig: cfg,
		Namespace:  viper.GetString("kubernetes.namespace"),
		InitImage:  viper.GetString("kubernetes.initimage"),
		KeepPorts:  viper.GetBool("kubernetes.keepports"),
	})
	return kub, nil
}
