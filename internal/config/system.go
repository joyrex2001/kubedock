package config

import (
	"runtime"
)

var (
	ID   = "com.joyrex2001.kubedock"
	Name = "kubedock"
	OS   = "kubernetes"

	GoVersion = runtime.Version()
	GOOS      = runtime.GOOS
	GOARCH    = runtime.GOARCH

	DockerVersion     = "1.18"
	DockerAPIVersion  = "1.18"
	DockerMaxAPIMinor = 40
)
