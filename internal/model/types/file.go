package types

import (
	"bytes"
	"os"
)

// File describes the details of a file (content and mode).
type File struct {
	FileMode os.FileMode
	Data     bytes.Buffer
}
