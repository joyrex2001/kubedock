package tar

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"

	"k8s.io/klog"
)

// PackFolder will write the given folder as a tar to the given Writer.
func PackFolder(src string, buf io.Writer) error {
	tw := tar.NewWriter(buf)

	// walk through every file in the folder
	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, file)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(rel)
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		klog.V(4).Infof("add to tar file: %s", header.Name)

		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}

	return nil
}

// UnpackFile will extract the given file from the given archive to the
// given dest writer.
func UnpackFile(dst, fname string, archive io.Reader, dest io.Writer) error {
	tr, err := NewReader(archive)
	if err != nil {
		return err
	}
	for {
		header, err := tr.Next()
		if err != nil {
			return err
		}
		if header != nil && filepath.Join(dst, header.Name) == fname {
			_, err = io.Copy(dest, tr)
			return err
		}
	}
}

// GetTargetFolderNames will return all affected folders in the archive
// provided.
func GetTargetFolderNames(dst string, archive io.Reader) ([]string, error) {
	return getTargets(dst, archive, tar.TypeDir)
}

// GetTargetFileNames will return all file names in the archive
// provided.
func GetTargetFileNames(dst string, archive io.Reader) ([]string, error) {
	return getTargets(dst, archive, tar.TypeReg)
}

// getTargets will return all given asset names of type (dir/file).
func getTargets(dst string, archive io.Reader, typ byte) ([]string, error) {
	res := []string{}
	tr := tar.NewReader(archive)
	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return res, nil
		case err != nil:
			return res, err
		case header == nil:
			continue
		}
		target := filepath.Join(dst, header.Name)
		if header.Typeflag == typ {
			res = append(res, target)
		}
	}
}

// IsSingleFileArchive will return true if there is only 1 file stored in the
// given archive.
func IsSingleFileArchive(archive []byte) bool {
	tr, err := NewReader(bytes.NewReader(archive))
	if err != nil {
		klog.Errorf("error reading tar archive: %v", err)
		return false
	}
	count := 0
	for count < 2 {
		header, err := tr.Next()
		if err != nil {
			return count == 1
		}
		if header.Typeflag == tar.TypeReg {
			count++
		}
	}
	return count == 1
}

// GetTarSize will return the actual size of the tar file for a byte array
// containing padded tar data.
func GetTarSize(dat []byte) (int, error) {
	var err error

	tr, err := NewReader(bytes.NewReader(dat))
	if err != nil {
		return 0, err
	}

	for {
		if _, err = tr.Next(); err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		io.Copy(io.Discard, tr)
	}

	return tr.ReadBytes(), err
}
