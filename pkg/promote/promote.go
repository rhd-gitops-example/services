package promote

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"

	"github.com/bigkevmcd/promote/pkg/cache"
	"github.com/bigkevmcd/promote/pkg/util"
)

// PromoteService is the main driver for promoting files between two
// repositories.
//
// It uses a GitCache to checkout the code to, and will copy the environment
// configuration for the `fromEnv` to the `toEnv` in a named branch.
//
// The Git repositories are looked up in the mapping, and a pull request is
// opened in the toEnv's project,
func PromoteService(cache cache.GitCache, token, service, fromEnv, toEnv, newBranchName string, mapping map[string]string) error {
	ctx := context.Background()
	fromURL, ok := mapping[fromEnv]
	if !ok {
		log.Fatalf("failed to find environment %s in mapping file\n", fromEnv)
	}

	toURL, ok := mapping[toEnv]
	if !ok {
		log.Fatalf("failed to find environment %s in mapping file\n", toEnv)
	}

	fileToUpdate := pathForService(service)
	newBody, err := cache.ReadFileFromBranch(ctx, fromURL, fileToUpdate, "master")
	if err != nil {
		return fmt.Errorf("failed to read the file %v from the %v environment: %s", fileToUpdate, fromEnv, err)
	}
	err = cache.CreateAndCheckoutBranch(ctx, toURL, "master", newBranchName)
	if err != nil {
		return fmt.Errorf("failed to create and checkout the new branch %v for the %v environment: %s", newBranchName, toEnv, err)
	}
	err = cache.WriteFileToBranchAndStage(ctx, toURL, newBranchName, fileToUpdate, newBody)
	if err != nil {
		return fmt.Errorf("failed to write the updated file to %v: %s", fileToUpdate, err)
	}

	err = cache.CommitAndPushBranch(ctx, toURL, newBranchName, "this is a test commit", token)
	if err != nil {
		return fmt.Errorf("failed to commit and push branch for environment %v: %s", toEnv, err)
	}

	pr, err := createPullRequest(ctx, fromEnv, fromURL, toEnv, toURL, token, newBranchName)
	if err != nil {
		// TODO: improve this error message
		return err
	}
	log.Printf("created PR %d", pr.Number)
	return nil
}

// LoadMappingFromFile takes a filename with a YAML mapping of environment name
// to git repository URLs.
func LoadMappingFromFile(fname string) (map[string]string, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	mapping := map[string]string{}
	err = dec.Decode(mapping)
	return mapping, err
}

func pathForService(s string) string {
	return fmt.Sprintf("%s/deployment.txt", s)
}

func createClient(token string) *scm.Client {
	client := github.NewDefault()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client.Client = oauth2.NewClient(context.Background(), ts)
	return client
}

// TODO: OptionFuncs for Base and Title?
// TODO: For the Head, should this try and determine whether or not this is a
// fork ("user" of both repoURLs) and if so, simplify the Head?
func makePullRequestInput(fromEnv, fromURL, toEnv, branchName string) (*scm.PullRequestInput, error) {
	title := fmt.Sprintf("promotion from %s to %s", fromEnv, toEnv)
	fromUser, _, err := util.ExtractUserAndRepo(fromURL)
	if err != nil {
		return nil, err
	}
	return &scm.PullRequestInput{
		Title: title,
		Head:  fmt.Sprintf("%s:%s", fromUser, branchName),
		Base:  "master",
		Body:  "this is a test body",
	}, nil
}

func createPullRequest(ctx context.Context, fromEnv, fromURL, toEnv, toURL, token, newBranchName string) (*scm.PullRequest, error) {
	prInput, err := makePullRequestInput(fromEnv, fromURL, toEnv, newBranchName)
	if err != nil {
		// TODO: improve this error message
		return nil, err
	}

	user, repo, err := util.ExtractUserAndRepo(toURL)
	if err != nil {
		// TODO: improve this error message
		return nil, err
	}
	// TODO: come up with a better way of generating the repo URL (this
	// only works for GitHub)
	pr, _, err := createClient(token).PullRequests.Create(ctx, fmt.Sprintf("%s/%s", user, repo), prInput)
	return pr, err
}
