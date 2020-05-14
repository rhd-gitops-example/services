package git

import (
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
func CopyService(serviceName string, source Source, dest Destination, sourceEnvironment, destinationEnvironment string) ([]string, error) {
	// filePath defines the root folder for serviceName's config in the repository
	// the lookup is done for the source repository
	filePath := pathForServiceConfig(serviceName, sourceEnvironment)
	copied := []string{}
	err := source.Walk(filePath, func(prefix, name string) error {
		sourcePath := path.Join(prefix, name)
		destPath := pathForServiceConfig(name, destinationEnvironment)
		if pathValidForPromotion(serviceName, destPath, destinationEnvironment) {
			err := dest.CopyFile(sourcePath, destPath)
			if err == nil {
				copied = append(copied, destPath)
			}
			return err
		}
		return nil
	})

	return copied, err
}

//  For a given serviceName, only files in environments/envName/services/serviceName/base/config/* are valid for promotion
func pathValidForPromotion(serviceName, filePath, environmentName string) bool {
	filterPath := filepath.Join(pathForServiceConfig(serviceName, environmentName), "base", "config")
	validPath := strings.HasPrefix(filePath, filterPath)
	return validPath
}

// pathForServiceConfig defines where in a 'gitops' repository the config for a given service should live.
func pathForServiceConfig(serviceName, environmentName string) string {
	pathForConfig := filepath.Join(string(filepath.Separator), "environments", environmentName, "services", serviceName)
	return pathForConfig
}
