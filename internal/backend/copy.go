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

	err = exec.RemoteCmd(exec.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        *pod,
		Container:  "main",
		Cmd:        []string{"sh", "-c", "if [ -d \"" + sanitizeFilename(target) + "\" ]; then echo folder; else echo file; fi"},
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

// FileExistsInContainer will check if the file exists in the container.
func (in *instance) FileExistsInContainer(tainr *types.Container, target string) (bool, error) {
	pod, err := in.cli.CoreV1().Pods(in.namespace).Get(context.Background(), tainr.GetPodName(), metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	err = exec.RemoteCmd(exec.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        *pod,
		Container:  "main",
		Cmd:        []string{"sh", "-c", "if [ -e \"" + sanitizeFilename(target) + "\" ]; then echo true; else echo false; fi"},
		Stdout:     writer,
	})

	if err != nil {
		return false, err
	}

	exists := false
	if strings.Contains(string(b.Bytes()), "true") {
		exists = true
	}

	return exists, nil
}

// sanitizeFilename will clean up unwanted characters from the filename to
// prevent injection attacks.
func sanitizeFilename(file string) string {
	file = strings.ReplaceAll(file, "`", "")
	file = strings.ReplaceAll(file, "$", "")
	file = strings.ReplaceAll(file, "\"", "\\\"")
	return file
}
