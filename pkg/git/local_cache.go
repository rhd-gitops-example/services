package git

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

type userCredentials struct {
	name  string
	email string
}

type LocalCache struct {
	cacheDir     string
	credentials  userCredentials
	cacheFactory cacheFactory
}

type cacheFactory func(cacheDir, repoURL string) (*git.Repository, string, error)

// NewLocalCache creates a new LocalCache and ensures that the provided cacheDir
// exists.
func NewLocalCache(cacheDir, name, email string) (*LocalCache, error) {
	cacheDirExists, err := dirExists(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("error getting the cache dir: %w", err)
	}
	if !cacheDirExists {
		// TODO: best-practices for the perms?
		err = os.MkdirAll(cacheDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("error getting the cache dir: %w", err)
		}
	}
	return &LocalCache{cacheDir: cacheDir, credentials: userCredentials{name: name, email: email}, cacheFactory: openCacheRepository}, nil
}

func (l *LocalCache) ReadFileFromBranch(ctx context.Context, repoURL, filePath, branch string) ([]byte, error) {
	r, repoCacheDir, err := l.cacheFactory(l.cacheDir, repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache repository %s, %s: %w", repoURL, branch, err)
	}

	err = checkoutBranch(r, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to checkout branch %s, %s: %w", repoURL, branch, err)
	}

	filename := path.Join(repoCacheDir, filePath)
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s, %s: %w", repoCacheDir, filePath, err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s, %s: %w", repoCacheDir, filePath, err)
	}
	return data, nil
}

func (l *LocalCache) WriteFileToBranchAndStage(ctx context.Context, repoURL, branch, filePath string, data []byte) error {
	r, repoCacheDir, err := l.cacheFactory(l.cacheDir, repoURL)
	if err != nil {
		return fmt.Errorf("failed to open cache repository %s, %s: %w", repoURL, branch, err)
	}

	err = checkoutBranch(r, branch)
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s, %s: %w", repoURL, branch, err)
	}

	filename := path.Join(repoCacheDir, filePath)
	err = ioutil.WriteFile(filename, data, 0755)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to create a WorkTree: %w", err)
	}

	_, err = w.Add(filePath)
	if err != nil {
		return fmt.Errorf("failed to add files to WorkTree: %w", err)
	}

	return nil
}

func (l *LocalCache) CreateAndCheckoutBranch(ctx context.Context, repoURL, fromBranch, newBranch string) error {
	r, _, err := l.cacheFactory(l.cacheDir, repoURL)
	if err != nil {
		return fmt.Errorf("failed to open cache repository %s, %s: %w", repoURL, fromBranch, err)
	}

	err = checkoutBranch(r, fromBranch)
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s, %s: %w", repoURL, fromBranch, err)
	}

	headRef, err := r.Head()
	if err != nil {
		return fmt.Errorf("failed to get the head ref %s, %s: %w", repoURL, fromBranch, err)
	}
	newRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(newBranch), headRef.Hash())
	err = r.Storer.SetReference(newRef)
	if err != nil {
		return fmt.Errorf("failed to set new reference %s, %v: %w", repoURL, newRef, err)
	}
	err = checkoutBranch(r, newBranch)
	if err != nil {
		return fmt.Errorf("failed to checkout new branch %s, %v: %w", repoURL, newBranch, err)
	}
	return nil
}

func (l *LocalCache) CommitAndPushBranch(ctx context.Context, repoURL, branch, message, token string) error {
	r, _, err := l.cacheFactory(l.cacheDir, repoURL)
	if err != nil {
		return fmt.Errorf("failed to open cache repository %s, %s: %w", repoURL, branch, err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to create a WorkTree: %w", err)
	}
	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  l.credentials.name,
			Email: l.credentials.email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	err = r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			// This isn't used by GitHub.
			Username: "promotion",
			Password: token,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	return nil
}

func openCacheRepository(cacheDir, repoURL string) (*git.Repository, string, error) {
	repoCacheDir := path.Join(cacheDir, encodeURL(repoURL))
	hasCache, err := dirExists(repoCacheDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check for cached repo %v: %w", repoURL, err)
	}
	var r *git.Repository
	if !hasCache {
		r, err = git.PlainClone(repoCacheDir, false, &git.CloneOptions{
			URL: repoURL,
		})
		if err != nil && err != git.ErrRepositoryAlreadyExists {
			return nil, "", fmt.Errorf("failed to clone repo %v: %w", repoURL, err)
		}
	} else {
		r, err = git.PlainOpen(repoCacheDir)
		if err != nil {
			return nil, "", fmt.Errorf("failed to open an existing repo for %v: %w", repoURL, err)
		}
	}
	return r, repoCacheDir, nil
}

func checkoutBranch(r *git.Repository, branch string) error {
	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to create a WorkTree: %w", err)
	}

	err = w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(branch)})
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branch, err)
	}
	return nil
}
