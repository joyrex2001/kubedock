package common

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
