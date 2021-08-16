package routes

// ContainerCreateRequest represents the json structure that
// is used for the /container/create post endpoint.
type ContainerCreateRequest struct {
	Name          string                 `json:"name"`
	Image         string                 `json:"image"`
	ExposedPorts  map[string]interface{} `json:"ExposedPorts"`
	Labels        map[string]string      `json:"Labels"`
	Entrypoint    []string               `json:"Entrypoint"`
	Cmd           []string               `json:"Cmd"`
	Env           []string               `json:"Env"`
	HostConfig    HostConfig             `json:"HostConfig"`
	NetworkConfig NetworkingConfig       `json:"NetworkingConfig"`
}

// ContainerExecRequest represents the json structure that
// is used for the /conteiner/:id/exec request.
type ContainerExecRequest struct {
	Cmd    []string `json:"Cmd"`
	Stdin  bool     `json:"AttachStdin"`
	Stdout bool     `json:"AttachStdout"`
	Stderr bool     `json:"AttachStderr"`
	Tty    bool     `json:"Tty"`
	Env    []string `json:"Env"`
}

// ExecStartRequest represents the json structure that is
// used for the /exec/:id/start request.
type ExecStartRequest struct {
	Detach bool `json:"Detach"`
	Tty    bool `json:"Tty"`
}

// NetworkCreateRequest represents the json structure that
// is used for the /networks/create post endpoint.
type NetworkCreateRequest struct {
	Name string `json:"name"`
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
	PortBindings map[string][]PortBinding
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
	Aliases []string `json:"Aliases"`
}
