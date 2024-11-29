package tar

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestConcatReader(t *testing.T) {
	data := make([]byte, 10)
	_, err := rand.Read(data)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// when bytes slice is empty
	reader := NewConcatReader(nil, bytes.NewReader(data))
	actual, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if !bytes.Equal(data, actual) {
		t.Errorf("data mismatch: expected %s, got %s", string(data), string(actual))
	}
	if reader.ReadBytes() != len(data) {
		t.Errorf("read bytes mismatch: expected %d, got %d", len(data), reader.ReadBytes())
	}

	// when bytes slice is not empty
	reader = NewConcatReader(data[:4], bytes.NewReader(data[4:]))
	actual, err = io.ReadAll(reader)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if !bytes.Equal(data, actual) {
		t.Errorf("data mismatch: expected %s, got %s", string(data), string(actual))
	}
	if reader.ReadBytes() != len(data) {
		t.Errorf("read bytes mismatch: expected %d, got %d", len(data), reader.ReadBytes())
	}
}
