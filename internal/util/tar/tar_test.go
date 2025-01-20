package tar

import (
	"bufio"
	"bytes"
	"slices"
	"testing"
)

func TestPackFolder(t *testing.T) {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	err := PackFolder("./", w)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	w.Flush()

	dat := b.Bytes()
	if IsSingleFileArchive(dat) {
		t.Error("archive contains more than 1 file, but IsSingleFileArchive says not")
	}

	files, err := GetTargetFileNames("", bytes.NewReader(dat))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if !slices.Contains(files, "tar_test.go") {
		t.Error("expected archive to contain tar_test.go")
	}

	folders, err := GetTargetFolderNames("", bytes.NewReader(dat))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if !slices.Contains(folders, ".") {
		t.Error("expected archive to contain .")
	}

	fileMode, err := GetFileMode("", "tar_test.go", bytes.NewReader(dat))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if fileMode == 0 {
		t.Error("expected archive to contain tar_test.go with valid file mode")
	}

	rsz := len(dat)
	csz, err := GetTarSize(append(dat, []byte{0, 0, 0, 0}...))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if csz != rsz {
		t.Errorf("GetTarSize returns %d instead of %d bytes", csz, rsz)
	}
}
