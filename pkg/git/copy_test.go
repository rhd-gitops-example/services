package git

import (
	"errors"
	"io"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCopyService(t *testing.T) {
	s := &mockSource{localPath: "/tmp/testing"}
	files := []string{"service-a/my-system/my-file.yaml", "service-a/my-system/this-file.yaml"}
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

func (s *mockSource) Walk(filePath string, cb func(string, string) error) error {
	if s.files == nil {
		return nil
	}
	for _, f := range s.files {
		if strings.HasPrefix(f, path.Join(s.localPath, filePath)) {
			err := cb(f, strings.TrimPrefix(f, s.localPath+"/"))
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
