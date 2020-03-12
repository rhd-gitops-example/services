package git

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// atomically copy files from source to destination, perserving permissions from
// the source file.
func fileCopy(src, dst string) error {
	perm, err := fileMode(src)
	if err != nil {
		return fmt.Errorf("failed to get permissions for existing file %s: %w", src, err)
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open file for copying: %w", err)
	}
	defer in.Close()
	tmp, err := ioutil.TempFile(filepath.Dir(dst), "")
	if err != nil {
		return fmt.Errorf("failed to create tempfile for copying: %w", err)
	}
	_, err = io.Copy(tmp, in)
	if err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return fmt.Errorf("failed to copy source file to destination: %w", err)
	}
	if err = tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("failed to close tempfile when copying: %w", err)
	}
	if err = os.Chmod(tmp.Name(), perm); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("failed to chmod tempfile when copying: %w", err)
	}
	if err = os.Rename(tmp.Name(), dst); err != nil {
		return fmt.Errorf("failed to rename tempfile when copying: %w", err)
	}
	return nil
}

func fileMode(filename string) (os.FileMode, error) {
	s, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return s.Mode(), nil
}
