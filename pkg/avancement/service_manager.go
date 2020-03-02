package avancement

import (
	"context"
	"fmt"
	"log"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/bigkevmcd/services/pkg/git"
	"github.com/bigkevmcd/services/pkg/util"
)

type ServiceManager struct {
	cache         git.Cache
	token         string
	clientFactory scmClientFactory
}

type scmClientFactory func(token string) *scm.Client

// New creates and returns a new ServiceManager.
//
// The token is used to create a Git token
func New(c git.Cache, token string) *ServiceManager {
	return &ServiceManager{cache: c, token: token, clientFactory: git.CreateGitHubClient}
}

// Promote is the main driver for promoting files between two
// repositories.
//
// It uses a Git cache to checkout the code to, and will copy the environment
// configuration for the `fromURL` to the `toURL` in a named branch.
func (s *ServiceManager) Promote(service, fromURL, toURL, newBranchName string) error {
	ctx := context.Background()
	err := s.cache.CreateAndCheckoutBranch(ctx, toURL, "master", newBranchName)
	if err != nil {
		return fmt.Errorf("failed to create and checkout the new branch %v for the %v environment: %w", newBranchName, toURL, err)
	}

	fileToUpdate := pathForService(service)
	newBody, err := s.cache.ReadFileFromBranch(ctx, fromURL, fileToUpdate, "master")
	if err != nil {
		return fmt.Errorf("failed to read the file %v from the %v environment: %w", fileToUpdate, fromURL, err)
	}
	err = s.cache.WriteFileToBranchAndStage(ctx, toURL, newBranchName, fileToUpdate, newBody)
	if err != nil {
		return fmt.Errorf("failed to write the updated file to %v: %w", fileToUpdate, err)
	}

	err = s.cache.CommitAndPushBranch(ctx, toURL, newBranchName, "this is a test commit", s.token)
	if err != nil {
		return fmt.Errorf("failed to commit and push branch for environment %v: %w", toURL, err)
	}

	pr, err := createPullRequest(ctx, fromURL, toURL, newBranchName, s.clientFactory(s.token))
	if err != nil {
		return fmt.Errorf("failed to create a pull-request for branch %s in %v: %w", newBranchName, toURL, err)
	}
	log.Printf("created PR %d", pr.Number)
	return nil
}

func pathForService(s string) string {
	return fmt.Sprintf("%s/deployment.txt", s)
}

func createPullRequest(ctx context.Context, fromURL, toURL, newBranchName string, client *scm.Client) (*scm.PullRequest, error) {
	prInput, err := makePullRequestInput(fromURL, toURL, newBranchName)
	if err != nil {
		// TODO: improve this error message
		return nil, err
	}

	user, repo, err := util.ExtractUserAndRepo(toURL)
	if err != nil {
		// TODO: improve this error message
		return nil, err
	}
	// TODO: come up with a better way of generating the repo URL (this
	// only works for GitHub)
	pr, _, err := client.PullRequests.Create(ctx, fmt.Sprintf("%s/%s", user, repo), prInput)
	return pr, err
}
