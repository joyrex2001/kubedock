package backend

import (
	"context"
	"io"
	"sync"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/util/attach"
	"github.com/joyrex2001/kubedock/internal/util/ioproxy"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AttachContainer will attach to a container and stream stdin/stdout/stderr.
func (in *instance) AttachContainer(tainr *types.Container, stdin io.Reader, stdout io.Writer, stderr io.Writer, tty bool) error {
	pod, err := in.cli.CoreV1().Pods(in.namespace).Get(context.Background(), tainr.GetPodName(), v1.GetOptions{})
	if err != nil {
		return err
	}

	req := attach.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        *pod,
		Container:  "main",
		TTY:        tty,
	}

	if stdin != nil {
		req.Stdin = stdin
	}

	// Attach uses same I/O multiplexing logic as ExecContainer
	if tty {
		req.Stdout = stdout
		req.Stderr = io.Discard
	} else {
		lock := sync.Mutex{}
		if stdout != nil {
			iop := ioproxy.New(stdout, ioproxy.Stdout, &lock)
			req.Stdout = iop
			defer iop.Flush()
		}
		if stderr != nil {
			iop := ioproxy.New(stdout, ioproxy.Stderr, &lock)
			req.Stderr = iop
			defer iop.Flush()
		}
	}

	return attach.RemoteAttach(req)
}
