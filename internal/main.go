package internal

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/config"
	"github.com/joyrex2001/kubedock/internal/kubernetes"
	"github.com/joyrex2001/kubedock/internal/reaper"
	"github.com/joyrex2001/kubedock/internal/server"
)

// Main is the main entry point for starting this service, based the settings
// initiated by cmd.
func Main(cmd *cobra.Command, args []string) {
	kub, err := getKubernetes()
	if err != nil {
		klog.Fatalf("error instantiating kubernetes: %s", err)
	}

	rpr, err := reaper.New(reaper.Config{
		KeepMax:    viper.GetDuration("reaper.keepmax"),
		Kubernetes: kub,
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

// getKubernetes will instantiate a the kubedock kubernetes object.
func getKubernetes() (kubernetes.Kubernetes, error) {
	cfg, err := config.GetKubernetes()
	if err != nil {
		return nil, err
	}
	cli, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	kub := kubernetes.New(kubernetes.Config{
		Client:     cli,
		RestConfig: cfg,
		Namespace:  viper.GetString("kubernetes.namespace"),
		InitImage:  viper.GetString("kubernetes.initimage"),
	})
	return kub, nil
}
