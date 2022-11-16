package backend

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/util/exec"
	"github.com/joyrex2001/kubedock/internal/util/ioproxy"
)

// ExecContainer will execute given exec object in kubernetes.
func (in *instance) ExecContainer(tainr *types.Container, ex *types.Exec, out io.Writer) (int, error) {
	pods, err := in.getPods(tainr)
	if err != nil {
		return 0, err
	}
	if len(pods) == 0 {
		return 0, fmt.Errorf("no matching pod found")
	}

	req := exec.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        pods[0],
		Container:  "main",
		Cmd:        ex.Cmd,
	}

	if ex.Stdout {
		iop := ioproxy.New(out, ioproxy.Stdout)
		req.Stdout = iop
		defer iop.Flush()
	}
	if ex.Stderr {
		iop := ioproxy.New(out, ioproxy.Stderr)
		req.Stderr = iop
		defer iop.Flush()
	}

	err = exec.RemoteCmd(req)
	return in.parseExecResponse(err)
}

// parseExecResponse will take the given error and will parse the string to
// get an exit code from it. if no exit code is found, it will return 0 and
// the original error.
func (in *instance) parseExecResponse(err error) (int, error) {
	if err == nil {
		return 0, err
	}

	const eterm = "command terminated with exit code"
	if !strings.Contains(err.Error(), eterm) {
		return 0, err
	}

	cod, cerr := strconv.Atoi(strings.TrimPrefix(err.Error(), eterm+" "))
	if cerr != nil {
		return 0, err
	}

	return cod, nil
}
