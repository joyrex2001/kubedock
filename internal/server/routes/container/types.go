package container

type ContainerCreateRequest struct {
	Name         string                 `json:"name"`
	Image        string                 `json:"image"`
	ExposedPorts map[string]interface{} `json:"ExposedPorts"`
	Labels       map[string]string      `json:"Labels"`
	Cmd          []string               `json:"Cmd"`
	Env          []string               `json:"Env"`
	// Mounts
}

type ContainerExecRequest struct {
	Cmd []string `json:"Cmd"`
}

type ExecStartRequest struct {
	Detach bool `json:"Detach"`
	Tty    bool `json:"Tty"`
}
