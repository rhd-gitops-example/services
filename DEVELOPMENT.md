# Notes for developers

## Building

You need Go version 1.13 or higher to build this project.

```shell
$ go build ./cmd/services
```

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

## Release process

When a new tag is pushed with the `v` prefix, a GitHub release will be created with binaries produced for 64-bit Linux, and Mac automatically.
