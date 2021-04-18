package config

import (
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// getKubernetes will return a kubernetes config object.
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
