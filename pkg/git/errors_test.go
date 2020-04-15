package git

import (
	"strings"
	"testing"
)

func TestCleanURLRemovesTokenFromError(t *testing.T) {
	url := "https://mytoken@github.com/my-repo"
	cleanedURL, err := CleanURL(url)
	if err != nil {
		t.Fatal("got an error cleaning a well-formed URL")
	}
	if strings.Contains(cleanedURL, "mytoken") {
		t.Fatal("error still contains access token")
	}
}

func TestCleanURLThrowsErrorIfInvalidURL(t *testing.T) {
	url := ":_ https://mytoken@github.com/my-repo"
	cleanedURL, err := CleanURL(url)
	if err == nil {
		t.Fatalf("did not get an an error cleaning a bad URL, cleaned URL is: %s", cleanedURL)
	}
}
