# services [![Build Status](https://travis-ci.org/rhd-gitops-example/services.svg?branch=master)](https://travis-ci.org/rhd-gitops-example/services) [![Docker Repository on Quay](https://quay.io/repository/redhat-developer/gitops-cli/status "Docker Repository on Quay")](https://quay.io/repository/redhat-developer/gitops-cli)

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


## Using environments 


If an `environments` folder exists in the GitOps repository you are promoting into, and that only has one folder, the files will be copied into the destination repository's `/environments/<the only folder>` directory.

Future support is planned for an `--env` like flag which will allow us to promote from/to different repositories with multiple environments.

## Testing

Linting should be done first (this is done on Travis, and what's good locally should be good there too)

Grab the linter if you haven't already: 

```shell
GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.26.0
```

Then you can do:

```shell
golangci-lint run
```

Run the unit tests:

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

- Fork the repository https://github.com/rhd-gitops-example/gitops-example-dev
- Fork the repository https://github.com/rhd-gitops-example/gitops-example-staging
- Inside the services folder build the code: `go build ./cmd/services`
- export GITHUB_TOKEN=[your token]
- Substitute your repository URLs for those in square brackets:

```shell
./services promote --from [url.to.dev] --to [url.to.staging] --service service-a`
```

At a high level the services command currently:

- git clones the source and target repositories into ~/.promotion/cache
- creates a branch
- checks out the branch
- copies the relevant files from the cloned source into the cloned target
- pushes the cloned target
- creates a PR from the new branch in the target to master in the target

## Important notes:

- We need to remove the local cache between requests. See https://github.com/rhd-gitops-example/services/issues/20. Until then, add `rm -rf ~/.promotion/cache; ` before subsequent requests.
- See https://github.com/rhd-gitops-example/services/issues/19 for an issue related to problems 'promoting' config from a source repo into a gitops repo. 

## Release process

When a new tag is pushed with the `v` prefix, a GitHub release will be created with binaries produced for 64-bit Linux, and Mac automatically.
