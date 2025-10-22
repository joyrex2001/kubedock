package docker

// ContainerCreateRequest represents the json structure that
// is used for the /container/create post endpoint.
type ContainerCreateRequest struct {
	Name          string                 `json:"name"`
	Hostname      string                 `json:"Hostname"`
	Image         string                 `json:"image"`
	ExposedPorts  map[string]interface{} `json:"ExposedPorts"`
	Labels        map[string]string      `json:"Labels"`
	Entrypoint    []string               `json:"Entrypoint"`
	Cmd           []string               `json:"Cmd"`
	Env           []string               `json:"Env"`
	User          string                 `json:"User"`
	HostConfig    HostConfig             `json:"HostConfig"`
	NetworkConfig NetworkingConfig       `json:"NetworkingConfig"`
	TTY           bool                   `json:"Tty"`
	AttachStdin   bool                   `json:"AttachStdin"`
	AttachStdout  bool                   `json:"AttachStdout"`
	AttachStderr  bool                   `json:"AttachStderr"`
}

// NetworkCreateRequest represents the json structure that
// is used for the /networks/create post endpoint.
type NetworkCreateRequest struct {
	Name   string            `json:"Name"`
	Labels map[string]string `json:"Labels"`
}

// NetworkConnectRequest represents the json structure that
// is used for the /networks/:id/connect post endpoint.
type NetworkConnectRequest struct {
	Container      string         `json:"container"`
	EndpointConfig EndpointConfig `json:"EndpointConfig"`
}

// NetworkDisconnectRequest represents the json structure that
// is used for the /networks/:id/disconnect post endpoint.
type NetworkDisconnectRequest struct {
	Container string `json:"container"`
}

// HostConfig contains to be mounted files from the host system.
type HostConfig struct {
	Binds        []string `json:"Binds"`
	Mounts       []Mount  `json:"Mounts"`
	PortBindings map[string][]PortBinding
	Memory       int    `json:"Memory"`
	NanoCpus     int    `json:"NanoCpus"`
	NetworkMode  string `json:"NetworkMode"`
}

// PortBinding represents a binding between to a port
type PortBinding struct {
	HostPort string `json:"HostPort"`
}

// NetworkingConfig contains network configuration
type NetworkingConfig struct {
	EndpointsConfig map[string]EndpointConfig `json:"EndpointsConfig"`
}

// NetworkConfig contains network configuration
type NetworkConfig struct {
	EndpointConfig EndpointConfig `json:"EndpointConfig"`
}

// EndpointConfig contains information about network endpoints
type EndpointConfig struct {
	Aliases   []string `json:"Aliases"`
	NetworkID string   `json:"NetworkID"`
}

// Mount contains information about mounted volumes/bindings
type Mount struct {
	Type     string `json:"Type"`
	Source   string `json:"Source"`
	Target   string `json:"Target"`
	ReadOnly bool   `json:"ReadOnly"`
}
