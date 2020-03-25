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

func TestPathValidForPromotion(t *testing.T) {

	serviceBeingPromoted := "service-name"
	servicesNotBeingPromoted := []string{"svc", "base", "serviceName", "services"}
	promoteTheseWhenServiceNameIsRight := []string{
		"services/service-name/base/config/kustomization.yaml",
		"services/service-name/base/config/deployment.yaml",
		"services/service-name/base/config/dir/below/it/may/contain/important.yaml",
	}
	for _, filePath := range promoteTheseWhenServiceNameIsRight {
		if !pathValidForPromotion(serviceBeingPromoted, filePath) {
			t.Fatalf("Valid path for promotion for %s incorrectly rejected: %s", serviceBeingPromoted, filePath)
		}
		for _, wrongService := range servicesNotBeingPromoted {
			if pathValidForPromotion(wrongService, filePath) {
				t.Fatalf("Path for service %s incorrectly accepted for promotion: %s", wrongService, filePath)
			}
		}
	}

	// These paths should never be promoted
	badServiceNames := []string{"svc", "badService"}
	neverPromoteThese := []string{
		"services/svc/overlays/kustomization.yaml",
		"services/svc/kustomization.yaml",
		"svc/base/config/deployment.yaml",
		"services/badService/any/other/path/resource.yaml",
	}
	for _, badPath := range neverPromoteThese {
		for _, badServiceName := range badServiceNames {
			if pathValidForPromotion(badServiceName, badPath) {
				t.Fatalf("Invalid path %s for promotion of service %s incorrectly accepted", badPath, badServiceName)
			}
		}
	}
}

func TestPathForServiceConfig(t *testing.T) {
	serviceName := "usefulService"
	correctPath := "services/usefulService"
	serviceConfigPath := pathForServiceConfig(serviceName)
	if serviceConfigPath != correctPath {
		t.Fatalf("Invalid result for pathForServiceConfig(%s): wanted %s got %s", serviceName, correctPath, serviceConfigPath)
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
