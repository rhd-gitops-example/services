# services [![Build Status](https://travis-ci.org/rhd-gitops-example/services.svg?branch=master)](https://travis-ci.org/rhd-gitops-example/services)

A tool for promoting between GitHub repositories.

This is a pre-alpha PoC for promoting versions of files between environments, represented as repositories.

## Building

You need Go version 1.14 to build this project.

```shell
$ go build ./cmd/services
```

## Running

You'll need a GitHub token to test this out.

```shell
$ export GITHUB_TOKEN=<paste in GitHub access token>
$ ./services promote --from https://github.com/organisation/first-environment.git --to https://github.com/organisation/second-environment.git --service service-a --commit-name <User to commit as> --commit-email <Email to commit as>
```

If the `commit-name` and `commit-email` are not provided, it will attempt to find them in `~/.gitconfig`, otherwise it will fail.

This will _copy_ all files under `/services/service-a/base/config/*` in `first-environment` to `second-environment`, commit and push, and open a PR for the change.

## Testing

```shell
$ go test ./...
```

To run the complete integration tests, including pushing to the Git repository:

```shell
$ TEST_GITHUB_TOKEN=<a valid github auth token> go test ./...
```

Note that the tests in pkg/git/repository_test.go will clone and manipulate a
remote Git repository locally.

To run a particular test: for example, 

```shell
go test ./pkg/git -run TestCopyServiceWithFailureCopying
```

## Getting started

This section is temporary. To create a sample promotion Pull Request, until https://github.com/rhd-gitops-example/services/issues/8 is done:

- Copy https://github.com/mnuttall/gitops-repo-testing 
- Copy https://github.com/mnuttall/staging
- Build the code: `go build ./cmd/services`
- export GITHUB_TOKEN=[your token]
- Substitute your repository URLs for those in square brackets:

```shell
./services promote --from [https://github.com/mnuttall/gitops-repo-testing] --to [https://github.com/mnuttall/staging] --service service-a`
```

At a high level the services command currently:

- git clones the source and target repositories into ~/.promotion/cache
- creates a branch (as per the given --branch-name)
- checks out the branch
- copies the relevant files from the cloned source into the cloned target
- pushes the cloned target
- creates a PR from the new branch in the target to master in the target

Important Notes:

- We need to remove the local cache between requests. See https://github.com/rhd-gitops-example/services/issues/20. Until then, add `rm -rf ~/.promotion/cache; ` before subsequent requests.
- New pull requests need new branches (i.e you cannot run the same command twice). Add `--branch [unique branch name]` before submitting further promotion PRs. See https://github.com/rhd-gitops-example/services/issues/21.
- See https://github.com/rhd-gitops-example/services/issues/19 for an issue related to problems 'promoting' config from a source repo into a gitops repo. 