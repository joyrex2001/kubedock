package backend

import (
	"context"
	"io"
	"strconv"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/util/exec"
	"github.com/joyrex2001/kubedock/internal/util/ioproxy"
)

// ExecContainer will execute given exec object in kubernetes.
func (in *instance) ExecContainer(tainr *types.Container, ex *types.Exec, stdin io.Reader, stdout io.Writer) (int, error) {
	pod, err := in.cli.CoreV1().Pods(in.namespace).Get(context.Background(), tainr.GetPodName(), metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	req := exec.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        *pod,
		Container:  "main",
		Cmd:        ex.Cmd,
		TTY:        ex.TTY,
	}

	if ex.Stdin {
		req.Stdin = stdin
	}
	if ex.TTY {
		req.Stdout = stdout
		req.Stderr = io.Discard
	} else {
		lock := sync.Mutex{}
		if ex.Stdout {
			iop := ioproxy.New(stdout, ioproxy.Stdout, &lock)
			req.Stdout = iop
			defer iop.Flush()
		}
		if ex.Stderr {
			iop := ioproxy.New(stdout, ioproxy.Stderr, &lock)
			req.Stderr = iop
			defer iop.Flush()
		}
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
