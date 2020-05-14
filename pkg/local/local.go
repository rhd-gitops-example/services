package local

import (
	"os"
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
// Only files under /path/to/local/repo/config/* are copied to the destination /environments/envName//services/[serviceName]/base/config/*
// Accepts an override target folder as well, in the event we wish to place the code at a given directory
// This directory is prepended to the destination path
// Returns the list of files that were copied, and possibly an error.
func CopyConfig(serviceName string, source git.Source, dest git.Destination, environmentName string) ([]string, error) {
	copied := []string{}
	err := source.Walk("", func(prefix, name string) error {
		sourcePath := filepath.Join(prefix, name)
		destPath := pathForDestServiceConfig(serviceName, name, environmentName)
		err := dest.CopyFile(sourcePath, destPath)
		if err == nil {
			copied = append(copied, destPath)
		}
		return err
	})
	return copied, err
}

// pathForDestServiceConfig defines where in a 'gitops' repository the config
// for a given service should live.
func pathForDestServiceConfig(serviceName, name, environmentName string) string {
	return filepath.Join("environments", environmentName, "services", serviceName, "base", name)
}

func (l *Local) Walk(_ string, cb func(prefix, name string) error) error {
	base := filepath.Join(l.LocalPath, "config")
	prefix := filepath.Dir(base) + string(filepath.Separator)
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

// GetName - we're using a directory that may not be a git repo, all we know is our path
func (l *Local) GetName() string {
	path := filepath.ToSlash(l.LocalPath)
	path = strings.TrimLeft(path, string(filepath.Separator))
	name := strings.ReplaceAll(path, string(filepath.Separator), "-")
	return name
}
