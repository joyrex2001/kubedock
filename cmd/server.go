package cmd

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal"
	"github.com/joyrex2001/kubedock/internal/config"
)

var cfgFile string
var labels []string
var annotations []string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the kubedock api server",
	Run: func(cmd *cobra.Command, args []string) {
		flag.Set("v", viper.GetString("verbosity"))
		addDefaultAnnotations(annotations)
		addDefaultLabels(labels)
		internal.Main()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().String("listen-addr", ":2475", "Webserver listen address")
	serverCmd.PersistentFlags().String("unix-socket", "", "Unix socket to listen to (instead of port)")
	serverCmd.PersistentFlags().Bool("tls-enable", false, "Enable TLS on api server")
	serverCmd.PersistentFlags().String("tls-key-file", "", "TLS keyfile")
	serverCmd.PersistentFlags().String("tls-cert-file", "", "TLS certificate file")
	serverCmd.PersistentFlags().StringP("namespace", "n", getContextNamespace(), "Namespace in which containers should be orchestrated")
	serverCmd.PersistentFlags().String("initimage", config.Image, "Image to use as initcontainer for volume setup")
	serverCmd.PersistentFlags().String("dindimage", config.Image, "Image to use as sidecar container for docker-in-docker support")
	serverCmd.PersistentFlags().Bool("disable-dind", false, "Disable docker-in-docker support")
	serverCmd.PersistentFlags().String("pull-policy", "ifnotpresent", "Pull policy that should be applied (ifnotpresent,never,always)")
	serverCmd.PersistentFlags().String("service-account", "default", "Service account that should be used for deployed pods")
	serverCmd.PersistentFlags().String("image-pull-secrets", "", "Comma separated list of image pull secrets that should be used")
	serverCmd.PersistentFlags().String("pod-template", "", "Pod file that should be used as the base for creating pods")
	serverCmd.PersistentFlags().String("pod-name-prefix", "kubedock", "The prefix of the name to be used in the created pods")
	serverCmd.PersistentFlags().BoolP("inspector", "i", false, "Enable image inspect to fetch container port config from a registry")
	serverCmd.PersistentFlags().DurationP("timeout", "t", 1*time.Minute, "Container creating/deletion timeout")
	serverCmd.PersistentFlags().DurationP("reapmax", "r", 60*time.Minute, "Reap all resources older than this time")
	serverCmd.PersistentFlags().String("request-cpu", "", "Default k8s cpu resource request (optionally add ,limit)")
	serverCmd.PersistentFlags().String("request-memory", "", "Default k8s memory resource request (optionally add ,limit)")
	serverCmd.PersistentFlags().String("node-selector", "", "A node selector in the form of key1=value1[,key2=value2]")
	serverCmd.PersistentFlags().Int64("active-deadline-seconds", -1, "Default value for pod deadline, in seconds (a negative value means no deadline)")
	serverCmd.PersistentFlags().String("runas-user", "", "Numeric UID to run pods as (defaults to UID in image)")
	serverCmd.PersistentFlags().Bool("lock", false, "Lock namespace for this instance")
	serverCmd.PersistentFlags().Duration("lock-timeout", 15*time.Minute, "Max time trying to acquire namespace lock")
	serverCmd.PersistentFlags().StringP("verbosity", "v", "1", "Log verbosity level")
	serverCmd.PersistentFlags().BoolP("prune-start", "P", false, "Prune all existing kubedock resources before starting")
	serverCmd.PersistentFlags().Bool("port-forward", false, "Open port-forwards for all services")
	serverCmd.PersistentFlags().Bool("reverse-proxy", false, "Reverse proxy all services via 0.0.0.0 on the kubedock host as well")
	serverCmd.PersistentFlags().Bool("pre-archive", false, "Enable support for copying single files to containers without starting them")
	serverCmd.PersistentFlags().Bool("disable-services", false, "Disable service creation (requires a network solution such as kubedock-dns)")

	viper.BindPFlag("server.listen-addr", serverCmd.PersistentFlags().Lookup("listen-addr"))
	viper.BindPFlag("server.socket", serverCmd.PersistentFlags().Lookup("unix-socket"))
	viper.BindPFlag("server.tls-enable", serverCmd.PersistentFlags().Lookup("tls-enable"))
	viper.BindPFlag("server.tls-cert-file", serverCmd.PersistentFlags().Lookup("tls-cert-file"))
	viper.BindPFlag("server.tls-key-file", serverCmd.PersistentFlags().Lookup("tls-key-file"))
	viper.BindPFlag("kubernetes.namespace", serverCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("kubernetes.initimage", serverCmd.PersistentFlags().Lookup("initimage"))
	viper.BindPFlag("kubernetes.dindimage", serverCmd.PersistentFlags().Lookup("dindimage"))
	viper.BindPFlag("kubernetes.disable-dind", serverCmd.PersistentFlags().Lookup("disable-dind"))
	viper.BindPFlag("kubernetes.pull-policy", serverCmd.PersistentFlags().Lookup("pull-policy"))
	viper.BindPFlag("kubernetes.service-account", serverCmd.PersistentFlags().Lookup("service-account"))
	viper.BindPFlag("kubernetes.image-pull-secrets", serverCmd.PersistentFlags().Lookup("image-pull-secrets"))
	viper.BindPFlag("kubernetes.pod-template", serverCmd.PersistentFlags().Lookup("pod-template"))
	viper.BindPFlag("kubernetes.pod-name-prefix", serverCmd.PersistentFlags().Lookup("pod-name-prefix"))
	viper.BindPFlag("kubernetes.timeout", serverCmd.PersistentFlags().Lookup("timeout"))
	viper.BindPFlag("kubernetes.request-cpu", serverCmd.PersistentFlags().Lookup("request-cpu"))
	viper.BindPFlag("kubernetes.request-memory", serverCmd.PersistentFlags().Lookup("request-memory"))
	viper.BindPFlag("kubernetes.node-selector", serverCmd.PersistentFlags().Lookup("node-selector"))
	viper.BindPFlag("kubernetes.active-deadline-seconds", serverCmd.PersistentFlags().Lookup("active-deadline-seconds"))
	viper.BindPFlag("kubernetes.runas-user", serverCmd.PersistentFlags().Lookup("runas-user"))
	viper.BindPFlag("registry.inspector", serverCmd.PersistentFlags().Lookup("inspector"))
	viper.BindPFlag("reaper.reapmax", serverCmd.PersistentFlags().Lookup("reapmax"))
	viper.BindPFlag("lock.enabled", serverCmd.PersistentFlags().Lookup("lock"))
	viper.BindPFlag("lock.timeout", serverCmd.PersistentFlags().Lookup("lock-timeout"))
	viper.BindPFlag("verbosity", serverCmd.PersistentFlags().Lookup("verbosity"))
	viper.BindPFlag("prune-start", serverCmd.PersistentFlags().Lookup("prune-start"))
	viper.BindPFlag("port-forward", serverCmd.PersistentFlags().Lookup("port-forward"))
	viper.BindPFlag("reverse-proxy", serverCmd.PersistentFlags().Lookup("reverse-proxy"))
	viper.BindPFlag("pre-archive", serverCmd.PersistentFlags().Lookup("pre-archive"))
	viper.BindPFlag("disable-services", serverCmd.PersistentFlags().Lookup("disable-services"))

	viper.BindEnv("server.listen-addr", "SERVER_LISTEN_ADDR")
	viper.BindEnv("server.tls-enable", "SERVER_TLS_ENABLE")
	viper.BindEnv("server.tls-cert-file", "SERVER_TLS_CERT_FILE")
	viper.BindEnv("server.tls-key-file", "SERVER_TLS_KEY_FILE")
	viper.BindEnv("kubernetes.namespace", "NAMESPACE")
	viper.BindEnv("kubernetes.initimage", "INIT_IMAGE")
	viper.BindEnv("kubernetes.dindimage", "DIND_IMAGE")
	viper.BindEnv("kubernetes.disable-dind", "DISABLE_DIND")
	viper.BindEnv("kubernetes.pull-policy", "PULL_POLICY")
	viper.BindEnv("kubernetes.service-account", "SERVICE_ACCOUNT")
	viper.BindEnv("kubernetes.image-pull-secrets", "IMAGE_PULL_SECRETS")
	viper.BindEnv("kubernetes.pod-template", "POD_TEMPLATE")
	viper.BindEnv("kubernetes.pod-name-prefix", "POD_NAME_PREFIX")
	viper.BindEnv("kubernetes.timeout", "TIME_OUT")
	viper.BindEnv("kubernetes.request-cpu", "K8S_REQUEST_CPU")
	viper.BindEnv("kubernetes.request-memory", "K8S_REQUEST_MEMORY")
	viper.BindEnv("kubernetes.node-selector", "K8S_NODE_SELECTOR")
	viper.BindEnv("kubernetes.active-deadline-seconds", "K8S_ACTIVE_DEADLINE_SECONDS")
	viper.BindEnv("kubernetes.runas-user", "K8S_RUNAS_USER")
	viper.BindEnv("kubernetes.timeout", "TIME_OUT")
	viper.BindEnv("reaper.reapmax", "REAPER_REAPMAX")
	viper.BindEnv("verbosity", "VERBOSITY")

	serverCmd.PersistentFlags().Lookup("tls-enable").Hidden = true
	serverCmd.PersistentFlags().Lookup("tls-key-file").Hidden = true
	serverCmd.PersistentFlags().Lookup("tls-cert-file").Hidden = true

	serverCmd.PersistentFlags().StringArrayVar(&annotations, "annotation", []string{}, "annotation that need to be added to every k8s resource (key=value)")
	serverCmd.PersistentFlags().StringArrayVar(&labels, "label", []string{}, "label that need to be added to every k8s resource (key=value)")

	// kubeconfig
	if home := homeDir(); home != "" {
		serverCmd.PersistentFlags().String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		serverCmd.PersistentFlags().String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	viper.BindPFlag("kubernetes.kubeconfig", serverCmd.PersistentFlags().Lookup("kubeconfig"))
}

// addDefaultLabels will add configured default labels (env or cli) to the
// set of labels that need to be added to all containers instantiated by
// this kubedock instance.
func addDefaultLabels(labels []string) {
	labels = append(getEnvVariables("K8S_LABEL_"), labels...)
	for _, label := range labels {
		key, value, found := strings.Cut(label, "=")
		if !found {
			klog.Errorf("could not label %s", label)
			continue
		}
		klog.Infof("adding %s with %s", key, value)
		config.AddDefaultLabel(key, value)
	}
}

// addDefaultAnnotations will add configured default annotations (env or cli)
// to the set of annotations that need to be added to all containers
// instantiated by this kubedock instance.
func addDefaultAnnotations(annotations []string) {
	annotations = append(getEnvVariables("K8S_ANNOTATION_"), annotations...)
	for _, annotation := range annotations {
		key, value, found := strings.Cut(annotation, "=")
		if !found {
			klog.Errorf("could not annotation %s", annotation)
			continue
		}
		klog.Infof("adding %s with %s", key, value)
		config.AddDefaultAnnotation(key, value)
	}
}

// getEnvVariables will return a list of values from environment
// variables that start with the given prefix.
func getEnvVariables(prefix string) []string {
	var envs []string
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, prefix) {
			key, value, _ := strings.Cut(env, "=")
			key = strings.ToLower(strings.TrimPrefix(key, prefix))
			envs = append(envs, key+"="+value)
		}
	}
	return envs
}

// homeDir returns the current home directory of the user.
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// getContextNamespace will return the namespace that is either available
// in the magic serviceaccount location, or is set in the current
// kubeconfig context, and returns 'default' if none is set.
func getContextNamespace() string {
	res := "default"
	ns, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err == nil {
		return strings.TrimSpace(string(ns))
	}
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
