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

// Promote is the main driver for promoting files between two
// repositories.
//
// It uses a Git cache to checkout the code to, and will copy the environment
// configuration for the `fromURL` to the `toURL` in a named branch.
func (s *ServiceManager) Promote(serviceName string, from, to EnvLocation, newBranchName, message string, keepCache bool) error {
	var reposToDelete []git.Repo
	if !keepCache {
		defer clearCache(&reposToDelete)
	}

	fromIsLocal, err := from.IsLocal()
	if err != nil {
		return fmt.Errorf("failed to determine if repository is local: %w", err)
	}

	var source git.Source
	if fromIsLocal {
		source = s.localFactory(from.RepoPath, s.debug)
	} else {
		source, err = s.checkoutSourceRepo(from.RepoPath, from.Branch)
		if err != nil {
			return git.GitError("error checking out source repository from Git", from.RepoPath)
		}
		reposToDelete = append(reposToDelete, source.(git.Repo))
	}
	if newBranchName == "" {
		newBranchName = generateBranchName(source)
	}

	destination, err := s.checkoutDestinationRepo(to.RepoPath, to.Branch, newBranchName)
	if err != nil {
		return err
	}
	reposToDelete = append(reposToDelete, destination)
	destinationEnvironment, err := getEnvironmentFolder(destination, to.Folder)
	if err != nil {
		return err
	}

	var copied []string
	if fromIsLocal {
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
		sourceEnvironment, err := getEnvironmentFolder(repo, from.Folder)
		if err != nil {
			return err
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
	pr, err := createPullRequest(ctx, from, to, newBranchName, message, s.clientFactory(s.author.Token, to.RepoPath, s.repoType, s.tlsVerify))
	if err != nil {
		message := fmt.Sprintf("failed to create a pull-request for branch %s, error: %s", newBranchName, err)
		return git.GitError(message, to.RepoPath)
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
func generateDefaultCommitMsg(source git.Source, serviceName string, from EnvLocation) string {
	repo, ok := source.(git.Repo)
	if ok {
		commit := repo.GetCommitID()
		return fmt.Sprintf("Promote service %s at commit %s from %v", serviceName, commit, from)
	} else {
		return fmt.Sprintf("Promote service %s from local filesystem directory %s", serviceName, from.RepoPath)
	}
}

func createPullRequest(ctx context.Context, from, to EnvLocation, newBranchName, commitMsg string, client *scm.Client) (*scm.PullRequest, error) {
	prInput, err := makePullRequestInput(from, to, newBranchName, commitMsg)
	if err != nil {
		return nil, err
	}

	u, _ := url.Parse(to.RepoPath)
	pathToUse := strings.TrimPrefix(strings.TrimSuffix(u.Path, ".git"), "/")
	pr, _, err := client.PullRequests.Create(ctx, pathToUse, prInput)
	return pr, err
}

// getEnvironmentFolder returns the name of the folder to use, or an error.
// Will return an error if the specified folder is not present in the repository,
// or if there are multiple folders and one isn't specified in the args here.
func getEnvironmentFolder(r git.Repo, folder string) (string, error) {
	if folder == "" {
		dir, err := r.GetUniqueEnvironmentFolder()
		if err != nil {
			return "", fmt.Errorf("could not determine unique environment name for source repository - check that only one directory exists under it and you can write to your cache folder")
		}
		return dir, nil
	}

	dirs, err := r.DirectoriesUnderPath("environments")
	if err != nil {
		return "", err
	}
	for _, dir := range dirs {
		if dir.Name() == folder {
			return dir.Name(), nil
		}
	}
	return "", fmt.Errorf("did not find environment folder matching '%v', only found '%v'", folder, dirs)
}
