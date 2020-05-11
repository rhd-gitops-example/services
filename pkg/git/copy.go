package git

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// CopyService takes the name of a service to copy from a Source to a Destination.
// Source, Destination implement Walk() and are typically Repository objects.
//
// Only files under /services/[serviceName]/base/config/* are copied to the destination
//
// Returns the list of files that were copied, and possibly an error.
func CopyService(serviceName string, source Source, dest Destination, overrideTargetFolder string) ([]string, error) {

	// filePath defines the root folder for serviceName's config in the repository
	filePath := pathForServiceConfig(serviceName)

	copied := []string{}
	err := source.Walk(filePath, func(prefix, name string) error {
		sourcePath := path.Join(prefix, name)
		destPath := pathForServiceConfig(name)
		if pathValidForPromotion(serviceName, destPath) {
			newPath := "/" + overrideTargetFolder + "/" + destPath
			err := dest.CopyFile(sourcePath, newPath)
			if err == nil {
				fmt.Printf("copied %s to %s", sourcePath, newPath)
				copied = append(copied, newPath)
			}
			return err
		}
		return nil
	})

	return copied, err
}

// pathValidForPromotion()
//  For a given serviceName, only files in services/serviceName/base/config/* are valid for promotion
//
func pathValidForPromotion(serviceName, filePath string) bool {
	filterPath := filepath.Join(pathForServiceConfig(serviceName), "base", "config")
	validPath := strings.HasPrefix(filePath, filterPath)
	return validPath
}

// pathForServiceConfig defines where in a 'gitops' repository the config for a given service should live.
func pathForServiceConfig(serviceName string) string {
	return filepath.Join("services", serviceName)
}
