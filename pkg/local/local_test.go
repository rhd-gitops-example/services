package local

import (
	"errors"
	"io"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCopyConfig(t *testing.T) {
	s := &mockSource{localPath: "/tmp/testing"}
	files := []string{"config/my-file.yaml", "config/this-file.yaml"}
	copiedfiles := []string{"services/service-a/base/config/my-file.yaml", "services/service-a/base/config/this-file.yaml"}
	for _, f := range files {
		s.addFile(f)
	}
	d := &mockDestination{}

	copied, err := CopyConfig("service-a", s, d, "")
	if err != nil {
		t.Fatal(err)
	}
	d.assertFilesWritten(t, copiedfiles)
	if !reflect.DeepEqual(copiedfiles, copied) {
		t.Fatalf("failed to copy the files, got %#v, want %#v", copied, files)
	}
}

func TestCopyConfigCanUseTargetDirectoryOverride(t *testing.T) {
	s := &mockSource{localPath: "/tmp/testing"}
	files := []string{"config/my-file.yaml", "config/this-file.yaml"}
	copiedfiles := []string{"environments/staging/services/service-a/base/config/my-file.yaml", "environments/staging/services/service-a/base/config/this-file.yaml"}
	for _, f := range files {
		s.addFile(f)
	}
	d := &mockDestination{}

	copied, err := CopyConfig("service-a", s, d, "environments/staging")
	if err != nil {
		t.Fatal(err)
	}
	d.assertFilesWritten(t, copiedfiles)
	if !reflect.DeepEqual(copiedfiles, copied) {
		t.Fatalf("failed to copy the files, got %#v, want %#v", copied, files)
	}
}

type mockSource struct {
	files     []string
	localPath string
}

// Walk: a mock function to emulate what happens in Repository.Walk()
// The Mock version is different: it iterates over mockSource.files[] and then drives
// the visitor callback in CopyService() as usual.
//
// To preserve the same behaviour, we see that Repository Walk receives /full/path/to/repo/services/service-name
// and then calls filePath.Walk() on /full/path/to/repo/services/ .
// When CopyService() drives Walk(), 'base' is typically services/service-name
// Thus we take each /full/path/to/file/in/mockSource.files[] and split it at 'services/' as happens in the Walk() method we're mocking.
func (s *mockSource) Walk(_ string, cb func(string, string) error) error {
	if s.files == nil {
		return nil
	}
	base := filepath.Join(s.localPath, "config")

	for _, f := range s.files {
		splitString := filepath.Dir(base) + "/"
		splitPoint := strings.Index(f, splitString) + len(splitString)
		prefix := f[:splitPoint]
		name := f[splitPoint:]
		err := cb(prefix, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *mockSource) GetName() string {
	return "/workspace/path/to/dir"
}

func (s *mockSource) addFile(name string) {
	if s.files == nil {
		s.files = []string{}
	}
	s.files = append(s.files, path.Join(s.localPath, name))
}

type mockDestination struct {
	written   []string
	copyError error
}

func (d *mockDestination) CopyFile(src, dst string) error {
	if d.written == nil {
		d.written = []string{}
	}
	if d.copyError != nil {
		return d.copyError
	}
	d.written = append(d.written, dst)
	return nil
}

// Returns true only if filepath is environments or environments/staging
// Same as FileExists, forgive me
func (d *mockDestination) DestFileExists(filepath string) bool {
	if filepath == "environments" || filepath == "environments/staging" {
		return true
	}
	return false
}

// Returns true only if filepath is environments or environments/staging
func (d *mockDestination) FileExists(filepath string) bool {
	if filepath == "environments" || filepath == "environments/staging" {
		return true
	}
	return false
}

func (d *mockDestination) WriteFile(src io.Reader, dst string) error {
	return errors.New("not implemented just now")
}

func (d *mockDestination) assertFilesWritten(t *testing.T, want []string) {
	if diff := cmp.Diff(want, d.written); diff != "" {
		t.Fatalf("written files do not match: %s", diff)
	}
}
