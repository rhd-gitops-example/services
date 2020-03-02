package git

import (
	"context"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/github"
	"golang.org/x/oauth2"
)

// CreateGitHubClient creates and returns a go-scm GitHub client, using the provided
// oauth2 token.
func CreateGitHubClient(token string) *scm.Client {
	client := github.NewDefault()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client.Client = oauth2.NewClient(context.Background(), ts)
	return client
}
