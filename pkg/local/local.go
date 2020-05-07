package local

import (
	"fmt"
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
// Accepts an override target folder as well, in the event we wish to place the code at a given directory
// This directory is prepended to the destination path
// Returns the list of files that were copied, and possibly an error.
func CopyConfig(serviceName string, source git.Source, dest git.Destination, overrideTargetFolder string) ([]string, error) {
	copied := []string{}
	err := source.Walk("", func(prefix, name string) error {
		sourcePath := path.Join(prefix, name)
		destPath := pathForDestServiceConfig(serviceName, name)
		if overrideTargetFolder != "" {
			destPath = fmt.Sprintf("%s/%s", overrideTargetFolder, destPath)
		}
		if !dest.DestFileExists(overrideTargetFolder) {
			// Log output if the destination folder doesn't yet exist - maybe they've typod
			// E.g. they wanted to use --env prod but went with --prody. Instead of failing, say something
			fmt.Printf("Note: the overridden --env directory to promote into (%s) does not yet exist and will be created\n", overrideTargetFolder)
		}
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
func pathForDestServiceConfig(serviceName, name string) string {
	return filepath.Join("services/", serviceName, "base", name)
}

func (l *Local) Walk(_ string, cb func(prefix, name string) error) error {
	base := filepath.Join(l.LocalPath, "config")
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

// GetName - we're using a directory that may not be a git repo, all we know is our path
func (l *Local) GetName() string {
	path := filepath.ToSlash(l.LocalPath)
	path = strings.TrimLeft(path, "/")
	name := strings.ReplaceAll(path, "/", "-")
	return name
}
