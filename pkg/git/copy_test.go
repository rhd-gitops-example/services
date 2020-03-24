package git

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

func TestCopyService(t *testing.T) {
	s := &mockSource{localPath: "/tmp/testing"}
	files := []string{"services/service-a/base/config/my-file.yaml", "services/service-a/base/config/this-file.yaml"}
	for _, f := range files {
		s.addFile(f)
	}
	d := &mockDestination{}

	copied, err := CopyService("service-a", s, d)
	if err != nil {
		t.Fatal(err)
	}
	d.assertFilesWritten(t, files)
	if !reflect.DeepEqual(files, copied) {
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
func (s *mockSource) Walk(base string, cb func(string, string) error) error {
	if s.files == nil {
		return nil
	}

	for _, f := range s.files {
		if strings.HasPrefix(f, path.Join(s.localPath, base)) {
			splitString := filepath.Dir(base) + "/"
			splitPoint := strings.Index(f, splitString) + len(splitString)
			prefix := f[:splitPoint]
			name := f[splitPoint:]
			err := cb(prefix, name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *mockSource) addFile(name string) {
	if s.files == nil {
		s.files = []string{}
	}
	s.files = append(s.files, path.Join(s.localPath, name))
}

type mockDestination struct {
	written []string
}

func (d *mockDestination) CopyFile(src, dst string) error {
	if d.written == nil {
		d.written = []string{}
	}
	d.written = append(d.written, dst)
	return nil
}

func (d *mockDestination) WriteFile(src io.Reader, dst string) error {
	return errors.New("not implemented just now")
}

func (d *mockDestination) assertFilesWritten(t *testing.T, want []string) {
	if diff := cmp.Diff(want, d.written); diff != "" {
		t.Fatalf("written files do not match: %s", diff)
	}
}
