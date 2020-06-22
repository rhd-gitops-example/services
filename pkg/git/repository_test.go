package git

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/rhd-gitops-example/services/test"
)

const testRepository = "https://github.com/mnuttall/staging.git"

func TestRepoName(t *testing.T) {
	nameTests := []struct {
		url      string
		wantName string
		wantErr  string
	}{
		{testRepository, "staging", ""},
		{"https://github.com/mnuttall", "", "could not identify repository name: https://github.com/mnuttall"},
	}

	for _, tt := range nameTests {
		n, err := repoName(tt.url)
		if tt.wantName != n {
			t.Errorf("repoName(%s) got name %s, want %s", tt.url, n, tt.wantName)
			continue
		}
		if tt.wantErr != "" && err != nil && err.Error() != tt.wantErr {
			t.Errorf("repoName(%s) got error %s, want %s", tt.url, err, tt.wantErr)
		}
	}
}

// If the directory provided doesn't exist e.g. "~/.promotion/cache" then the
// directory is created during the clone process.
func TestCloneCreatesDirectory(t *testing.T) {
	tempDir, cleanup := makeTempDir(t)
	defer cleanup()
	r, err := NewRepository(testRepository, path.Join(tempDir, "path"), true, false)
	assertNoError(t, err)
	err = r.Clone()
	assertNoError(t, err)

	contents, err := ioutil.ReadFile(path.Join(tempDir, "path", "staging/services/service-a/base/config/deployment.txt"))
	assertNoError(t, err)

	want := "This is the staging version of this file.\n"
	if diff := cmp.Diff(want, string(contents)); diff != "" {
		t.Fatalf("failed to read file: %s", diff)
	}
}

func TestClone(t *testing.T) {
	r, cleanup := cloneTestRepository(t)
	defer cleanup()

	contents, err := ioutil.ReadFile(path.Join(r.LocalPath, "staging/services/service-a/base/config/deployment.txt"))
	assertNoError(t, err)
	want := "This is the staging version of this file.\n"
	if diff := cmp.Diff(want, string(contents)); diff != "" {
		t.Fatalf("failed to read file: %s", diff)
	}
}

func TestCloneIsIdempotent(t *testing.T) {
	r, cleanup := cloneTestRepository(t)
	defer cleanup()

	err := r.Clone()
	if err != nil {
		t.Fatalf("failed to clone repository after it had already been cloned: %v", err)
	}
}

func TestWalk(t *testing.T) {
	r, cleanup := cloneTestRepository(t)
	defer cleanup()

	visited := []string{}
	err := r.Walk("services/service-a", func(prefix, name string) error {
		visited = append(visited, name)
		return nil
	})

	assertNoError(t, err)
	want := []string{
		"service-a/base/config/300-deployment.yaml",
		"service-a/base/config/310-service.yaml",
		"service-a/base/config/deployment.txt",
		"service-a/base/config/kustomization.yaml",
		"service-a/base/kustomization.yaml",
		"service-a/overlays/kustomization.yaml",
		"service-a/overlays/staging-deployment.yaml",
		"service-a/overlays/staging-service.yaml",
	}
	if diff := cmp.Diff(want, visited); diff != "" {
		t.Fatalf("failed to read file: %s", diff)
	}
}

func TestWriteFile(t *testing.T) {
	r, cleanup := cloneTestRepository(t)
	defer cleanup()

	err := r.WriteFile(strings.NewReader("this is some text"), "services/service-a/base/config/new-file.txt")
	assertNoError(t, err)

	visited := []string{}
	err = r.Walk("services/service-a", func(prefix, name string) error {
		visited = append(visited, name)
		return nil
	})
	assertNoError(t, err)
	sort.Strings(visited)
	want := []string{
		"service-a/base/config/300-deployment.yaml",
		"service-a/base/config/310-service.yaml",
		"service-a/base/config/deployment.txt",
		"service-a/base/config/kustomization.yaml",
		"service-a/base/config/new-file.txt",
		"service-a/base/kustomization.yaml",
		"service-a/overlays/kustomization.yaml",
		"service-a/overlays/staging-deployment.yaml",
		"service-a/overlays/staging-service.yaml",
	}
	if diff := cmp.Diff(want, visited); diff != "" {
		t.Fatalf("failed to read file: %s", diff)
	}
}

func TestCopyFile(t *testing.T) {
	r, cleanup := cloneTestRepository(t)
	defer cleanup()
	tmpfile, err := ioutil.TempFile("", "source")
	assertNoError(t, err)
	defer os.Remove(tmpfile.Name())
	content := []byte(`test content`)
	_, err = tmpfile.Write(content)
	assertNoError(t, err)
	err = tmpfile.Close()
	assertNoError(t, err)

	err = r.CopyFile(tmpfile.Name(), "services/service-a/copy/copied.txt")
	assertNoError(t, err)

	visited := []string{}
	err = r.Walk("services/service-a", func(prefix, name string) error {
		visited = append(visited, name)
		return nil
	})

	assertNoError(t, err)
	sort.Strings(visited)
	want := []string{
		"service-a/base/config/300-deployment.yaml",
		"service-a/base/config/310-service.yaml",
		"service-a/base/config/deployment.txt",
		"service-a/base/config/kustomization.yaml",
		"service-a/base/kustomization.yaml",
		"service-a/copy/copied.txt",
		"service-a/overlays/kustomization.yaml",
		"service-a/overlays/staging-deployment.yaml",
		"service-a/overlays/staging-service.yaml",
	}
	if diff := cmp.Diff(want, visited); diff != "" {
		t.Fatalf("failed to read file: %s", diff)
	}
}

func TestCopyWithMissingSource(t *testing.T) {
	r, cleanup := cloneTestRepository(t)
	defer cleanup()
	tempDir, cleanup := makeTempDir(t)
	defer cleanup()

	err := r.CopyFile(path.Join(tempDir, "unknown.txt"), "service-a/copy/copied.txt")
	test.AssertErrorMatch(t, "failed to get permissions for existing file.*stat.*no such file or directory", err)
}

func TestStageFiles(t *testing.T) {
	r, cleanup := cloneTestRepository(t)
	defer cleanup()
	err := r.WriteFile(strings.NewReader("this is some text"), "services/service-a/new-file.txt")
	assertNoError(t, err)

	err = r.StageFiles("services/service-a/new-file.txt")
	assertNoError(t, err)

	out := assertExecGit(t, r, r.repoPath("services/service-a"), "status", "--porcelain")
	want := "A  services/service-a/new-file.txt\n"
	if diff := cmp.Diff(want, string(out)); diff != "" {
		t.Fatalf("file status not modified: %s", diff)
	}
}

// The output of the git log -n looks like this:
// commit c88ebbcdef14604aed32f20166582be2762348fd (HEAD -> master)
// Author: Git User <user@example.com>
// Date:   Wed Mar 18 11:48:20 2020 +0000
//
//    testing.
func TestCommit(t *testing.T) {
	r, cleanup := cloneTestRepository(t)
	defer cleanup()
	err := r.WriteFile(strings.NewReader("this is some text"), "services/service-a/new-file.txt")
	assertNoError(t, err)
	err = r.StageFiles("services/service-a/new-file.txt")
	assertNoError(t, err)

	err = r.Commit("this is a test commit", &Author{Name: "Test User", Email: "testing@example.com"})
	assertNoError(t, err)

	out := strings.Split(string(assertExecGit(t, r, r.repoPath("services/service-a"), "log", "-n", "1")), "\n")
	want := []string{"Author: Test User <testing@example.com>", "    this is a test commit"}
	if diff := cmp.Diff(want, out, cmpopts.IgnoreSliceElements(func(s string) bool {
		return strings.HasPrefix(s, "commit") || strings.HasPrefix(s, "Date:") || s == ""
	})); diff != "" {
		t.Fatalf("file commit match failed: %s", diff)
	}
}

func TestExecGit(t *testing.T) {
	r, cleanup := cloneTestRepository(t)
	r.debug = true
	captured := ""
	r.logger = func(f string, v ...interface{}) {
		captured = fmt.Sprintf(f, v...)
	}
	defer cleanup()
	_, err := r.execGit(r.repoPath(), nil, "unknown")
	test.AssertErrorMatch(t, "exit status 1", err)

	want := "DEBUG: git: 'unknown' is not a git command. See 'git --help'."
	if strings.TrimSpace(captured) != want {
		t.Fatalf("execGit debug log failed: got %s, want %s", captured, want)
	}
}

func TestPush(t *testing.T) {
	if authToken() == "" {
		t.Skip("no auth token to push the branch upstream")
	}
	r, cleanup := cloneTestRepository(t)
	defer cleanup()
	err := r.CheckoutAndCreate("my-new-branch")
	assertNoError(t, err)
	err = r.WriteFile(strings.NewReader("this is some text"), "services/service-a/new-file.txt")
	assertNoError(t, err)
	err = r.StageFiles("services/service-a/new-file.txt")
	assertNoError(t, err)
	err = r.Commit("this is a test commit", &Author{Name: "Test User", Email: "testing@example.com"})
	assertNoError(t, err)

	err = r.Push("my-new-branch")
	assertNoError(t, err)
}

func TestDebugEnabled(t *testing.T) {
	tempDir, cleanup := makeTempDir(t)
	defer cleanup()
	r, err := NewRepository(testRepository, path.Join(tempDir, "path"), false, true)
	assertNoError(t, err)
	if !r.debug {
		t.Fatalf("Debug not set to true")
	}
}

func cloneTestRepository(t *testing.T) (*Repository, func()) {
	tempDir, cleanup := makeTempDir(t)
	r, err := NewRepository(authenticatedURL(t), tempDir, false, false)
	assertNoError(t, err)
	err = r.Clone()
	assertNoError(t, err)
	return r, cleanup
}

func authenticatedURL(t *testing.T) string {
	t.Helper()
	parsed, err := url.Parse(testRepository)
	if err != nil {
		t.Fatalf("failed to parse git repo url %v: %v", testRepository, err)
	}
	parsed.User = url.UserPassword("promotion", authToken())
	return parsed.String()
}

func makeTempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := ioutil.TempDir(os.TempDir(), "promote")
	assertNoError(t, err)
	return dir, func() {
		err := os.RemoveAll(dir)
		assertNoError(t, err)
	}
}

func assertExecGit(t *testing.T, r *Repository, gitPath string, args ...string) []byte {
	t.Helper()
	out, err := r.execGit(gitPath, nil, args...)
	if err != nil {
		t.Fatalf("assertExecGit failed: %s (%s)", err, out)
	}
	return out
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func authToken() string {
	return os.Getenv("TEST_GITHUB_TOKEN")
}
