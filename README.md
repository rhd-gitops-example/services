# services [![Build Status](https://travis-ci.org/rhd-gitops-example/services.svg?branch=master)](https://travis-ci.org/rhd-gitops-example/services)

A tool for promoting between GitHub repositories.

This is a pre-alpha PoC for promoting versions of files between environments, represented as repositories.

## Building

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

This will _copy_ a single file `deployment.txt` from `service-a` in `first-environment` to `service-a` in `second-environment`, commit and push, and open a PR for the change.

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
