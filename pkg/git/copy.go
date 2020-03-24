package git

import (
	"path"
	"strings"
)

// CopyService takes the name of a service to copy from a Source to a Destination.
// Source, Destination implement Walk() and are typically Repository objects.
//
// Only files under /services/[serviceName]/base/config/* are copied to the destination
//
// Returns the list of files that were copied, and possibly an error.
func CopyService(serviceName string, source Source, dest Destination) ([]string, error) {

	// filePath defines the root folder for serviceName's config in the repository
	filePath := path.Join("services", serviceName)

	copied := []string{}
	err := source.Walk(filePath, func(prefix, name string) error {
		sourcePath := path.Join(prefix, name)
		destPath := path.Join("services", name)

		// Only copy files in services/serviceName/base/config/*
		filterPath := path.Join("services", serviceName, "base/config")
		if strings.HasPrefix(destPath, filterPath) {
			err := dest.CopyFile(sourcePath, destPath)
			copied = append(copied, destPath)
			return err
		}
		return nil
	})
	return copied, err
}
