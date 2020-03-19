package git

import "path"

// CopyService takes a path within a repository, and a source and destination,
// and copies the files from within the source, to the destination.
//
// Returns the list of files that were copied, and possibly an error.
func CopyService(filePath string, source Source, dest Destination) ([]string, error) {
	copied := []string{}
	err := source.Walk(filePath, func(prefix, name string) error {
		err := dest.CopyFile(path.Join(prefix, name), name)
		copied = append(copied, name)
		return err
	})
	return copied, err
}
