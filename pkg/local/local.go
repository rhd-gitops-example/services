package local

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"github.com/rhd-gitops-example/services/pkg/git"
)


type Local struct {
	LocalPath string
	Debug     bool
	Logger    func(fmt string, v ...interface{})
}


// CopyConfig takes the name of a service and a Source local service root path to be copied to a Destination.
//
// Only files under /path/to/local/repo/config/* are copied to the destination /services/[serviceName]/base/config/* 
//
// Returns the list of files that were copied, and possibly an error.
func CopyConfig(serviceName string, source git.Source, dest git.Destination) ([]string, error) {

	copied := []string{}
	err := source.Walk("", func(prefix, name string) error {
		sourcePath := path.Join(prefix, name)
		destPath := pathForDestServiceConfig(serviceName, name)
		err := dest.CopyFile(sourcePath, destPath)
		if err == nil {
			copied = append(copied, destPath)
		}
		return err
	})
	return copied, err
}

// pathForDestServiceConfig defines where in a 'gitops' repository the local config file
// for a given service should live.
func pathForDestServiceConfig(serviceName, name string) string {
	return filepath.Join("services/", serviceName, "base", name)
}

func (l *Local) Walk(base string, cb func(prefix, name string) error) error {
	base = filepath.Join(l.LocalPath, "config")
	prefix := filepath.Dir(base) + "/"
	return filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		return cb(prefix, strings.TrimPrefix(path, prefix))
	})
}
