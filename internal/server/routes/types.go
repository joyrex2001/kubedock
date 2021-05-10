package routes

// ContainerCreateRequest represents the json structure that
// is used for the /container/create post endpoint.
type ContainerCreateRequest struct {
	Name         string                 `json:"name"`
	Image        string                 `json:"image"`
	ExposedPorts map[string]interface{} `json:"ExposedPorts"`
	Labels       map[string]string      `json:"Labels"`
	Cmd          []string               `json:"Cmd"`
	Env          []string               `json:"Env"`
	HostConfig   ContainerHostConfig    `json:"HostConfig"`
}

// ContainerHostConfig contains to be mounted files from the host system.
type ContainerHostConfig struct {
	Binds []string `json:"Binds"`
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
	Container string `json:"container"`
}

// NetworkDisconnectRequest represents the json structure that
// is used for the /networks/:id/disconnect post endpoint.
type NetworkDisconnectRequest struct {
	Container string `json:"container"`
}
