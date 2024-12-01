package tar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"github.com/davecgh/go-spew/spew"
	"github.com/dsnet/compress/bzip2"
	"github.com/ulikunitz/xz"
	"io"
	"testing"
)

func TestReader(t *testing.T) {
	// write tar archive
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	filename := "some.file"
	if err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     filename,
		Size:     10,
	}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	data := make([]byte, 10)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	archive := make([]byte, buf.Len())
	copy(archive, buf.Bytes())

	// read uncompressed archive
	tr, err := NewReader(bytes.NewReader(archive))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assertTarContent(t, tr, filename, data)

	// read gzip archive
	buf.Reset()
	gz := gzip.NewWriter(buf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if _, err = gz.Write(archive); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	tr, err = NewReader(buf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assertTarContent(t, tr, filename, data)

	// read bzip2 archive
	buf.Reset()
	bz, err := bzip2.NewWriter(buf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if _, err = bz.Write(archive); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err := bz.Close(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	tr, err = NewReader(buf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assertTarContent(t, tr, filename, data)

	// read xz archive
	buf.Reset()
	xzw, err := xz.NewWriter(buf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if _, err = xzw.Write(archive); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err = xzw.Close(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	tr, err = NewReader(buf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assertTarContent(t, tr, filename, data)
}

func assertTarContent(t *testing.T, tr *Reader, filename string, fileContent []byte) {
	t.Helper()
	hdr, err := tr.Next()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if hdr.Name != filename {
		t.Errorf("filename mismatch: expected %s, got %s", filename, hdr.Name)
	}
	data := make([]byte, hdr.Size)
	_, err = tr.Read(data)
	if err != nil && err != io.EOF {
		t.Fatalf("unexpected error: %s", err)
	}
	if !bytes.Equal(data, fileContent) {
		t.Errorf("fileContent mismatch: expected %s, got %s", spew.Sdump(fileContent), spew.Sdump(data))
	}
}
