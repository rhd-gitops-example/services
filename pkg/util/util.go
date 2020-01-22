package util

import (
	"fmt"
	"net/url"
	"strings"
)

// ExtractUserAndRepo extracts the "user" or "organisation" from a GitHub
// repository URL, assuming that the first part of the path is the user and the
// second part of the path is the repository.
func ExtractUserAndRepo(repoURL string) (string, string, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", err
	}
	split := strings.Split(u.Path, "/")
	if len(split) < 3 {
		return "", "", fmt.Errorf("could not determine user and repo from URL: %v", u)
	}
	return split[1], strings.TrimSuffix(split[2], ".git"), nil
}
