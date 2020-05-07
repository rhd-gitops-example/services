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
	"github.com/rhd-gitops-example/services/test"
)

const (
	dev     = "https://example.com/testing/dev-env"
	ldev    = "/root/repo"
	staging = "https://example.com/testing/staging-env"
)

func TestPromoteWithSuccessKeepCacheTrueWithGHE(t *testing.T) {
	promoteWithSuccess(t, true, "ghe", true, "")
}

func TestPromoteWithSuccessKeepCacheFalseWithGitHub(t *testing.T) {
	promoteWithSuccess(t, false, "github", false, "")
}

func TestPromoteWithSuccessCustomMsg(t *testing.T) {
	promoteLocalWithSuccess(t, true, "custom commit message here")
}

func promoteWithSuccess(t *testing.T, keepCache bool, repoType string, tlsVerify bool, msg string) {
	dstBranch := "test-branch"
	author := &git.Author{Name: "Testing User", Email: "testing@example.com", Token: "test-token"}
	devRepo, stagingRepo := mock.New("/dev", "master"), mock.New("/environments/staging", "master")
	repos := map[string]*mock.Repository{
		mustAddCredentials(t, dev, author):     devRepo,
		mustAddCredentials(t, staging, author): stagingRepo,
	}
	sm := New("tmp", author)
	sm.repoType = repoType
	sm.tlsVerify = tlsVerify
	sm.clientFactory = func(s, ty, r string, v bool) *scm.Client {
		client, _ := fakescm.NewDefault()
		if r != repoType {
			t.Fatalf("repoType doesn't match %s != %s\n", r, repoType)
		}
		if v != tlsVerify {
			t.Fatalf("tlsVerify doesn't match in ClientFactory %v != %v\n", v, tlsVerify)
		}
		return client
	}
	sm.repoFactory = func(url, _ string, v bool, _ bool) (git.Repo, error) {
		if v != tlsVerify {
			t.Fatalf("tlsVerify doesn't match in RepoFactory %v != %v\n", v, tlsVerify)
		}
		return git.Repo(repos[url]), nil
	}
	devRepo.AddFiles("/services/my-service/base/config/myfile.yaml")

	err := sm.Promote("my-service", dev, staging, dstBranch, msg, "", keepCache)
	if err != nil {
		t.Fatal(err)
	}

	expectedCommitMsg := msg
	if msg == "" {
		commit := devRepo.GetCommitID()
		expectedCommitMsg = fmt.Sprintf("Promoting service `my-service` at commit `%s` from branch `master` in `%s`.", commit, dev)
	}

	stagingRepo.AssertBranchCreated(t, "master", dstBranch)
	stagingRepo.AssertFileCopiedInBranch(t, dstBranch, "/dev/services/my-service/base/config/myfile.yaml", "/environments/staging/services/my-service/base/config/myfile.yaml")
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
	stagingRepo := mock.New("/environments/staging", "master")
	devRepo := NewLocal("/dev")

	sm := New("tmp", author)
	sm.clientFactory = func(s, t, r string, v bool) *scm.Client {
		client, _ := fakescm.NewDefault()
		return client
	}
	sm.repoFactory = func(url, _ string, _ bool, _ bool) (git.Repo, error) {
		return git.Repo(stagingRepo), nil
	}
	sm.localFactory = func(path string, _ bool) git.Source {
		return git.Source(devRepo)
	}
	sm.debug = true
	devRepo.AddFiles("/config/myfile.yaml")

	err := sm.Promote("my-service", ldev, staging, dstBranch, msg, "", keepCache)
	if err != nil {
		t.Fatal(err)
	}

	expectedCommitMsg := msg
	if expectedCommitMsg == "" {
		expectedCommitMsg = "Promotion of service `my-service` from local filesystem directory `/root/repo`."
	}

	stagingRepo.AssertBranchCreated(t, "master", dstBranch)
	stagingRepo.AssertFileCopiedInBranch(t, dstBranch, "/dev/config/myfile.yaml", "/environments/staging/services/my-service/base/config/myfile.yaml")
	stagingRepo.AssertCommit(t, dstBranch, expectedCommitMsg, author)
	stagingRepo.AssertPush(t, dstBranch)

	if keepCache {
		stagingRepo.AssertNotDeletedFromCache(t)
	} else {
		stagingRepo.AssertDeletedFromCache(t)
	}
}

func PromoteLocalWithSuccessOneEnvAndIsUsed(t *testing.T) {
	// Destination repo (GitOps repo) to have /environments/staging
	// Promotion should copy files into that staging directory
	dstBranch := "test-branch"
	author := &git.Author{Name: "Testing User", Email: "testing@example.com", Token: "test-token"}
	devRepo, stagingRepo := mock.New("/dev", "master"), mock.New("/environments/staging", "master")
	repos := map[string]*mock.Repository{
		mustAddCredentials(t, dev, author):     devRepo,
		mustAddCredentials(t, staging, author): stagingRepo,
	}
	sm := New("tmp", author)
	sm.clientFactory = func(s, ty, r string, v bool) *scm.Client {
		client, _ := fakescm.NewDefault()
		return client
	}
	sm.repoFactory = func(url, _ string, v bool, _ bool) (git.Repo, error) {
		return git.Repo(repos[url]), nil
	}
	devRepo.AddFiles("/services/my-service/base/config/myfile.yaml")

	msg := "foo message"
	// Env not specified, it'll get picked up automatically based on dir structure
	err := sm.Promote("my-service", dev, staging, dstBranch, msg, "", false)
	if err != nil {
		t.Fatal(err)
	}

	expectedCommitMsg := "msg"
	if msg == "" {
		commit := devRepo.GetCommitID()
		expectedCommitMsg = fmt.Sprintf("Promoting service `my-service` at commit `%s` from branch `master` in `%s`.", commit, dev)
	}

	stagingRepo.AssertBranchCreated(t, "master", dstBranch)
	stagingRepo.AssertFileCopiedInBranch(t, dstBranch, "/dev/services/my-service/base/config/myfile.yaml", "/environments/staging/services/my-service/base/config/myfile.yaml")
	stagingRepo.AssertCommit(t, dstBranch, expectedCommitMsg, author)
	stagingRepo.AssertPush(t, dstBranch)
}

func PromoteLocalWithSuccessWithEnvFlag(t *testing.T) {
	// Destination repo (GitOps repo) to have /environments/staging and /environments/prod
	// Promotion should copy files into that prod directory when --env prod is used
	dstBranch := "test-branch"
	author := &git.Author{Name: "Testing User", Email: "testing@example.com", Token: "test-token"}
	devRepo, stagingRepo := mock.New("/dev", "master"), mock.New("/environments/staging", "master")
	stagingRepo.AddFiles("/environments/prod")
	repos := map[string]*mock.Repository{
		mustAddCredentials(t, dev, author):     devRepo,
		mustAddCredentials(t, staging, author): stagingRepo,
	}
	sm := New("tmp", author)
	sm.clientFactory = func(s, ty, r string, v bool) *scm.Client {
		client, _ := fakescm.NewDefault()
		return client
	}
	sm.repoFactory = func(url, _ string, v bool, _ bool) (git.Repo, error) {
		return git.Repo(repos[url]), nil
	}
	devRepo.AddFiles("/services/my-service/base/config/myfile.yaml")

	msg := "foo message"
	err := sm.Promote("my-service", dev, staging, dstBranch, msg, "prod", false)
	if err != nil {
		t.Fatal(err)
	}

	expectedCommitMsg := "msg"
	if msg == "" {
		commit := devRepo.GetCommitID()
		expectedCommitMsg = fmt.Sprintf("Promoting service `my-service` at commit `%s` from branch `master` in `%s`.", commit, dev)
	}

	stagingRepo.AssertBranchCreated(t, "master", dstBranch)
	stagingRepo.AssertFileCopiedInBranch(t, dstBranch, "/dev/services/my-service/base/config/myfile.yaml", "/environments/prod/services/my-service/base/config/myfile.yaml")
	stagingRepo.AssertCommit(t, dstBranch, expectedCommitMsg, author)
	stagingRepo.AssertPush(t, dstBranch)
}

func PromoteLocalWithSuccessFlagGetsPriority(t *testing.T) {
	// Destination repo (GitOps repo) to have /environments/staging
	// --env prod used
	// prod folder created and used
	// Destination repo (GitOps repo) to have /environments/staging and /environments/prod
	// Promotion should copy files into that prod directory when --env prod is used
	dstBranch := "test-branch"
	author := &git.Author{Name: "Testing User", Email: "testing@example.com", Token: "test-token"}
	devRepo, stagingRepo := mock.New("/dev", "master"), mock.New("/environments/staging", "master")
	repos := map[string]*mock.Repository{
		mustAddCredentials(t, dev, author):     devRepo,
		mustAddCredentials(t, staging, author): stagingRepo,
	}
	sm := New("tmp", author)
	sm.clientFactory = func(s, ty, r string, v bool) *scm.Client {
		client, _ := fakescm.NewDefault()
		return client
	}
	sm.repoFactory = func(url, _ string, v bool, _ bool) (git.Repo, error) {
		return git.Repo(repos[url]), nil
	}
	devRepo.AddFiles("/services/my-service/base/config/myfile.yaml")

	msg := "foo message"
	err := sm.Promote("my-service", dev, staging, dstBranch, msg, "prod", false)
	if err != nil {
		t.Fatal(err)
	}

	expectedCommitMsg := "msg"
	if msg == "" {
		commit := devRepo.GetCommitID()
		expectedCommitMsg = fmt.Sprintf("Promoting service `my-service` at commit `%s` from branch `master` in `%s`.", commit, dev)
	}

	stagingRepo.AssertBranchCreated(t, "master", dstBranch)
	stagingRepo.AssertFileCopiedInBranch(t, dstBranch, "/dev/services/my-service/base/config/myfile.yaml", "/environments/prod/services/my-service/base/config/myfile.yaml")
	stagingRepo.AssertCommit(t, dstBranch, expectedCommitMsg, author)
	stagingRepo.AssertPush(t, dstBranch)
}

func promoteLocalErrorsIfNoFlagMultipleEnvironments(t *testing.T) {
	// Destination repo (GitOps repo) to have /environments/staging and /environments/prod
	// No flags provided, so error
	dstBranch := "test-branch"
	author := &git.Author{Name: "Testing User", Email: "testing@example.com", Token: "test-token"}
	devRepo, stagingRepo := mock.New("/dev", "master"), mock.New("/environments/staging", "master")
	stagingRepo.AddFiles("/environments/prod")
	repos := map[string]*mock.Repository{
		mustAddCredentials(t, dev, author):     devRepo,
		mustAddCredentials(t, staging, author): stagingRepo,
	}
	sm := New("tmp", author)
	sm.clientFactory = func(s, ty, r string, v bool) *scm.Client {
		client, _ := fakescm.NewDefault()
		return client
	}
	sm.repoFactory = func(url, _ string, v bool, _ bool) (git.Repo, error) {
		return git.Repo(repos[url]), nil
	}
	devRepo.AddFiles("/services/my-service/base/config/myfile.yaml")

	msg := "foo message"
	err := sm.Promote("my-service", dev, staging, dstBranch, msg, "", false)
	if err != nil {
		t.Fatal(err)
	}
	stagingRepo.AssertBranchNotCreated(t, "master", dstBranch)
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
	sm.clientFactory = func(s, t, r string, v bool) *scm.Client {
		client, _ := fakescm.NewDefault()
		return client
	}
	sm.repoFactory = func(url, _ string, _ bool, _ bool) (git.Repo, error) {
		return git.Repo(repos[url]), nil
	}
	devRepo.AddFiles("/services/my-service/base/config/myfile.yaml")

	err := sm.Promote("my-service", dev, staging, dstBranch, "", "", false)
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
	generateBranchWithSuccess(t, repo)
}

func TestGenerateBranchForLocalSource(t *testing.T) {
	source := NewLocal("/path/to/topLevel")
	generateBranchForLocalWithSuccess(t, source)
}

func generateBranchWithSuccess(t *testing.T, repo git.Repo) {
	branch := generateBranch(repo)
	nameRegEx := "^([0-9A-Za-z]+)-([0-9a-z]{7})-([0-9A-Za-z]{5})$"
	matched, err := regexp.Match(nameRegEx, []byte(branch))
	if err != nil {
		t.Fatalf("Regexp err %s in generateBranchWithSuccess matching pattern %s to %s", err, nameRegEx, branch)
	}
	if !matched {
		t.Fatalf("generated name `%s` failed to match pattern %s", branch, nameRegEx)
	}
}

func generateBranchForLocalWithSuccess(t *testing.T, source git.Source) {
	branch := generateBranchForLocalSource(source)
	nameRegEx := "^path-to-topLevel-local-dir-([0-9A-Za-z]{5})$"
	matched, err := regexp.Match(nameRegEx, []byte(branch))
	if err != nil {
		t.Fatalf("Regexp err %s in generateBranchForLocalWithSuccess matching pattern %s to %s", err, nameRegEx, branch)
	}
	if !matched {
		t.Fatalf("generated name `%s` for local case %s failed to matching pattern %s", branch, source.GetName(), nameRegEx)
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
	path := filepath.ToSlash(s.localPath)
	path = strings.TrimLeft(path, "/")
	name := strings.ReplaceAll(path, "/", "-")
	return name
}

func (s *mockSource) AddFiles(name string) {
	if s.files == nil {
		s.files = []string{}
	}
	s.files = append(s.files, path.Join(s.localPath, name))
}

func TestRepositoryCloneErrorOmitsToken(t *testing.T) {
	dstBranch := "test-branch"
	author := &git.Author{Name: "Testing User", Email: "testing@example.com", Token: "test-token"}
	client, _ := fakescm.NewDefault()
	fakeClientFactory := func(s, t, r string, v bool) *scm.Client {
		return client
	}
	sm := New("tmp", author)
	sm.clientFactory = fakeClientFactory

	sm.repoFactory = func(url, _ string, _ bool, _ bool) (git.Repo, error) {
		// This actually causes the error and results in trying to create a repository
		// which can surface the token
		errorMessage := fmt.Errorf("failed to clone repository %s: exit status 128", dev)
		return nil, errorMessage
	}
	err := sm.Promote("my-service", dev, staging, dstBranch, "", "", false)
	if err != nil {
		devRepoToUseInError := fmt.Sprintf(".*%s", dev)
		test.AssertErrorMatch(t, devRepoToUseInError, err)
	}
}
