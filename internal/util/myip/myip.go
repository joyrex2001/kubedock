package myip

import (
	"net"
	"os"

	"k8s.io/klog"
)

// Get returns the IP address of the pod if running in Kubernetes, or
// the IP address of the host's network interface.
func Get() (string, error) {
	podIP := os.Getenv("POD_IP")
	if podIP != "" {
		return podIP, nil
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1", err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			klog.V(2).Infof("error getting addresses for interface %s: %s", iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if !ipNet.IP.IsLoopback() && !ipNet.IP.IsLinkLocalUnicast() {
					return ipNet.IP.String(), nil
				}
			}
		}
	}

	return "127.0.0.1", nil
}
