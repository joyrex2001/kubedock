package backend

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/util/exec"
	"k8s.io/klog"
)

// CopyToContainer will copy given (tar) archive to given path of the container.
func (in *instance) CopyToContainer(tainr *types.Container, archive []byte, path string) error {
	pods, err := in.getPods(tainr)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("no matching pod found")
	}

	if path != "/" && strings.HasSuffix(string(path[len(path)-1]), "/") {
		path = path[:len(path)-1]
	}

	reader, writer := io.Pipe()
	go func() {
		writer.Write(archive)
		writer.Close()
	}()

	klog.Infof("copy %d bytes to %s:%s", len(archive), tainr.ShortID, path)

	return exec.RemoteCmd(exec.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        pods[0],
		Container:  "main",
		Cmd:        []string{"tar", "-xf", "-", "-C", path},
		Stdin:      reader,
		Stderr:     writer,
	})
}

// CopyFromContainer will copy given path from the container as return it as a
// tar archive. Note that this requires tar to be present on the container.
func (in *instance) CopyFromContainer(tainr *types.Container, path string) ([]byte, error) {
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
		Cmd:        []string{"tar", "-cf", "-", path},
		Stdout:     writer,
	})
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
