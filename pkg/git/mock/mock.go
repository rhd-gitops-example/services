package mock

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

type MockCache struct {
	sourceFiles    map[string][]byte
	addRepoFileErr error

	writtenFiles map[string][]byte
	writeFileErr error

	branches             map[string][]string
	createAndCheckoutErr error

	commits   map[string][]commit
	commitErr error
}

type commit struct {
	msg    string
	branch string
}

// New creates and returns a new git.Cache implementation that operates entirely
// in-memory
func New() *MockCache {
	return &MockCache{
		sourceFiles:  make(map[string][]byte),
		writtenFiles: make(map[string][]byte),
		branches:     make(map[string][]string),
		commits:      make(map[string][]commit),
	}
}

// ReadFileFromBranch fulfils the git.Cache interface.
func (m *MockCache) ReadFileFromBranch(ctx context.Context, repoURL, filePath, branch string) ([]byte, error) {
	return m.sourceFiles[key(repoURL, branch, filePath)], m.addRepoFileErr
}

// CreateAndCheckoutBranch fulfils the git.Cache interface.
func (m *MockCache) CreateAndCheckoutBranch(ctx context.Context, repoURL, fromBranch, newBranch string) error {
	branches, ok := m.branches[repoURL]
	if !ok {
		branches = make([]string, 0)
	}
	branches = append(branches, key(fromBranch, newBranch))
	m.branches[repoURL] = branches
	return m.createAndCheckoutErr
}

// WriteFileToBranchAndStage fulfils the git.Cache interface.
func (m *MockCache) WriteFileToBranchAndStage(ctx context.Context, repoURL, branch, filePath string, data []byte) error {
	m.writtenFiles[key(repoURL, branch, filePath)] = data
	return m.writeFileErr
}

// CommitAndPushBranch fulfils the git.Cache interface.
func (m *MockCache) CommitAndPushBranch(ctx context.Context, repoURL, branch, message, token string) error {
	k := key(repoURL, branch)
	commits, ok := m.commits[k]
	if !ok {
		commits = make([]commit, 0)
	}
	commits = append(commits, commit{message, token})
	m.commits[k] = commits
	return m.commitErr
}

func (m *MockCache) AddRepoFile(repoURL, branch, path string, body []byte) {
	m.sourceFiles[key(repoURL, branch, path)] = body
}

func (m *MockCache) AssertBranchCreated(t *testing.T, repoURL, fromBranch, newBranch string) {
	t.Helper()
	branches, ok := m.branches[repoURL]
	if !ok {
		t.Fatalf("no branches created for %s", repoURL)
	}
	want := key(fromBranch, newBranch)
	for _, b := range branches {
		if b == want {
			return
		}
	}
	t.Fatalf("branch %s from %s not created in repo %s", newBranch, fromBranch, repoURL)
}

func (m *MockCache) AssertFileWrittenToBranch(t *testing.T, repoURL, branch, path string, body []byte) {
	t.Helper()
	contentsKey := key(repoURL, branch, path)
	if !reflect.DeepEqual(m.writtenFiles[contentsKey], body) {
		t.Fatalf("file written to %s got %v, want %v", contentsKey, m.writtenFiles[contentsKey], body)
	}
}

func (m *MockCache) AssertCommitAndPush(t *testing.T, repoURL, branch, message, token string) {
	t.Helper()
	k := key(repoURL, branch)
	commits, ok := m.commits[k]
	if !ok {
		t.Fatalf("branch %s not pushed to repo %s", branch, repoURL)
	}
	want := commit{message, token}
	for _, c := range commits {
		if c == want {
			return
		}
	}
	t.Fatalf("commit %#v not created in repo %s", want, repoURL)
}

func key(v ...string) string {
	return strings.Join(v, ":")
}
