package ioproxy

import (
	"encoding/binary"
	"io"
	"k8s.io/klog"
	"sync"
	"time"
)

// StdType is the type of standard stream
// a writer can multiplex to.
type StdType byte

const (
	// Stdin represents standard input stream type.
	Stdin StdType = iota
	// Stdout represents standard output stream type.
	Stdout
	// Stderr represents standard error steam type.
	Stderr
)

// IoProxy is a proxy writer which adds the output prefix before writing data.
type IoProxy struct {
	io.Writer
	out     io.Writer
	prefix  StdType
	buf     []byte
	flusher bool
	lock    *sync.Mutex
}

// New will return a new IoProxy instance.
func New(w io.Writer, prefix StdType, lock *sync.Mutex) *IoProxy {
	return &IoProxy{
		out:    w,
		prefix: prefix,
		buf:    []byte{},
		lock:   lock,
	}
}

// Write will write given data to the an internal buffer, which will be
// flushed if a newline is encountered, of when the maximum size of the
// buffer has been reached.
func (w *IoProxy) Write(p []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.buf = append(w.buf, p...)
	for w.process() != 0 {
	}
	if len(w.buf) > 0 && !w.flusher {
		w.flusher = true
		go func() {
			time.Sleep(100 * time.Millisecond)
			w.Flush()
		}()
	}
	return len(p), nil
}

func (w *IoProxy) writeAll(writer io.Writer, buf []byte) error {
	for len(buf) > 0 {
		n, err := writer.Write(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]
	}
	return nil
}

// process iterates over the buffer and writes chunks that end with
// a newline character to the output writer.
func (w *IoProxy) process() int {
	bufferLength := len(w.buf)
	// Iterate over the buffer to find newline characters
	for pos := 0; pos < bufferLength; pos++ {
		if w.buf[pos] == '\n' {
			w.write(w.buf[:pos+1])
			w.buf = w.buf[pos+1:]
			return pos + 1
		}
	}
	return 0
}

// write will write data to the configured writer, using the correct header.
func (w *IoProxy) write(p []byte) error {
	header := [8]byte{}
	header[0] = byte(w.prefix)
	binary.BigEndian.PutUint32(header[4:], uint32(len(p)))
	err := w.writeAll(w.out, header[:])
	if err != nil {
		klog.Errorf("Error when writing docker log header: %v", err)
		return err
	}
	err = w.writeAll(w.out, p)
	if err != nil {
		klog.Errorf("Error ehen writing docker log content: %v", err)
	}
	return err
}

// Flush will write all buffer data still present.
func (w *IoProxy) Flush() error {
	w.lock.Lock()
	defer w.lock.Unlock()
	if len(w.buf) == 0 {
		return nil
	}
	err := w.write(w.buf)
	w.buf = []byte{}
	w.flusher = false
	return err
}
