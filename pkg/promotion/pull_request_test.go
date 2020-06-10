package promotion

import (
	"testing"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/google/go-cmp/cmp"
)

func TestMakePullRequestInput(t *testing.T) {
	from := EnvLocation{
		RepoPath: "https://example.com/project/dev-env.git",
		Branch:   "example",
	}
	to := EnvLocation{
		RepoPath: "https://example.com/project/prod-env.git",
		Branch:   "master",
	}
	pr, err := makePullRequestInput(from, to, "my-test-branch", "foo bar wibble")
	if err != nil {
		t.Fatal(err)
	}

	want := &scm.PullRequestInput{
		Title: "promotion from dev-env to prod-env",
		Head:  "my-test-branch",
		Base:  "master",
		Body:  "foo bar wibble",
	}

	if diff := cmp.Diff(want, pr); diff != "" {
		t.Fatalf("pull request input is different from expected: %s", diff)
	}
}
