package config

import (
	"runtime"
)

const (
	// ID is the id as advertised when calling /info
	ID = "com.joyrex2001.kubedock"
	// Name is the name as advertised when calling /info
	Name = "kubedock"
	// OS is the operating system as advertised when calling /info
	OS = "kubernetes"
	// DockerVersion is the docker version as advertised when calling /version
	DockerVersion = "1.25"
	// DockerMinAPIVersion is the minimum docker version as advertised when calling /version
	DockerMinAPIVersion = "1.25"
	// DockerAPIVersion is the api version as advertised when calling /version
	DockerAPIVersion = "1.25"
	// LibpodAPIVersion is the api version as advertised in libpod rest calls
	LibpodAPIVersion = "4.2.0"
)

var (
	// GoVersion is the version of go as advertised when calling /version
	GoVersion = runtime.Version()
	// GOOS is the runtime operating system as advertised when calling /version
	GOOS = runtime.GOOS
	// GOARCH is runtime architecture of go as advertised when calling /version
	GOARCH = runtime.GOARCH
)
