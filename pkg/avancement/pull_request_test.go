package avancement

import (
	"testing"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/google/go-cmp/cmp"
)

func TestMakePullRequestInput(t *testing.T) {
	pr, err := makePullRequestInput("https://example.com/project/dev-env.git", "https://example.com/project/prod-env.git", "my-test-branch")
	if err != nil {
		t.Fatal(err)
	}

	want := &scm.PullRequestInput{
		Title: "promotion from dev-env to prod-env",
		Head:  "project:my-test-branch",
		Base:  "master",
		Body:  "this is a test body",
	}

	if diff := cmp.Diff(want, pr); diff != "" {
		t.Fatalf("pull request input is different from expected: %s", diff)
	}
}
