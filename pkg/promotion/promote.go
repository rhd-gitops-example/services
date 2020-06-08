package promotion

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/rhd-gitops-example/services/pkg/git"
	"github.com/rhd-gitops-example/services/pkg/local"

	"github.com/google/uuid"
	"github.com/jenkins-x/go-scm/scm"
)

type EnvLocale struct {
	Path   string // URL or local path
	Branch string
}

func (env EnvLocale) IsLocal() bool {
	parsed, err := url.Parse(env.Path)
	if err != nil {
		log.Printf("Error while parsing URL for environment path: '%v', assuming it is local.\n", err)
		return true
	}
	return parsed.Scheme == ""
}

// Promote is the main driver for promoting files between two
// repositories.
//
// It uses a Git cache to checkout the code to, and will copy the environment
// configuration for the `fromURL` to the `toURL` in a named branch.
func (s *ServiceManager) Promote(serviceName string, from, to EnvLocale, newBranchName, message string, keepCache bool) error {
	var reposToDelete []git.Repo
	if !keepCache {
		defer clearCache(&reposToDelete)
	}

	var source git.Source
	var err error
	if from.IsLocal() {
		source = s.localFactory(from.Path, s.debug)
	} else {
		source, err = s.checkoutSourceRepo(from.Path, from.Branch)
		if err != nil {
			return git.GitError("error checking out source repository from Git", from.Path)
		}
		reposToDelete = append(reposToDelete, source.(git.Repo))
	}
	if newBranchName == "" {
		newBranchName = generateBranchName(source)
	}

	destination, err := s.checkoutDestinationRepo(to.Path, newBranchName)
	if err != nil {
		return err
	}
	reposToDelete = append(reposToDelete, destination)
	destinationEnvironment, err := destination.GetUniqueEnvironmentFolder()
	if err != nil {
		return fmt.Errorf("could not determine unique environment name for destination repository - check that only one directory exists under it and you can write to your cache folder")
	}

	var copied []string
	if from.IsLocal() {
		copied, err = local.CopyConfig(serviceName, source, destination, destinationEnvironment)
		if err != nil {
			return fmt.Errorf("failed to set up local repository: %w", err)
		}
	} else {
		repo, ok := source.(git.Repo)
		if !ok {
			// should not happen, but just in case
			return fmt.Errorf("failed to convert source '%v' to Git Repo", source)
		}
		sourceEnvironment, err := repo.GetUniqueEnvironmentFolder()
		if err != nil {
			return fmt.Errorf("could not determine unique environment name for source repository - check that only one directory exists under it and you can write to your cache folder")
		}

		copied, err = git.CopyService(serviceName, source, destination, sourceEnvironment, destinationEnvironment)
		if err != nil {
			return fmt.Errorf("failed to copy service: %w", err)
		}
	}

	if message == "" {
		message = generateDefaultCommitMsg(source, serviceName, from)
	}
	if err := destination.StageFiles(copied...); err != nil {
		return fmt.Errorf("failed to stage files %s: %w", copied, err)
	}
	if err := destination.Commit(message, s.author); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	if err := destination.Push(newBranchName); err != nil {
		return fmt.Errorf("failed to push to Git repository - check the access token is correct with sufficient permissions: %w", err)
	}

	ctx := context.Background()
	pr, err := createPullRequest(ctx, from, to, newBranchName, message, s.clientFactory(s.author.Token, to.Path, s.repoType, s.tlsVerify))
	if err != nil {
		message := fmt.Sprintf("failed to create a pull-request for branch %s, error: %s", newBranchName, err)
		return git.GitError(message, to.Path)
	}
	log.Printf("created PR %d", pr.Number)
	return nil
}

func clearCache(repos *[]git.Repo) {
	for _, repo := range *repos {
		err := repo.DeleteCache()
		if err != nil {
			log.Printf("failed deleting files from cache: %s", err)
		}
	}
}

// generateBranchName constructs a branch name based on the source (a commit or local) and
// a random UUID
func generateBranchName(source git.Source) string {
	var commitID string
	repo, ok := source.(git.Repo)
	if ok {
		commitID = repo.GetCommitID()
	} else {
		commitID = "local-dir"
	}

	uniqueString := uuid.New()
	runes := []rune(uniqueString.String())
	branchName := source.GetName() + "-" + commitID + "-" + string(runes[0:5])
	return strings.Replace(branchName, "\n", "", -1)
}

// generateDefaultCommitMsg constructs a default commit message based on the source information.
func generateDefaultCommitMsg(source git.Source, serviceName string, from EnvLocale) string {
	repo, ok := source.(git.Repo)
	if ok {
		commit := repo.GetCommitID()
		return fmt.Sprintf("Promoting service %s at commit %s from branch %s in %s.", serviceName, commit, from.Branch, from.Path)
	} else {
		return fmt.Sprintf("Promoting service %s from local filesystem directory %s.", serviceName, from.Path)
	}
}

func createPullRequest(ctx context.Context, from, to EnvLocale, newBranchName, commitMsg string, client *scm.Client) (*scm.PullRequest, error) {
	prInput, err := makePullRequestInput(from, to, newBranchName, commitMsg)
	if err != nil {
		return nil, err
	}

	u, _ := url.Parse(to.Path)
	pathToUse := strings.TrimPrefix(strings.TrimSuffix(u.Path, ".git"), "/")
	pr, _, err := client.PullRequests.Create(ctx, pathToUse, prInput)
	return pr, err
}
