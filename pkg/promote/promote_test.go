package promote

import (
	"testing"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/google/go-cmp/cmp"
)

func TestLoadMappingFromFile(t *testing.T) {
	mapping, err := LoadMappingFromFile("testdata/env-mapping.yaml")
	if err != nil {
		t.Fatal(err)
	}

	u, ok := mapping["dev"]
	if !ok {
		t.Fatal("failed to find env 'dev' in mapping")
	}
	want := "https://example.com/testing/env-dev.git"
	if u != want {
		t.Fatalf("env 'dev' url incorrect: got %v, wanted %v", u, want)
	}

	if u, ok := mapping["unknown"]; ok {
		t.Fatalf("found mapping for 'unknown' env: %s", u)
	}
}

func TestLoadMappingFromFileWithError(t *testing.T) {
	_, err := LoadMappingFromFile("testdata/bad-mapping.yaml")
	if err == nil {
		t.Fatal(err)
	}
}

func TestMakePullRequestInput(t *testing.T) {
	pr, err := makePullRequestInput("dev", "https://example.com/project/dev-env.git", "staging", "my-test-branch")
	if err != nil {
		t.Fatal(err)
	}

	want := &scm.PullRequestInput{
		Title: "promotion from dev to staging",
		Head:  "project:my-test-branch",
		Base:  "master",
		Body:  "this is a test body",
	}

	if diff := cmp.Diff(want, pr); diff != "" {
		t.Fatalf("pull request input is different from expected: %s", diff)
	}
}
