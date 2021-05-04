package kubernetes

import (
	"fmt"
	"io"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/util/exec"
)

// ExecContainer will execute given exec object in kubernetes.
func (in *instance) ExecContainer(tainr *types.Container, ex *types.Exec, out io.Writer) error {
	pods, err := in.getPods(tainr)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("no matching pod found")
	}

	req := exec.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        pods[0],
		Container:  "main",
		Cmd:        ex.Cmd,
	}

	if ex.Stdout {
		req.Stdout = out
	}
	if ex.Stderr {
		req.Stderr = out
	}

	return exec.RemoteCmd(req)
}
