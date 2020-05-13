package avancement

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/rhd-gitops-example/services/pkg/git"
	"github.com/rhd-gitops-example/services/pkg/local"

	"github.com/google/uuid"
)

type ServiceManager struct {
	cacheDir      string
	author        *git.Author
	clientFactory scmClientFactory
	repoFactory   repoFactory
	localFactory  localFactory
	tlsVerify     bool
	repoType      string
	debug         bool
}

type scmClientFactory func(token, toURL, repoType string, tlsVerify bool) *scm.Client
type repoFactory func(url, localPath string, tlsVerify, debug bool) (git.Repo, error)
type localFactory func(localPath string, debug bool) git.Source
type serviceOpt func(*ServiceManager)

// New creates and returns a new ServiceManager.
//
// The cacheDir used to checkout the source and destination repos.
// The token is used to create a Git token
func New(cacheDir string, author *git.Author, opts ...serviceOpt) *ServiceManager {
	sm := &ServiceManager{
		cacheDir:      cacheDir,
		author:        author,
		clientFactory: git.CreateClient,
		repoFactory: func(url, localPath string, tlsVerify, debug bool) (git.Repo, error) {
			r, err := git.NewRepository(url, localPath, tlsVerify, debug)
			return git.Repo(r), err
		},
		localFactory: func(localPath string, debug bool) git.Source {
			l := &local.Local{LocalPath: localPath, Debug: debug, Logger: log.Printf}
			return git.Source(l)
		},
	}
	for _, o := range opts {
		o(sm)
	}
	return sm
}

// WithDebug is a service option that configures the ServiceManager for
// additional debugging.
func WithDebug(f bool) serviceOpt {
	return func(sm *ServiceManager) {
		sm.debug = f
	}
}

// WithTlsVerify is a service option that configures the ServiceManager for
// TLS verify option.
func WithInsecureSkipVerify(f bool) serviceOpt {
	return func(sm *ServiceManager) {
		sm.tlsVerify = !f
	}
}

// WithRepType is a service option that configures the ServiceManager for
// repository type option.  Supported values are "github", "gitlab, and "ghe"
func WithRepoType(f string) serviceOpt {
	return func(sm *ServiceManager) {
		sm.repoType = strings.ToLower(f)
	}
}

// TODO: make this a command-line parameter defaulting to "master"
const fromBranch = "master"

// Promote is the main driver for promoting files between two
// repositories.
//
// It uses a Git cache to checkout the code to, and will copy the environment
// configuration for the `fromURL` to the `toURL` in a named branch.
func (s *ServiceManager) Promote(serviceName, fromURL, toURL, newBranchName, message string, keepCache bool) error {
	var source, destination git.Repo
	var reposToDelete []git.Repo

	defer func(keepRepos bool, repos *[]git.Repo) {
		if !keepRepos {
			for _, repo := range *repos {
				err := repo.DeleteCache()
				if err != nil {
					log.Printf("failed deleting files from cache: %s", err)
				}
			}
		}
	}(keepCache, &reposToDelete)

	var localSource git.Source
	var errorSource error
	isLocal := fromLocalRepo(fromURL)

	if isLocal {
		localSource = s.localFactory(fromURL, s.debug)
		if newBranchName == "" {
			newBranchName = generateBranchForLocalSource(localSource)
		}
	} else {
		source, errorSource = s.checkoutSourceRepo(fromURL, fromBranch)
		if errorSource != nil {
			return git.GitError("error checking out source repository from Git", fromURL)
		}
		reposToDelete = append(reposToDelete, source)
		if newBranchName == "" {
			newBranchName = generateBranch(source)
		}
	}

	destination, err := s.checkoutDestinationRepo(toURL, newBranchName)
	if err != nil {
		if git.IsGitError(err) {
			return git.GitError(fmt.Sprintf("failed to clone destination repository, error: %s", err.Error()), toURL)
		}
		// This would be a checkout error as the clone error gives us the above gitError instead
		return fmt.Errorf("failed to checkout destination repository, error: %w", err)
	}
	reposToDelete = append(reposToDelete, destination)

	var copied []string

	if isLocal {
		if destination != nil {
			overrideTargetFolder, err := destination.GetUniqueEnvironmentFolder()
			if err != nil {
				return fmt.Errorf("could not determine unique environment name for destination repository, error: %s", err.Error())
			}
			if overrideTargetFolder == "" {
				return fmt.Errorf("could not determine destination environment name")
			}
			copied, err = local.CopyConfig(serviceName, localSource, destination, path.Join("environments", overrideTargetFolder))
			if err != nil {
				return fmt.Errorf("failed to set up local repository: %w", err)
			}
		}
	} else {
		sourceEnvironment := ""
		destinationEnvironment := ""
		sourceEnvironment, err := source.GetUniqueEnvironmentFolder()
		if err != nil {
			return fmt.Errorf("could not determine unique environment name for source repository, error: %s", err.Error())
		}

		// Shouldn't hit this, but if so... we don't wanna continue as that'd go panic
		if sourceEnvironment == "" {
			return fmt.Errorf("could not determine source environment name")
		}

		destinationEnvironment, err = destination.GetUniqueEnvironmentFolder()

		if destinationEnvironment == "" {
			return fmt.Errorf("could not determine destination environment name")
		}

		if err != nil {
			return fmt.Errorf("could not determine unique environment name for destination repository, error: %s", err.Error())
		}

		copied, err = git.CopyService(serviceName, source, destination, sourceEnvironment, path.Join("environments", destinationEnvironment))
		if err != nil {
			return fmt.Errorf("failed to copy service: %w", err)
		}
	}

	commitMsg := message
	if commitMsg == "" {
		if isLocal {
			commitMsg = fmt.Sprintf("Promotion of service %s from local filesystem directory %s.", serviceName, fromURL)
		} else {
			commitMsg = generateDefaultCommitMsg(source, serviceName, fromURL, fromBranch)
		}
	}
	if err := destination.StageFiles(copied...); err != nil {
		return fmt.Errorf("failed to stage files %s: %w", copied, err)
	}
	if err := destination.Commit(commitMsg, s.author); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	if err := destination.Push(newBranchName); err != nil {
		return fmt.Errorf("failed to push to Git repository - check the access token is correct with sufficient permissions: %w", err)
	}

	ctx := context.Background()
	pr, err := createPullRequest(ctx, fromURL, toURL, newBranchName, commitMsg, s.clientFactory(s.author.Token, toURL, s.repoType, s.tlsVerify), isLocal)
	if err != nil {
		message := fmt.Sprintf("failed to create a pull-request for branch %s, error: %s", newBranchName, err)
		return git.GitError(message, toURL)
	}
	log.Printf("created PR %d", pr.Number)
	return nil
}

func (s *ServiceManager) checkoutSourceRepo(repoURL, branch string) (git.Repo, error) {
	repo, err := s.cloneRepo(repoURL, branch)
	if err != nil {
		if git.IsGitError(err) {
			return nil, git.GitError("failed to clone source repository", repoURL)
		}
		return nil, err
	}
	err = repo.Checkout(branch)
	if err != nil {
		return nil, git.GitError(fmt.Sprintf("failed to checkout branch %s, error: %s", branch, err.Error()), repoURL)
	}
	return repo, nil
}

func (s *ServiceManager) checkoutDestinationRepo(repoURL, branch string) (git.Repo, error) {
	repo, err := s.cloneRepo(repoURL, branch)
	if err != nil {
		return nil, git.GitError(err.Error(), repoURL)
	}
	err = repo.CheckoutAndCreate(branch)
	if err != nil {
		return nil, fmt.Errorf("failed to checkout branch %s, error: %w", branch, err)
	}
	return repo, nil
}

func (s *ServiceManager) cloneRepo(repoURL, branch string) (git.Repo, error) {
	// This ensures that the URL has credentials for the author.
	repoURL, err := addCredentialsIfNecessary(repoURL, s.author)
	if err != nil {
		return nil, err
	}
	repo, err := s.repoFactory(repoURL, path.Join(s.cacheDir, encode(repoURL, branch)), s.tlsVerify, s.debug)
	if err != nil {
		message := fmt.Sprintf("failed to clone repository, error is: %s", err.Error())
		return nil, git.GitError(message, repoURL)
	}
	err = repo.Clone()
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func createPullRequest(ctx context.Context, fromURL, toURL, newBranchName, commitMsg string, client *scm.Client, fromLocal bool) (*scm.PullRequest, error) {
	prInput, err := makePullRequestInput(fromLocal, fromURL, toURL, newBranchName, commitMsg)
	if err != nil {
		// TODO: improve this error message
		return nil, err
	}

	u, _ := url.Parse(toURL)
	// take out ".git" at the end
	pr, _, err := client.PullRequests.Create(ctx, u.Path[1:len(u.Path)-4], prInput)
	return pr, err
}

func encode(gitURL, branch string) string {
	return url.QueryEscape(gitURL) + "-" + url.QueryEscape(branch)
}

func addCredentialsIfNecessary(s string, a *git.Author) (string, error) {
	parsed, err := url.Parse(s)
	// If this does error (e.g. if there's a leading ":" character") one would see
	// parse ":thetoken@github.com": missing protocol scheme
	// thus surfacing the token. So intentionally display a generic parse problem
	if err != nil {
		return "", errors.New("failed to parse git repository url")
	}
	if parsed.Scheme != "https" || parsed.User != nil {
		return s, nil
	}
	parsed.User = url.UserPassword("promotion", a.Token)
	return parsed.String(), nil
}

func fromLocalRepo(s string) bool {
	parsed, err := url.Parse(s)
	if err != nil || parsed.Scheme == "" {
		return true
	}
	return false
}

func generateBranch(repo git.Repo) string {
	uniqueString := uuid.New()
	runes := []rune(uniqueString.String())
	branchName := repo.GetName() + "-" + repo.GetCommitID() + "-" + string(runes[0:5])
	branchName = strings.Replace(branchName, "\n", "", -1)
	return branchName
}

func generateBranchForLocalSource(source git.Source) string {
	uniqueString := uuid.New()
	runes := []rune(uniqueString.String())
	branchName := source.GetName() + "-" + "local-dir" + "-" + string(runes[0:5])
	branchName = strings.Replace(branchName, "\n", "", -1)
	return branchName
}

// generateDefaultCommitMsg constructs a default commit message based on the source information.
func generateDefaultCommitMsg(sourceRepo git.Repo, serviceName, from, fromBranch string) string {
	commit := sourceRepo.GetCommitID()
	msg := fmt.Sprintf("Promoting service %s at commit %s from branch %s in %s.", serviceName, commit, fromBranch, from)
	return msg
}
