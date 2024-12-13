package ioproxy

import (
	"bytes"
	"sync"
	"testing"
)

type ShortWriteBuffer struct {
	bytes.Buffer
}

func (buf *ShortWriteBuffer) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return buf.Buffer.Write(b)
	}
	return buf.Buffer.Write(b[:1])
}

func (buf *ShortWriteBuffer) Bytes() []byte {
	return buf.Buffer.Bytes()
}

type TestBuffer interface {
	Write(b []byte) (int, error)
	Bytes() []byte
}

func TestWrite(t *testing.T) {
	tests := []struct {
		write string
		read  []byte
		flush []byte
	}{
		{
			write: "hello\n\nto the bat-mobile\nlet's go",
			read: []byte{
				0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x6, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0xa,
				0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0xa,
				0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x12, 0x74, 0x6f, 0x20, 0x74, 0x68, 0x65, 0x20, 0x62, 0x61, 0x74, 0x2d, 0x6d, 0x6f, 0x62, 0x69, 0x6c, 0x65, 0xa,
			},
			flush: []byte{
				0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x6, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0xa,
				0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0xa,
				0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x12, 0x74, 0x6f, 0x20, 0x74, 0x68, 0x65, 0x20, 0x62, 0x61, 0x74, 0x2d, 0x6d, 0x6f, 0x62, 0x69, 0x6c, 0x65, 0xa,
				0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x8, 0x6c, 0x65, 0x74, 0x27, 0x73, 0x20, 0x67, 0x6f,
			},
		},
	}
	for i, tst := range tests {
		// with manual flush
		writeToBufferfuncName(t, &bytes.Buffer{}, tst, i)
		// with manual flush and short writes
		writeToBufferfuncName(t, &ShortWriteBuffer{}, tst, i)

		// without manual flushing
		// There is no automatic flushing. Automatic caused an issue where data could be written to
		// the gin.Context after the request was finished and the gin.Context was returned to the pool.
		// This causes issues where sometime a length 0 byte array (with 8 byte stream header was written
		// to another connection that reused the gin.Context from the pool.
	}
}

func writeToBufferfuncName(t *testing.T, buf TestBuffer, tst struct {
	write string
	read  []byte
	flush []byte
}, i int) {
	iop := New(buf, Stdout, &sync.Mutex{})
	iop.Write([]byte(tst.write))
	if !bytes.Equal(buf.Bytes(), tst.read) {
		t.Errorf("failed read %d - expected %v, but got %v", i, tst.read, buf.Bytes())
	}
	iop.Flush()
	if !bytes.Equal(buf.Bytes(), tst.flush) {
		t.Errorf("failed flush %d - expected %v, but got %v", i, tst.flush, buf.Bytes())
	}
	if len(iop.buf) > 0 {
		t.Errorf("failed flush %d - buffer not empty...", i)
	}
}

func TestLargeLine(t *testing.T) {
	buf := &bytes.Buffer{}
	iop := New(buf, Stdout, &sync.Mutex{})

	data := make([]byte, 1350)
	for i := 0; i < len(data); i++ {
		data[i] = 65
	}
	data[1349] = 10
	iop.Write(data)
	iop.Flush()

	if len(buf.Bytes()) != 1350+8 {
		t.Errorf("failed large line test - buffer size was not linesize + header (%d) but %d", 1350+8, len(buf.Bytes()))
	}
}
