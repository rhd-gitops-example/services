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

At a high level the services command currently:

- git clones the source and target repositories into ~/.promotion/cache
- creates a branch
- checks out the branch
- copies the relevant files from the cloned source into the cloned target
- pushes the cloned target
- creates a PR from the new branch in the target to master in the target

See the [tekton-example](./tekton-example/README.md) directory for more on using the `promote` task with Tekton. See [automerge](./automerge/README.md) for some suggestions as to how promotion PullRequests may be merged automatically where appropriate.

## Release process

When a new tag is pushed with the `v` prefix, a GitHub release will be created with binaries produced for 64-bit Linux, and Mac automatically.

## Experimental plugin section

Inside of the `plugin` folder you'll see documentation and other files related to using the `services` binary as a plugin to `oc`. This has been tested with the following version on OpenShift 4.3:

```
Client Version: openshift-clients-4.3.13-202004121622
Server Version: 4.3.13
Kubernetes Version: v1.16.2
```
