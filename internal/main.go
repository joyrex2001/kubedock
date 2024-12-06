package internal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/config"
	"github.com/joyrex2001/kubedock/internal/reaper"
	"github.com/joyrex2001/kubedock/internal/server"
	"github.com/joyrex2001/kubedock/internal/util/myip"
)

// Main is the main entry point for starting this service.
func Main() {
	klog.Infof("%s / kubedock.id=%s", config.VersionString(), config.InstanceID)

	cfg, err := config.GetKubernetes()
	if err != nil {
		klog.Fatalf("error instantiating kubernetes client: %s", err)
	}

	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("error instantiating kubernetes client: %s", err)
	}

	kub, err := getBackend(cfg, cli)
	if err != nil {
		klog.Fatalf("error instantiating backend: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	exitHandler(kub, cancel)

	// check if this instance requires locking of the namespace, if not
	// just start the show...
	if !viper.GetBool("lock.enabled") {
		run(ctx, kub)
		select {}
	}

	// exclusive mode, use the k8s leader election as a locking mechanism
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "kubedock-lock",
			Namespace: viper.GetString("kubernetes.namespace"),
		},
		Client: cli.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: config.InstanceID,
		},
	}

	ready := lockTimeoutHandler()
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				ready <- struct{}{}
				run(ctx, kub)
			},
			OnStoppedLeading: func() {
				klog.V(3).Infof("lost lock on namespace %s", viper.GetString("kubernetes.namespace"))
			},
			OnNewLeader: func(identity string) {
				klog.V(3).Infof("new leader elected: %s", identity)
			},
		},
	})
	select {}
}

// getBackend will instantiate the kubedock kubernetes object.
func getBackend(cfg *rest.Config, cli kubernetes.Interface) (backend.Backend, error) {
	ns := viper.GetString("kubernetes.namespace")
	initimg := viper.GetString("kubernetes.initimage")
	dindimg := viper.GetString("kubernetes.dindimage")
	disdind := viper.GetBool("kubernetes.disable-dind")
	timeout := viper.GetDuration("kubernetes.timeout")
	podtmpl := viper.GetString("kubernetes.pod-template")
	imgpsr := strings.ReplaceAll(viper.GetString("kubernetes.image-pull-secrets"), " ", "")
	dissvcs := viper.GetBool("disable-services")

	optlog := ""
	imgps := []string{}
	if imgpsr != "" {
		optlog = fmt.Sprintf(", pull secrets=%s", imgpsr)
		imgps = strings.Split(imgpsr, ",")
	}

	klog.Infof("kubernetes config: namespace=%s, initimage=%s, dindimage=%s, ready timeout=%s%s", ns, initimg, dindimg, timeout, optlog)
	if disdind {
		klog.Infof("docker-in-docker support disabled")
	}

	kuburl, err := getKubedockURL()
	if err != nil {
		return nil, err
	}
	klog.V(3).Infof("kubedock url: %s", kuburl)

	return backend.New(backend.Config{
		Client:           cli,
		RestConfig:       cfg,
		Namespace:        ns,
		InitImage:        initimg,
		DindImage:        dindimg,
		DisableDind:      disdind,
		ImagePullSecrets: imgps,
		PodTemplate:      podtmpl,
		KubedockURL:      kuburl,
		TimeOut:          timeout,
		DisableServices:  dissvcs,
	})
}

// getKubedockURL returns the uri that can be used externally to reach
// this kubedock instance.
func getKubedockURL() (string, error) {
	ip, err := myip.Get()
	if err != nil {
		return "", err
	}

	port := strings.Split(viper.GetString("server.listen-addr")+":", ":")[1]
	if port == "" {
		return "", fmt.Errorf("expected a port to be configured for listen-addr")
	}

	proto := "http"
	if viper.GetBool("server.tls-enable") {
		proto = "https"
	}
	return fmt.Sprintf("%s://%s:%s", proto, ip, port), nil
}

// run will start all components, based the settings initiated by cmd.
func run(ctx context.Context, kub backend.Backend) {
	reapmax := viper.GetDuration("reaper.reapmax")
	rpr, err := reaper.New(reaper.Config{
		KeepMax: reapmax,
		Backend: kub,
	})
	if err != nil {
		klog.Fatalf("error instantiating reaper: %s", err)
	}

	klog.Infof("reaper started with max container age %s", reapmax)
	rpr.Start()

	if viper.GetBool("prune-start") {
		klog.Info("pruning all existing kubedock resources from namespace")
		if err := kub.DeleteAll(); err != nil {
			klog.Errorf("error pruning resources: %s", err)
		}
	}

	svr := server.New(kub)
	if err := svr.Run(ctx); err != nil {
		klog.Errorf("error instantiating server: %s", err)
	}
}

// lockTimeoutHandler will wait until the return channel recieved a message,
// if this is not done within configured lock.timeout, it will exit the
// process.
func lockTimeoutHandler() chan struct{} {
	ready := make(chan struct{}, 1)
	go func() {
		for {
			tmr := time.NewTimer(viper.GetDuration("lock.timeout"))
			select {
			case <-ready:
				return
			case <-tmr.C:
				klog.Errorf("timeout acquiring lock")
				// no cleanup required, as nothing was done yet...
				os.Exit(1)
			}
		}
	}()
	return ready
}

// exitHandler will clean up resources before actually stopping kubedock.
func exitHandler(kub backend.Backend, cancel context.CancelFunc) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		c := getExitCode(<-sigc)
		cancel()
		klog.Info("exit signal recieved, removing pods, configmaps and services")
		if err := kub.DeleteWithKubedockID(config.InstanceID); err != nil {
			klog.Errorf("error pruning resources: %s", err)
		}
		os.Exit(c)
	}()
}

// getExitCode will map signal to a meaningfull exit code.
func getExitCode(sig os.Signal) int {
	c := 0
	switch sig := sig.(type) {
	case syscall.Signal:
		c = 128 + int(sig)
	}
	return c
}
