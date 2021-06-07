package ioproxy

import (
	"encoding/binary"
	"io"
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
type logProxy struct {
	io.Writer
	out    io.Writer
	prefix StdType
}

// New will return a new logproxy instance.
func New(w io.Writer, prefix StdType) io.Writer {
	return &logProxy{out: w, prefix: prefix}
}

// Write will write data to the configured writer, using the correct header.
func (w *logProxy) Write(p []byte) (int, error) {
	header := [8]byte{}
	header[0] = byte(w.prefix)
	binary.BigEndian.PutUint32(header[4:], uint32(len(p)))
	w.out.Write(header[:])
	return w.out.Write(p)
}
