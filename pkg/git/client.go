package git

import (
	"context"
	"net/http"
	"net/url"
	"crypto/tls"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/github"
	"github.com/jenkins-x/go-scm/scm/driver/gitlab"
	"golang.org/x/oauth2"
)

// CreateClient creates and returns a go-scm GitHub client, using the provided
// oauth2 token.
func CreateClient(token, toURL, repoType  string, tlsVerify bool) *scm.Client {
	var client *scm.Client
	u, _ := url.Parse(toURL)
	if repoType == "gitlab" {
		client = gitlab.NewDefault()
	} else if repoType == "ghe" {
		client, _ = github.New(u.Scheme + "://" + u.Host + "/api/v3")
	} else {
		client = github.NewDefault()
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	if tlsVerify {
		client.Client = oauth2.NewClient(context.Background(), ts)
	} else {
		client.Client = &http.Client{
			Transport: &oauth2.Transport{
				Source: ts,
				Base: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		}
	}
	if repoType == "gitlab" {
		u, _ := url.Parse(toURL)
		client.BaseURL.Host = u.Host
	}
	return client
}

