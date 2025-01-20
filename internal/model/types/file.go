package types

import (
	"bytes"
	"os"
)

type File struct {
	FileMode os.FileMode
	Data     bytes.Buffer
}
