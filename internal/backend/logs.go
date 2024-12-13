package backend

import (
	"context"
	"io"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/util/ioproxy"
)

// LogOptions describe the supported log options
type LogOptions struct {
	// Keep connection after returning logs.
	Follow bool
	// Only return logs since this time, as a UNIX timestamp
	SinceTime *time.Time
	// Add timestamps to every log line
	Timestamps bool
	// Number of lines to show from the end of the logs
	TailLines *uint64
}

// GetLogs will write the logs for given container to given writer.
func (in *instance) GetLogs(tainr *types.Container, opts *LogOptions, stop chan struct{}, w io.Writer) error {
	options := newPodLogOptions(opts)

	_, err := in.cli.CoreV1().Pods(in.namespace).Get(context.Background(), tainr.GetPodName(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	req := in.cli.CoreV1().Pods(in.namespace).GetLogs(tainr.GetPodName(), &options)
	stream, err := req.Stream(context.Background())
	if err != nil {
		return err
	}
	defer stream.Close()

	stopL := make(chan struct{}, 1)

	if opts.Follow {
		go func() {
			<-stop
			stopL <- struct{}{}
			stream.Close()
		}()
	}

	out := ioproxy.New(w, ioproxy.Stdout, &sync.Mutex{})
	defer out.Flush()
	for {
		// close when container is done
		select {
		case <-stopL:
			close(stopL)
			return nil
		default:
		}
		// read log input (blocking read)
		buf := make([]byte, 255)
		n, err := stream.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if n == 0 {
			if !opts.Follow {
				break
			}
			continue
		}
		// write log to output
		if n, err = out.Write(buf[:n]); n == 0 || err != nil {
			break
		}
	}

	return nil
}

func newPodLogOptions(opts *LogOptions) v1.PodLogOptions {
	var sinceTime *metav1.Time = nil
	if opts.SinceTime != nil {
		t := metav1.NewTime(*opts.SinceTime)
		sinceTime = &t
	}

	var tailLines *int64 = nil
	if opts.TailLines != nil {
		l := int64(*opts.TailLines)
		tailLines = &l
	}

	return v1.PodLogOptions{
		Container:  "main",
		Follow:     opts.Follow,
		TailLines:  tailLines,
		SinceTime:  sinceTime,
		Timestamps: opts.Timestamps,
	}
}
