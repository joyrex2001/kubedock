package tar

import "io"

// ConcatReader is an [io.Reader] that first returns data from a wrapped bytes slice and then continues to return data
// from a wrapped [io.Reader]
type ConcatReader struct {
	data   []byte
	offset int
	reader io.Reader
}

// NewConcatReader creates a new ConcatReader instance that sequentially reads data from a provided byte slice
// and then continues reading from an underlying io.Reader.
func NewConcatReader(data []byte, reader io.Reader) *ConcatReader {
	return &ConcatReader{data: data, reader: reader}
}

func (r *ConcatReader) Read(p []byte) (int, error) {
	if r.offset >= len(r.data) {
		n, err := r.reader.Read(p)
		r.offset += n
		return n, err
	}
	n := copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

// ReadBytes returns the number of read bytes
func (r *ConcatReader) ReadBytes() int {
	return r.offset
}
