package tar

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"io"

	"github.com/ulikunitz/xz"
)

// Reader is able to read tar archive uncompressed, compressed with gzip, xz, or bzip2
type Reader struct {
	concatReader *ConcatReader
	tr           *tar.Reader
	close        func() error
}

// NewReader initializes a new Reader instance to process compressed or uncompressed archive data.
// It takes an `io.Reader` as input, reads the first 5 bytes to detect the compression type,
// and returns a `*Reader` configured to handle the detected compression format (e.g., gzip,
// bzip2, xz, or none).
func NewReader(reader io.Reader) (r *Reader, err error) {
	first5Bytes := make([]byte, 5)
	_, err = reader.Read(first5Bytes)
	if err != nil {
		if err != io.EOF {
			return
		}
	}
	r = &Reader{
		concatReader: NewConcatReader(first5Bytes, reader),
	}
	switch detectCompressionType(first5Bytes) {
	case Gzip:
		zr, err := gzip.NewReader(r.concatReader)
		if err != nil {
			return nil, err
		}
		r.close = zr.Close
		r.tr = tar.NewReader(zr)
	case Bzip2:
		r.tr = tar.NewReader(bzip2.NewReader(r.concatReader))
	case Xz:
		xzr, err := xz.NewReader(r.concatReader)
		if err != nil {
			return nil, err
		}
		r.tr = tar.NewReader(xzr)
	default:
		r.tr = tar.NewReader(r.concatReader)
	}
	return r, nil
}

// Next advances to the next entry in the tar archive.
// The Header.Size determines how many bytes can be read for the next file.
// Any remaining data in the current file is automatically discarded.
// At the end of the archive, Next returns the error io.EOF.
//
// If Next encounters a non-local name (as defined by [filepath.IsLocal])
// and the GODEBUG environment variable contains `tarinsecurepath=0`,
// Next returns the header with an [ErrInsecurePath] error.
// A future version of Go may introduce this behavior by default.
// Programs that want to accept non-local names can ignore
// the [ErrInsecurePath] error and use the returned header.
func (r *Reader) Next() (*tar.Header, error) {
	return r.tr.Next()
}

// Read reads from the current file in the tar archive.
// It returns (0, io.EOF) when it reaches the end of that file,
// until [Next] is called to advance to the next file.
//
// If the current file is sparse, then the regions marked as a hole
// are read back as NUL-bytes.
//
// Calling Read on special types like [TypeLink], [TypeSymlink], [TypeChar],
// [TypeBlock], [TypeDir], and [TypeFifo] returns (0, [io.EOF]) regardless of what
// the [Header.Size] claims.
func (r *Reader) Read(p []byte) (int, error) {
	return r.tr.Read(p)
}

// ReadBytes returns the number of read bytes
func (r *Reader) ReadBytes() int {
	return r.concatReader.ReadBytes()
}

// Close closes the reader for further reading and returns an error on failure.
func (r *Reader) Close() error {
	if r.close != nil {
		return r.close()
	}
	return nil
}
