package git

import (
	"errors"
	"fmt"
	"net/url"
)

// We can choose to throw one of these errors if we want to handle it
// in a particularly special way. This will be useful for Git API calls where
// we know there's going to be a repository URL involved and it's not
// some other error (e.g. to do with file permissions)
type gitError struct {
	msg string
}

// GitError returns an error whereby the user part of the url is removed (e.g. an access token)
// Accepts the message to use and the url to remove the user part from
func GitError(msg, url string) error {
	cleanedURL, err := CleanURL(url)
	if err != nil {
		return err
	}
	return gitError{msg: fmt.Sprintf("%s. repository URL = %s", msg, cleanedURL)}
}

// IsGitError performs a cast to see if an error really is of type gitError
func IsGitError(err error) bool {
	_, ok := err.(gitError)
	return ok
}

// CleanURL removes the .User part of the string (typically an access token)
func CleanURL(inputURL string) (string, error) {
	parsedURL, parseErr := url.Parse(inputURL)
	if parseErr != nil {
		// Intentionally don't display any real parse error as it would contain any token
		return "", errors.New("failed to parse git repository URL, check it's well-formed")
	}
	parsedURL.User = nil
	return parsedURL.String(), nil
}

func (e gitError) Error() string {
	return e.msg
}
