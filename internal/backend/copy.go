package backend

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"io/fs"
	"path"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/util/exec"
)

// CopyToContainer will copy given (tar) archive to given path of the container.
func (in *instance) CopyToContainer(tainr *types.Container, reader io.Reader, target string) error {
	pod, err := in.cli.CoreV1().Pods(in.namespace).Get(context.Background(), tainr.GetPodName(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	if target != "/" && strings.HasSuffix(string(target[len(target)-1]), "/") {
		target = target[:len(target)-1]
	}

	klog.Infof("copy archive to %s:%s", tainr.ShortID, target)

	return exec.RemoteCmd(exec.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        *pod,
		Container:  "main",
		Cmd:        []string{"tar", "-xf", "-", "-C", target},
		Stdin:      reader,
	})
}

// CopyFromContainer will copy given path from the container and return the
// contents as a tar archive through the given writer. Note that this requires
// tar to be present on the container.
func (in *instance) CopyFromContainer(tainr *types.Container, target string, writer io.Writer) error {
	pod, err := in.cli.CoreV1().Pods(in.namespace).Get(context.Background(), tainr.GetPodName(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	klog.Infof("copy archive from %s to %s", tainr.ShortID, target)

	return exec.RemoteCmd(exec.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        *pod,
		Container:  "main",
		Cmd:        []string{"tar", "-cf", "-", "-C", path.Dir(target), path.Base(target)},
		Stdout:     writer,
	})
}

// GetFileModeInContainer will return the file mode (directory or file) of a given path
// inside the container.
func (in *instance) GetFileModeInContainer(tainr *types.Container, target string) (fs.FileMode, error) {
	pod, err := in.cli.CoreV1().Pods(in.namespace).Get(context.Background(), tainr.GetPodName(), metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	target = strings.ReplaceAll(target, "`", "")
	target = strings.ReplaceAll(target, "$", "")
	target = strings.ReplaceAll(target, "\"", "\\\"")

	err = exec.RemoteCmd(exec.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        *pod,
		Container:  "main",
		Cmd:        []string{"sh", "-c", "if [ -d \"" + target + "\" ]; then echo folder; else echo file; fi"},
		Stdout:     writer,
	})
	if err != nil {
		return 0, err
	}

	mode := fs.FileMode(fs.ModePerm)
	if strings.Contains(string(b.Bytes()), "folder") {
		mode |= fs.ModeDir
	}

	return mode, nil
}
