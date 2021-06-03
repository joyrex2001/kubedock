package config

import (
	"runtime"
)

var (
	// ID is the id as advertised when calling /info
	ID = "com.joyrex2001.kubedock"
	// Name is the name as advertised when calling /info
	Name = "kubedock"
	// OS is the operating system as advertised when calling /info
	OS = "kubernetes"

	// GoVersion is the version of go as advertised when calling /version
	GoVersion = runtime.Version()
	// GOOS is the runtime operating system as advertised when calling /version
	GOOS = runtime.GOOS
	// GOARCH is runtime architecture of go as advertised when calling /version
	GOARCH = runtime.GOARCH

	// DockerVersion is the docker version as advertised when calling /version
	DockerVersion = "1.22"
	// DockerAPIVersion is the api version as advertised when calling /version
	DockerAPIVersion = "1.22"
)
