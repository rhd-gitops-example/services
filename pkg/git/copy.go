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
// Note this currently relies on the environment folder being in *both* the source and destination repos.
// sourceEnvironment is
// --env = destination environment
// todo: determine source before we call copy service, then make sure we use env as a true dest environment only
func CopyService(serviceName string, source Source, dest Destination, sourceEnvironment, destinationEnvironment string) ([]string, error) {
	fmt.Printf("in copy\n")
	// filePath defines the root folder for serviceName's config in the repository
	// the lookup is done for the source repository
	filePath := pathForServiceConfig(serviceName, sourceEnvironment)
	fmt.Printf("about to walk and filepath is: %s\n", filePath)
	copied := []string{}
	err := source.Walk(filePath, func(prefix, name string) error {
		sourcePath := path.Join(prefix, name)
		fmt.Printf("Source path: %s\n", sourcePath)
		destPath := pathForServiceConfig(name, destinationEnvironment)
		fmt.Printf("Dest path: %s\n", destPath)
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

// pathValidForPromotion()
//  For a given serviceName, only files in services/serviceName/base/config/* are valid for promotion
//
func pathValidForPromotion(serviceName, filePath, environmentName string) bool {
	filterPath := filepath.Join(pathForServiceConfig(serviceName, environmentName), "base", "config")
	validPath := strings.HasPrefix(filePath, filterPath)
	fmt.Printf("Valid for promotion: %t\n", validPath)
	return validPath
}

// pathForServiceConfig defines where in a 'gitops' repository the config for a given service should live.
func pathForServiceConfig(serviceName, environmentName string) string {
	pathForConfig := filepath.Join("environments", environmentName, "services", serviceName)
	fmt.Printf("pathForConfig: %s\n", pathForConfig)
	return pathForConfig
}
