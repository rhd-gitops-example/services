package git

import (
	"net/url"
	"os"
)

func encodeURL(s string) string {
	return url.QueryEscape(s)
}

// dirExists returns true if the provided path exists.
// TODO: this should probably check to see if the thing that's statted is a
// directory.
func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}
		return false, nil
	}
	return true, nil
}
