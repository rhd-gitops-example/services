package promotion

import (
	"context"
	"errors"
	"io/ioutil"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
	fakescm "github.com/jenkins-x/go-scm/scm/driver/fake"
	"github.com/rhd-gitops-example/services/pkg/git"
	"github.com/rhd-gitops-example/services/pkg/git/mock"
)

const (
	exampleDevRepo         = "https://github.com/rhd-gitops-example/gitops-example-dev.git"
	exampleStagingRepo     = "https://github.com/rhd-gitops-example/gitops-example-staging.git"
	exampleAlternativeRepo = "https://github.com/JustinKuli/gitops-alternative-layouts.git"
	// TODO: move this ^ out of a personal github?
)

var (
	dev_repo           = EnvLocation{exampleDevRepo, "master", ""}
	staging_repo       = EnvLocation{exampleStagingRepo, "master", ""}
	team_a_branch      = EnvLocation{exampleAlternativeRepo, "team-a", ""}
	team_b_branch      = EnvLocation{exampleAlternativeRepo, "team-b", ""}
	dev_env            = EnvLocation{exampleAlternativeRepo, "main", "dev"}
	staging_env        = EnvLocation{exampleAlternativeRepo, "main", "staging"}
	teams_branch_a_env = EnvLocation{exampleAlternativeRepo, "team-envs", "team-a"}
)

func TestRepoPromotionPath(t *testing.T) {
	promotionPath(t, dev_repo, staging_repo, "dev")
}
func TestBranchPromotionPath(t *testing.T) {
	promotionPath(t, team_a_branch, team_b_branch, "dev")
}
func TestEnvPromotionPath(t *testing.T) {
	promotionPath(t, dev_env, staging_env, "dev")
}
func TestToMultiEnvAcrossBranchesPromotionPath(t *testing.T) {
	promotionPath(t, team_a_branch, staging_env, "dev")
}
func TestToMultiEnvAcrossReposPromotionPath(t *testing.T) {
	promotionPath(t, dev_repo, staging_env, "dev")
}
func TestFromMultiEnvAcrossReposPromotionPath(t *testing.T) {
	promotionPath(t, teams_branch_a_env, dev_repo, "team-a")
}
func TestMultiEnvToMultiEnvAcrossReposPromotionPath(t *testing.T) {
	promotionPath(t, teams_branch_a_env, dev_env, "team-a")
}

func promotionPath(t *testing.T, from, to EnvLocation, fromEnv string) {
	t.Helper()
	src, dest, clean := mustGetCaches(t, from, to)
	defer clean()

	author := &git.Author{Name: "Testing User", Email: "testing@example.com", Token: "test-token"}
	srcURL := mustAddCredentials(t, from.RepoPath, author)
	destURL := mustAddCredentials(t, to.RepoPath, author)
	fakeSCMClient, _ := fakescm.NewDefault()

	sm := New("tmp", author)
	sm.clientFactory = func(s, ty, r string, v bool) *scm.Client {
		return fakeSCMClient
	}
	sm.repoFactory = func(url, localPath string, v bool, _ bool) (git.Repo, error) {
		if url == srcURL && strings.HasSuffix(localPath, neturl.QueryEscape(from.Branch)) {
			return src, nil
		}
		if url == destURL { // This needs to handle newly created branches as well as the original destination
			return dest, nil
		}
		return nil, errors.New("Attempted to clone a repository the tests didn't specially prepare.")
	}

	testFileName := mustAddTestFile(t, src, fromEnv, "service-a", author)

	err := sm.Promote("service-a", from, to, "", "", true)
	if err != nil {
		t.Fatal("Promote failed unexpectedly: ", err)
	}

	assertTestFilePresent(t, dest, testFileName)
	assertPullRequestCorrect(t, fakeSCMClient, to.Branch)
}

func mustGetCaches(t *testing.T, from, to EnvLocation) (src, dest *git.Repository, clean func()) {
	t.Helper()
	src, cleanSrc, err := getNewCacheOf(from)
	if err != nil {
		if er := cleanSrc(); er != nil {
			t.Errorf("failed to clean cache for %#v: %v", from, er)
		}
		t.Fatalf("failed to create cache for %#v: %v", from, err)
	}
	dest, cleanDest, err := getNewCacheOf(to)
	clean = func() {
		if er := cleanSrc(); er != nil {
			t.Errorf("failed to clean cache for %#v: %v", from, er)
		}
		if er := cleanDest(); er != nil {
			t.Errorf("failed to clean cache for %#v: %v", to, er)
		}
	}
	if err != nil {
		clean()
		t.Fatalf("failed to create cache for %#v: %v", to, err)
	}
	return src, dest, clean
}

func getNewCacheOf(l EnvLocation) (r *git.Repository, clean func() error, err error) {
	clean = func() error { return nil } // no-op in case we return early.
	dir, err := ioutil.TempDir(os.TempDir(), "promote")
	if err != nil {
		return nil, clean, err
	}
	clean = func() error {
		return os.RemoveAll(dir)
	}

	r, err = git.NewRepository(l.RepoPath, dir, true, true)
	if err != nil {
		return r, clean, err
	}
	r.DisablePush() // Don't pollute the repositories with new commits/branches
	err = r.Clone()
	if err != nil {
		return r, clean, err
	}
	err = r.Checkout(l.Branch)
	return r, clean, err
}

func mustAddTestFile(t *testing.T, r *git.Repository, env, svc string, author *git.Author) string {
	t.Helper()
	f, err := ioutil.TempFile(os.TempDir(), "test")
	if err != nil {
		t.Fatal("failed to create temporary test file: ", err)
	}
	defer func() {
		err := os.Remove(f.Name())
		if err != nil {
			t.Fatal("failed to remove temporary test file: ", err)
		}
	}()

	filename := filepath.Base(f.Name())
	pathInRepo := filepath.Join("environments", env, "", "services", svc, "base", "config", filename)
	err = r.CopyFile(f.Name(), pathInRepo)
	if err != nil {
		t.Fatalf("failed to copy test file to %s in repository %v: %v", pathInRepo, r, err)
	}

	err = r.StageFiles(pathInRepo)
	if err != nil {
		t.Fatalf("failed to stage test file in repository %v: %v", r, err)
	}
	err = r.Commit("Add test file for unit test", author)
	if err != nil {
		t.Fatalf("failed to commit test file in repository %v: %v", r, err)
	}

	return filename
}

func assertTestFilePresent(t *testing.T, repo git.Source, filename string) {
	t.Helper()
	found := false
	err := repo.Walk(".", func(prefix, name string) error {
		if filepath.Base(name) == filename {
			found = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("error while walking through repository %v: %v", repo, err)
	}
	if !found {
		t.Fatal("Test file not found in destination repo cache")
	}
}

func assertPullRequestCorrect(t *testing.T, client *scm.Client, wantBase string) {
	t.Helper()
	pr, _, err := client.PullRequests.Find(context.Background(), "", 1)
	if err != nil {
		t.Fatal("Pull request not found")
	}

	if wantBase != pr.Base.Ref {
		t.Errorf("Pull Request Base got '%s', want '%s'", pr.Base.Ref, wantBase)
	}
}

func TestGetEnvironmentFolder(t *testing.T) {
	noFolders := mock.New("nonePath", "master")
	oneFolder := mock.New("onePath", "master")
	oneFolder.AddFiles("example")
	twoFolders := mock.New("twoPath", "master")
	twoFolders.AddFiles("example")
	twoFolders.AddFiles("another")

	// Tests various configurations of the repository and inputs to the function.
	// For example, if there are no environment folders in the repository, we should
	// fail every time. If there are multiple folders, then the function should only
	// succeed if we specify one that actually exists
	tests := []struct {
		repo           git.Repo
		checkFolder    string
		expectedFolder string
		expectedError  bool
	}{
		{noFolders, "", "", true},
		{noFolders, "example", "", true},
		{oneFolder, "", "example", false},
		{oneFolder, "example", "example", false}, // input matches actual folder
		{oneFolder, "another", "", true},         // input folder doesn't match actual folder
		{twoFolders, "", "", true},
		{twoFolders, "example", "example", false},
		{twoFolders, "another", "another", false},
		{twoFolders, "fake", "", true}, // input does not match an actual folder
	}

	for _, test := range tests {
		f, err := getEnvironmentFolder(test.repo, test.checkFolder)
		if test.expectedError && err == nil {
			t.Errorf("getEnvironmentFolder(%+v, '%v') gave nil err, but we wanted an error.", test.repo, test.checkFolder)
		} else if !test.expectedError && err != nil {
			t.Errorf("getEnvironmentFolder(%+v, '%v') gave error '%v', but we wanted no error.", test.repo, test.checkFolder, err)
		}
		if f != test.expectedFolder {
			t.Errorf("getEnvironmentFolder(%+v, '%v') gave folder '%v', but we wanted '%v'.", test.repo, test.checkFolder, f, test.expectedFolder)
		}
	}
}
