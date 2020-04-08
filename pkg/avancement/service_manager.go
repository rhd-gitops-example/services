package avancement

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/rhd-gitops-example/services/pkg/git"
	"github.com/rhd-gitops-example/services/pkg/local"
	"github.com/rhd-gitops-example/services/pkg/util"

	"github.com/google/uuid"
)

type ServiceManager struct {
	cacheDir      string
	author        *git.Author
	clientFactory scmClientFactory
	repoFactory   repoFactory
	localFactory  localFactory
	debug         bool
}

const (
	defaultCommitMsg = "this is a commit"
)

type scmClientFactory func(token string) *scm.Client
type repoFactory func(url, localPath string, debug bool) (git.Repo, error)
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
		clientFactory: git.CreateGitHubClient,
		repoFactory: func(url, localPath string, debug bool) (git.Repo, error) {
			r, err := git.NewRepository(url, localPath, debug)
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

// TODO: make this a command-line parameter defaulting to "master"
const fromBranch = "master"

// Promote is the main driver for promoting files between two
// repositories.
//
// It uses a Git cache to checkout the code to, and will copy the environment
// configuration for the `fromURL` to the `toURL` in a named branch.
func (s *ServiceManager) Promote(serviceName, fromURL, toURL, newBranchName string, keepCache bool) error {
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
	if fromLocalRepo(fromURL) {
		localSource = s.localFactory(fromURL, s.debug)
	} else {
		source, errorSource = s.checkoutSourceRepo(fromURL, fromBranch)
		if errorSource != nil {
			return fmt.Errorf("failed to checkout repo: %w", errorSource)
		}
		reposToDelete = append(reposToDelete, source)
	}

	if newBranchName == "" {
		newBranchName = generateBranch(source)
	}

	destination, err := s.checkoutDestinationRepo(toURL, newBranchName)
	if err != nil {
		return fmt.Errorf("failed to checkout repo: %w", err)
	}
	reposToDelete = append(reposToDelete, destination)

	var copied []string
	if fromLocalRepo(fromURL) {
		copied, err = local.CopyConfig(serviceName, localSource, destination)
		if err != nil {
			return fmt.Errorf("failed to setup local repo: %w", err)
		}
	} else {
		copied, err = git.CopyService(serviceName, source, destination)
		if err != nil {
			return fmt.Errorf("failed to copy service: %w", err)
		}
	}

	if err := destination.StageFiles(copied...); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}
	if err := destination.Commit(defaultCommitMsg, s.author); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	if err := destination.Push(newBranchName); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	ctx := context.Background()
	pr, err := createPullRequest(ctx, fromURL, toURL, newBranchName, s.clientFactory(s.author.Token))
	if err != nil {
		return fmt.Errorf("failed to create a pull-request for branch %s in %v: %w", newBranchName, toURL, err)
	}
	log.Printf("created PR %d", pr.Number)
	return nil
}

func (s *ServiceManager) checkoutSourceRepo(repoURL, branch string) (git.Repo, error) {
	repo, err := s.cloneRepo(repoURL, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to clone source repo %s: %w", repoURL, err)
	}
	err = repo.Checkout(branch)
	if err != nil {
		return nil, fmt.Errorf("failed to checkout branch %s in repo %s: %w", branch, repoURL, err)
	}
	return repo, nil
}

func (s *ServiceManager) checkoutDestinationRepo(repoURL, branch string) (git.Repo, error) {
	repo, err := s.cloneRepo(repoURL, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to clone destination repo %s: %w", repoURL, err)
	}
	err = repo.CheckoutAndCreate(branch)
	if err != nil {
		return nil, fmt.Errorf("failed to checkout branch %s in repo %s: %w", branch, repoURL, err)
	}
	return repo, nil
}

func (s *ServiceManager) cloneRepo(repoURL, branch string) (git.Repo, error) {
	// This ensures that the URL has credentials for the author.
	repoURL, err := addCredentialsIfNecessary(repoURL, s.author)
	if err != nil {
		return nil, err
	}
	repo, err := s.repoFactory(repoURL, path.Join(s.cacheDir, encode(repoURL, branch)), s.debug)
	if err != nil {
		return nil, fmt.Errorf("failed to create repo for URL %s: %w", repoURL, err)
	}
	err = repo.Clone()
	if err != nil {
		return nil, err
	}
	return repo, nil
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

func encode(gitURL, branch string) string {
	return url.QueryEscape(gitURL) + "-" + url.QueryEscape(branch)
}

func addCredentialsIfNecessary(s string, a *git.Author) (string, error) {
	parsed, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("failed to parse git repo url %v: %w", s, err)
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
	branchName = strings.Replace(branchName, "\n","",-1)
	return branchName
}
