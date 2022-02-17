package ioproxy

import (
	"encoding/binary"
	"io"
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

// logProxy is a proxy writer which adds the output prefix before writing data.
type IoProxy struct {
	io.Writer
	out     io.Writer
	prefix  StdType
	buf     []byte
	flusher bool
	lock    sync.Mutex
}

// New will return a new logproxy instance.
func New(w io.Writer, prefix StdType) *IoProxy {
	return &IoProxy{out: w, prefix: prefix, buf: []byte{}}
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

// process will go through the buffer and writes chunks that end with
// a newline to the output writer.
func (w *IoProxy) process() int {
	for pos := 0; pos < len(w.buf)-1; pos++ {
		if w.buf[pos] == 10 { // write chunk if newline char is found
			w.write(w.buf[:pos+1])
			w.buf = w.buf[pos+1:]
			return pos + 1
		}
	}
	return 0
}

// write will write data to the configured writer, using the correct header.
func (w *IoProxy) write(p []byte) (int, error) {
	header := [8]byte{}
	header[0] = byte(w.prefix)
	binary.BigEndian.PutUint32(header[4:], uint32(len(p)))
	w.out.Write(header[:])
	return w.out.Write(p)
}

// Flush will write all buffer data still present.
func (w *IoProxy) Flush() error {
	w.lock.Lock()
	defer w.lock.Unlock()
	_, err := w.write(w.buf)
	w.buf = []byte{}
	w.flusher = false
	return err
}
