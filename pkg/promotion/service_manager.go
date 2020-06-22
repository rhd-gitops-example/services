package promotion

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/rhd-gitops-example/services/pkg/git"
	"github.com/rhd-gitops-example/services/pkg/local"
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

// checkoutDestinationRepo clones the specified repo to the cache, and creates a new
// branch (the "tip" branch) which forks off of the "base" branch.
func (s *ServiceManager) checkoutDestinationRepo(repoURL, baseBranch, tipBranch string) (git.Repo, error) {
	repo, err := s.cloneRepo(repoURL, tipBranch)
	if err != nil {
		return nil, git.GitError(fmt.Sprintf("failed to clone destination repository, error: %s", err.Error()), repoURL)
	}
	// need to checkout the base branch first, in case it is different than the default branch of the repository.
	err = repo.Checkout(baseBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to checkout existing branch %s, error: %w", baseBranch, err)
	}
	err = repo.CheckoutAndCreate(tipBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to create new branch %s, error: %w", tipBranch, err)
	}
	if repo == nil {
		// Should never happen, but if it does...
		return nil, fmt.Errorf("destination repository was not initialised despite being no errors")
	}
	return repo, nil
}

// cloneRepo clones the default branch of the repository into the cache.
// NOTE: the path the repo is cloned to includes the specified branch, but the
// actual contents are from the default branch.
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
