package mock

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rhd-gitops-example/services/pkg/git"
)

type Repository struct {
	currentBranch string
	knownBranches []string
	cloned        bool

	files     []string
	localPath string

	cloneErr    error
	checkoutErr error

	branchesCreated []string

	copiedFiles []string
	copyFileErr error

	commits   []string
	CommitErr error

	pushedBranches []string
	pushErr        error

	deleted   bool
	DeleteErr error

	commitID string

	repoName string
}

// New creates and returns a new git.Cache implementation that operates entirely
// in-memory.
func New(localPath string, branches ...string) *Repository {
	return &Repository{localPath: localPath, currentBranch: branches[0], knownBranches: branches}
}

// Checkout fulfils the git.Repo interface.
func (m *Repository) Checkout(branch string) error {
	if hasString(branch, m.knownBranches) {
		m.currentBranch = branch
	}
	return m.checkoutErr
}

func (m *Repository) GetName() string {
	m.repoName = "fakeRepoName"
	return m.repoName
}

func (m *Repository) GetCommitID() string {
	m.commitID = "a1b2c3d"
	return m.commitID
}

// CheckoutAndCreate fulfils the git.Repo interface.
func (m *Repository) CheckoutAndCreate(branch string) error {
	if m.branchesCreated == nil {
		m.branchesCreated = []string{}
	}
	m.branchesCreated = append(m.branchesCreated, key(m.currentBranch, branch))
	m.currentBranch = branch
	return m.checkoutErr
}

// Clone fulfils the git.Repo interface.
func (m *Repository) Clone() error {
	m.cloned = true
	return m.cloneErr
}

// StageFiles fulfils the git.Repo interface.
func (m *Repository) StageFiles(filenames ...string) error {
	return nil
}

// Commit fulfils the git.Repo interface.
func (m *Repository) Commit(msg string, author *git.Author) error {
	if m.commits == nil {
		m.commits = []string{}
	}
	m.commits = append(m.commits, key(m.currentBranch, msg, author.Token))
	return m.CommitErr
}

// Not implemented
func (m *Repository) DirectoriesUnderPath(path string) ([]os.FileInfo, error) {
	return nil, nil
}

func (m *Repository) GetUniqueEnvironmentFolder() (string, error) {
	// Cheap way to do it for mocking.
	// For environments/dev and environments/staging we want to return dev and staging.
	// The paths have / in them from the test code itself so / here is safe
	if len(m.files) != 1 {
		return "", fmt.Errorf("mock tried to get unique environment folder name but found no files for this repository, or several env folders: %s", m.repoName)
	}
	splits := strings.Split(m.files[0], "/")
	foundEnv := splits[1]
	return foundEnv, nil
}

// Push fulfils the git.Repo interface.
func (m *Repository) Push(branch string) error {
	if m.pushedBranches == nil {
		m.pushedBranches = []string{}
	}
	m.pushedBranches = append(m.pushedBranches, branch)
	return m.pushErr
}

// CopyFile fulfils the git.Repo interface.
func (m *Repository) CopyFile(src, dst string) error {
	if m.copiedFiles == nil {
		m.copiedFiles = []string{}
	}
	m.copiedFiles = append(m.copiedFiles, key(m.currentBranch, src, dst))
	return m.copyFileErr
}

// WriteFile fulfils the git.Repo interface.
func (m *Repository) WriteFile(src io.Reader, dst string) error {
	return nil
}

// Walk fulfils the git.Repo interface. It's a mock function to emulate what happens in Repository.Walk()
// The Mock version is different: it iterates over mockSource.files[] and then drives
// the visitor callback in CopyService() as usual.
//
// To preserve the same behaviour, we see that Repository Walk receives /full/path/to/repo/services/service-name
// and then calls filePath.Walk() on /full/path/to/repo/services/ .
// When CopyService() drives Walk(), 'base' is typically services/service-name
// Thus we take each /full/path/to/file/in/mockSource.files[] and split it at 'services/' as happens in the Walk() method we're mocking.
func (m *Repository) Walk(base string, cb func(string, string) error) error {
	if m.files == nil {
		return nil
	}
	for _, f := range m.files {
		// Basically pathToService
		if strings.HasPrefix(f, m.localPath) {
			splitString := filepath.Dir(base) + "/"
			splitPoint := strings.Index(f, splitString) + len(splitString)
			prefix := f[:splitPoint]
			name := f[splitPoint:]
			err := cb(prefix, name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AddFiles is part of the mock implementation, it records filenames so that
// they're used in the Walk implementation.
func (m *Repository) AddFiles(names ...string) {
	if m.files == nil {
		m.files = []string{}
	}
	for _, f := range names {
		m.files = append(m.files, path.Join(m.localPath, f))
	}
}

// DeleteCache deletes the repo from the local cache directory.
func (m *Repository) DeleteCache() error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	m.deleted = true
	return nil
}

// AssertBranchCreated asserts that the named branch was created from the from
// branch, using the `CheckoutAndCreate` implementation.
func (m *Repository) AssertBranchCreated(t *testing.T, from, name string) {
	if !hasString(key(from, name), m.branchesCreated) {
		t.Fatalf("branch %s was not created from %s", name, from)
	}
}

// AssertBranchNotCreated asserts that the named branch was *not* created from the from
// branch, using the `CheckoutAndCreate` implementation.
func (m *Repository) AssertBranchNotCreated(t *testing.T, from, name string) {
	if hasString(key(from, name), m.branchesCreated) {
		t.Fatalf("branch %s was created from %s", name, from)
	}
}

// AssertFileCopiedInBranch asserts the filename was copied from and to in a
// branch.
func (m *Repository) AssertFileCopiedInBranch(t *testing.T, branch, from, name string) {
	fmt.Printf("copied: %s\n", m.copiedFiles)
	if !hasString(key(branch, from, name), m.copiedFiles) {
		t.Fatalf("file %s was not copied from %s to branch %s", name, from, branch)
	}
}

// AssertFileNotCopiedInBranch asserts the filename was *not* copied from and to in a
// branch.
func (m *Repository) AssertFileNotCopiedInBranch(t *testing.T, branch, from, name string) {
	if hasString(key(branch, from, name), m.copiedFiles) {
		t.Fatalf("file %s was copied from %s to branch %s", name, from, branch)
	}
}

// AssertCommit asserts that a commit was created for the named branch with the
// message and auth token.
func (m *Repository) AssertCommit(t *testing.T, branch, msg string, a *git.Author) {
	if !hasString(key(branch, msg, a.Token), m.commits) {
		t.Fatalf("no matching commit %#v in branch %s using token %s.  Commits available: %+v", msg, branch, a.Token, m.commits)
	}
}

// AssertPush asserts that the branch was pushed.
func (m *Repository) AssertPush(t *testing.T, branch string) {
	if !hasString(branch, m.pushedBranches) {
		t.Fatalf("branch %s was not pushed", branch)
	}
}

// AssertDeletedFromCache asserts that delete was called to remove the local repo
func (m *Repository) AssertDeletedFromCache(t *testing.T) {
	if !m.deleted {
		t.Fatalf("repo was not deleted from the promotion cache directory")
	}
}

// AssertNotDeletedFromCache asserts that delete was called to remove the local repo
func (m *Repository) AssertNotDeletedFromCache(t *testing.T) {
	if m.deleted {
		t.Fatal("repo was unexpectedly deleted from the promotion cache directory")
	}
}

func key(v ...string) string {
	return strings.Join(v, ":")
}

func hasString(find string, list []string) bool {
	for _, v := range list {
		if find == v {
			return true
		}
	}
	return false
}
