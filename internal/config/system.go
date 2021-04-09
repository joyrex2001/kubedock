package config

import (
	"runtime"
)

var (
	ID   = "com.joyrex2001.donk"
	Name = "donk"
	OS   = "kubernetes"

	GoVersion = runtime.Version()
	GOOS      = runtime.GOOS
	GOARCH    = runtime.GOARCH

	DockerVersion    = "1.0"
	DockerAPIVersion = "1.0"
)
