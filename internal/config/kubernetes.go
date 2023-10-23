package config

import (
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// enable auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/joyrex2001/kubedock/internal/util/stringid"
)

// SystemLabels are the labels that are added to every kubedock
// managed k8s resource and which should not be altered.
var SystemLabels = map[string]string{
	"kubedock":    "true",
	"kubedock.id": "",
}

// DefaultLabels are the labels that are added to every kubedock
// managed k8s resource.
var DefaultLabels = map[string]string{}

// DefaultAnnotations are the annotations that are added to every
// kubedock managed k8s resource.
var DefaultAnnotations = map[string]string{}

// InstanceID contains an unique ID to identify this running instance.
var InstanceID = ""

// init will set an unique instance id in the default labels to identify
// this speciffic instance of kubedock.
func init() {
	InstanceID = stringid.TruncateID(stringid.GenerateRandomID())
	SystemLabels["kubedock.id"] = InstanceID
}

// AddDefaultLabel will add a label that will be added to all containers
// started by this kubedock instance.
func AddDefaultLabel(key, value string) {
	DefaultLabels[key] = value
}

// AddDefaultAnnotation will add an annotation that will be added to all
// containers started by this kubedock instance.
func AddDefaultAnnotation(key, value string) {
	DefaultAnnotations[key] = value
}

// GetKubernetes will return a kubernetes config object.
func GetKubernetes() (*rest.Config, error) {
	var err error
	config := &rest.Config{}
	kubeconfig := viper.GetString("kubernetes.kubeconfig")
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if kubeconfig == "" || err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}
