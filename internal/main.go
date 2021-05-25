package internal

import (
	"os"
	"os/signal"
	"syscall"

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
func Main() {
	kub, err := getBackend()
	if err != nil {
		klog.Fatalf("error instantiating backend: %s", err)
	}

	rpr, err := reaper.New(reaper.Config{
		KeepMax: viper.GetDuration("reaper.reapmax"),
		Backend: kub,
	})
	if err != nil {
		klog.Fatalf("error instantiating reaper: %s", err)
	}
	rpr.Start()

	if viper.GetBool("prune-start") {
		if err := kub.DeleteAll(); err != nil {
			klog.Fatalf("error pruning resources: %s", err)
		}
	}

	if viper.GetBool("prune-exit") {
		pruneAtExit(kub)
	}

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
		TimeOut:    viper.GetDuration("kubernetes.timeout"),
	})
	return kub, nil
}

// pruneAtExit will clean up resources when kubedock exits
func pruneAtExit(kub backend.Backend) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		if err := kub.DeleteWithKubedockID(config.DefaultLabels["kubedock.id"]); err != nil {
			klog.Fatalf("error pruning resources: %s", err)
		}
		os.Exit(0)
	}()
}
