package backend

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/util/exec"
	"k8s.io/klog"
)

// CopyToContainer will copy given (tar) archive to given path of the container.
func (in *instance) CopyToContainer(tainr *types.Container, archive []byte, target string) error {
	pods, err := in.getPods(tainr)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("no matching pod found")
	}

	if target != "/" && strings.HasSuffix(string(target[len(target)-1]), "/") {
		target = target[:len(target)-1]
	}

	reader, writer := io.Pipe()
	go func() {
		writer.Write(archive)
		writer.Close()
	}()

	klog.Infof("copy %d bytes to %s:%s", len(archive), tainr.ShortID, target)

	return exec.RemoteCmd(exec.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        pods[0],
		Container:  "main",
		Cmd:        []string{"tar", "-xf", "-", "-C", target},
		Stdin:      reader,
		Stderr:     writer,
	})
}

// CopyFromContainer will copy given path from the container as return it as a
// tar archive. Note that this requires tar to be present on the container.
func (in *instance) CopyFromContainer(tainr *types.Container, target string) ([]byte, error) {
	pods, err := in.getPods(tainr)
	if err != nil {
		return nil, err
	}
	if len(pods) == 0 {
		return nil, fmt.Errorf("no matching pod found")
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	err = exec.RemoteCmd(exec.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        pods[0],
		Container:  "main",
		Cmd:        []string{"tar", "-cf", "-", "-C", path.Dir(target), path.Base(target)},
		Stdout:     writer,
	})
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
