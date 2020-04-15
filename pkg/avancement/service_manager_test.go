package avancement

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
	fakescm "github.com/jenkins-x/go-scm/scm/driver/fake"
	"github.com/rhd-gitops-example/services/pkg/git"
	"github.com/rhd-gitops-example/services/pkg/git/mock"
)

const (
	dev     = "https://example.com/testing/dev-env"
	ldev    = "/root/repo"
	staging = "https://example.com/testing/staging-env"
)

func TestPromoteWithSuccessKeepCacheTrue(t *testing.T) {
	promoteWithSuccess(t, true, "")
}

func TestPromoteWithSuccessKeepCacheFalse(t *testing.T) {
	promoteWithSuccess(t, false, "")
}

func TestPromoteWithSuccessCustomMsg(t *testing.T) {
	promoteLocalWithSuccess(t, true, "custom commit message here")
}

func promoteWithSuccess(t *testing.T, keepCache bool, msg string) {
	dstBranch := "test-branch"
	author := &git.Author{Name: "Testing User", Email: "testing@example.com", Token: "test-token"}
	devRepo, stagingRepo := mock.New("/dev", "master"), mock.New("/staging", "master")
	repos := map[string]*mock.Repository{
		mustAddCredentials(t, dev, author):     devRepo,
		mustAddCredentials(t, staging, author): stagingRepo,
	}
	sm := New("tmp", author)
	sm.clientFactory = func(s string) *scm.Client {
		client, _ := fakescm.NewDefault()
		return client
	}
	sm.repoFactory = func(url, _ string, _ bool) (git.Repo, error) {
		return git.Repo(repos[url]), nil
	}
	devRepo.AddFiles("/services/my-service/base/config/myfile.yaml")

	err := sm.Promote("my-service", dev, staging, dstBranch, msg, keepCache)
	if err != nil {
		t.Fatal(err)
	}

	expectedCommitMsg := msg
	if msg == "" {
		commit := devRepo.GetCommitID()
		expectedCommitMsg = fmt.Sprintf("Promoting service `my-service` at commit `%s` from branch `master` in `%s`.", commit, dev)
	}

	stagingRepo.AssertBranchCreated(t, "master", dstBranch)
	stagingRepo.AssertFileCopiedInBranch(t, dstBranch, "/dev/services/my-service/base/config/myfile.yaml", "/staging/services/my-service/base/config/myfile.yaml")
	stagingRepo.AssertCommit(t, dstBranch, expectedCommitMsg, author)
	stagingRepo.AssertPush(t, dstBranch)

	if keepCache {
		stagingRepo.AssertNotDeletedFromCache(t)
		devRepo.AssertNotDeletedFromCache(t)
	} else {
		stagingRepo.AssertDeletedFromCache(t)
		devRepo.AssertDeletedFromCache(t)
	}
}

func TestPromoteLocalWithSuccessKeepCacheFalse(t *testing.T) {
	promoteLocalWithSuccess(t, false, "")
}

func TestPromoteLocalWithSuccessKeepCacheTrue(t *testing.T) {
	promoteLocalWithSuccess(t, true, "")
}

func TestPromoteLocalWithSuccessCustomMsg(t *testing.T) {
	promoteLocalWithSuccess(t, true, "custom commit message supplied")
}

func promoteLocalWithSuccess(t *testing.T, keepCache bool, msg string) {
	dstBranch := "test-branch"
	author := &git.Author{Name: "Testing User", Email: "testing@example.com", Token: "test-token"}
	stagingRepo := mock.New("/staging", "master")
	devRepo := NewLocal("/dev")

	sm := New("tmp", author)
	sm.clientFactory = func(s string) *scm.Client {
		client, _ := fakescm.NewDefault()
		return client
	}
	sm.repoFactory = func(url, _ string, _ bool) (git.Repo, error) {
		return git.Repo(stagingRepo), nil
	}
	sm.localFactory = func(path string, _ bool) git.Source {
		return git.Source(devRepo)
	}
	sm.debug = true
	devRepo.AddFiles("/config/myfile.yaml")

	err := sm.Promote("my-service", ldev, staging, dstBranch, msg, keepCache)
	if err != nil {
		t.Fatal(err)
	}

	expectedCommitMsg := msg
	if expectedCommitMsg == "" {
		expectedCommitMsg = "Promotion of service `my-service` from local filesystem directory `/root/repo`."
	}

	stagingRepo.AssertBranchCreated(t, "master", dstBranch)
	stagingRepo.AssertFileCopiedInBranch(t, dstBranch, "/dev/config/myfile.yaml", "/staging/services/my-service/base/config/myfile.yaml")
	stagingRepo.AssertCommit(t, dstBranch, expectedCommitMsg, author)
	stagingRepo.AssertPush(t, dstBranch)

	if keepCache {
		stagingRepo.AssertNotDeletedFromCache(t)
	} else {
		stagingRepo.AssertDeletedFromCache(t)
	}
}

func TestAddCredentials(t *testing.T) {
	testUser := &git.Author{Name: "Test User", Email: "test@example.com", Token: "test-token"}
	tests := []struct {
		repoURL string
		a       *git.Author
		want    string
	}{
		{"https://testing.example.com/test", testUser, "https://promotion:test-token@testing.example.com/test"},
		{"https://promotion:my-token@testing.example.com/test", testUser, "https://promotion:my-token@testing.example.com/test"},
		{"https://testing:atoken@testing.example.com/test", testUser, "https://testing:atoken@testing.example.com/test"},
		{"/mydir/test", testUser, "/mydir/test"},
	}

	for i, tt := range tests {
		got, err := addCredentialsIfNecessary(tt.repoURL, tt.a)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("addCredentials() %d got %s, want %s", i, got, tt.want)
		}
	}
}

func mustAddCredentials(t *testing.T, repoURL string, a *git.Author) string {
	u, err := addCredentialsIfNecessary(repoURL, a)
	if err != nil {
		t.Fatalf("failed to add credentials to %s: %e", repoURL, err)
	}
	return u
}

func TestPromoteWithCacheDeletionFailure(t *testing.T) {
	dstBranch := "test-branch"
	author := &git.Author{Name: "Testing User", Email: "testing@example.com", Token: "test-token"}
	devRepo, stagingRepo := mock.New("/dev", "master"), mock.New("/staging", "master")
	stagingRepo.DeleteErr = errors.New("failed test delete")
	repos := map[string]*mock.Repository{
		mustAddCredentials(t, dev, author):     devRepo,
		mustAddCredentials(t, staging, author): stagingRepo,
	}
	sm := New("tmp", author)
	sm.clientFactory = func(s string) *scm.Client {
		client, _ := fakescm.NewDefault()
		return client
	}
	sm.repoFactory = func(url, _ string, _ bool) (git.Repo, error) {
		return git.Repo(repos[url]), nil
	}
	devRepo.AddFiles("/services/my-service/base/config/myfile.yaml")

	err := sm.Promote("my-service", dev, staging, dstBranch, "", false)
	if err != nil {
		t.Fatal(err)
	}

	commit := devRepo.GetCommitID()
	expectedCommitMsg := fmt.Sprintf("Promoting service `my-service` at commit `%s` from branch `master` in `%s`.", commit, dev)

	stagingRepo.AssertBranchCreated(t, "master", dstBranch)
	stagingRepo.AssertFileCopiedInBranch(t, dstBranch, "/dev/services/my-service/base/config/myfile.yaml", "/staging/services/my-service/base/config/myfile.yaml")
	stagingRepo.AssertCommit(t, dstBranch, expectedCommitMsg, author)
	stagingRepo.AssertPush(t, dstBranch)

	stagingRepo.AssertNotDeletedFromCache(t)
	devRepo.AssertDeletedFromCache(t)
}

func TestGenerateBranchWithSuccess(t *testing.T) {
	repo := mock.New("/dev", "master")
	GenerateBranchWithSuccess(t, repo)
}

func TestGenerateBranchForLocalSource(t *testing.T) {
	source := NewLocal("/path/to/topLevel")
	GenerateBranchForLocalWithSuccess(t, source)
}

func GenerateBranchWithSuccess(t *testing.T, repo git.Repo) {
	branch := generateBranch(repo)
	nameRegEx := "^([0-9A-Za-z]+)-([0-9a-z]{7})-([0-9A-Za-z]{5})$"
	_, err := regexp.Match(nameRegEx, []byte(branch))
	if err != nil {
		t.Fatalf("failed to generate a branch name matching pattern %s", nameRegEx)
	}
}

func GenerateBranchForLocalWithSuccess(t *testing.T, source git.Source) {
	branch := generateBranchForLocalSource(source)
	nameRegEx := "^path-to-topLevel-local-dir-([0-9A-Za-z]{5})$"
	_, err := regexp.Match(nameRegEx, []byte(branch))
	if err != nil {
		t.Fatalf("generated name `%s` for local case %s failed to matching pattern %s", nameRegEx, source.GetName(), nameRegEx)
	}
}

type mockSource struct {
	files     []string
	localPath string
}

func NewLocal(localPath string) *mockSource {
	return &mockSource{localPath: localPath}
}

// Walk: a mock function to emulate what happens in Repository.Walk()
// The Mock version is different: it iterates over mockSource.files[] and then drives
// the visitor callback in CopyService() as usual.
//
// To preserve the same behaviour, we see that Repository Walk receives /full/path/to/repo/services/service-name
// and then calls filePath.Walk() on /full/path/to/repo/services/ .
// When CopyService() drives Walk(), 'base' is typically services/service-name
// Thus we take each /full/path/to/file/in/mockSource.files[] and split it at 'services/' as happens in the Walk() method we're mocking.
func (s *mockSource) Walk(_ string, cb func(string, string) error) error {
	if s.files == nil {
		return nil
	}
	base := filepath.Join(s.localPath, "config")

	for _, f := range s.files {
		splitString := filepath.Dir(base) + "/"
		splitPoint := strings.Index(f, splitString) + len(splitString)
		prefix := f[:splitPoint]
		name := f[splitPoint:]
		err := cb(prefix, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *mockSource) GetName() string {
	return "mock-source-name"
}

func (s *mockSource) AddFiles(name string) {
	if s.files == nil {
		s.files = []string{}
	}
	s.files = append(s.files, path.Join(s.localPath, name))
}
