package tar

import (
	"archive/tar"
	"bufio"
	"bytes"
	"slices"
	"strings"
	"testing"
)

// singleFileArchive creates an in-memory tar archive containing a single
// regular file with the given name and contents.
func singleFileArchive(t *testing.T, name, contents string) []byte {
	t.Helper()
	var b bytes.Buffer
	tw := tar.NewWriter(&b)
	hdr := &tar.Header{
		Name:     name,
		Mode:     0644,
		Size:     int64(len(contents)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("unexpected error writing header: %s", err)
	}
	if _, err := tw.Write([]byte(contents)); err != nil {
		t.Fatalf("unexpected error writing contents: %s", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("unexpected error closing archive: %s", err)
	}
	return b.Bytes()
}

// TestContainerPathsUseForwardSlash is a regression test for running kubedock
// on Windows. The dst is a container (linux) path and tar entry names are
// always forward-slash, so joining them must use POSIX separators regardless
// of the host OS. Previously path/filepath was used, which produced
// backslashes on Windows (e.g. \opt\keycloak\data\import\realm-export.json)
// and resulted in invalid container mount paths.
func TestContainerPathsUseForwardSlash(t *testing.T) {
	const dst = "/opt/keycloak/data/import"
	const file = "realm-export.json"
	const want = "/opt/keycloak/data/import/realm-export.json"

	dat := singleFileArchive(t, file, "{}")

	files, err := GetTargetFileNames(dst, bytes.NewReader(dat))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(files) != 1 || files[0] != want {
		t.Errorf("GetTargetFileNames returned %v, expected [%s]", files, want)
	}
	if strings.ContainsRune(files[0], '\\') {
		t.Errorf("target path %q contains a backslash; container paths must use forward slashes", files[0])
	}

	mode, err := GetFileMode(dst, want, bytes.NewReader(dat))
	if err != nil {
		t.Fatalf("unexpected error from GetFileMode: %s", err)
	}
	if mode == 0 {
		t.Error("expected a valid file mode")
	}

	var out bytes.Buffer
	if err := UnpackFile(dst, want, bytes.NewReader(dat), &out); err != nil {
		t.Fatalf("unexpected error from UnpackFile: %s", err)
	}
	if out.String() != "{}" {
		t.Errorf("UnpackFile returned %q, expected %q", out.String(), "{}")
	}
}

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

	mode, err := GetFileMode("", "tar_test.go", bytes.NewReader(dat))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if mode == 0 {
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

	if IsCompressed(dat[:5]) {
		t.Error("IsCompressed returns that archive is compressed, expected uncompressed")
	}
}
