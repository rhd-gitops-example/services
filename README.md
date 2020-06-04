# services [![Build Status](https://travis-ci.org/rhd-gitops-example/services.svg?branch=master)](https://travis-ci.org/rhd-gitops-example/services) [![Docker Repository on Quay](https://quay.io/repository/redhat-developer/gitops-cli/status "Docker Repository on Quay")](https://quay.io/repository/redhat-developer/gitops-cli)

'GitOps' emphasises the use of Git repositories as sources of truth. Changes to a system should be made by changing its associated Git repository, rather than by operating directly on that system. This tool works with GitOps repositories that follow [this specification](https://github.com/rhd-gitops-example/docs/tree/master/model#gitops-repository). If provides a way to promote versions of files between environments, by operating on their associated Git repositories.

## Overview

At a high level the `services promote` command currently:

- git clones the source and target repositories into `~/.promotion/cache`
- creates a branch
- checks out the branch
- copies the relevant files from the cloned source into the cloned target
- pushes the cloned target
- creates a PR from the new branch in the target to master in the target

See the [tekton-example](./tekton-example/README.md) directory for more on using the `promote` task with Tekton. See [automerge](./automerge/README.md) for some suggestions as to how promotion PullRequests may be merged automatically where appropriate.

## Installation

Built versions of the tool can be obtained [here](https://github.com/rhd-gitops-example/services/releases). See [here](DEVELOPMENT.md) for instructions on building it yourself. The binary is also available as an OCI image at `quay.io/redhat-developer/gitops-cli`.

## Usage

`services` currently offers one verb: `promote`. This packages up the tasks required to take the version of a [service](https://github.com/rhd-gitops-example/docs/tree/master/model#gitops-repository) from one environment and deploy it into another. 

### Promotion uses Pull Requests in GitHub and Merge Requests in GitLab

A key point about GitOps in general is that it incorporates the ideas of workflow and approval. The `promote` task creates Pull Requests in GitHub, or Merge Requests in GitLab. In the simplest case, these will often require manual approval by someone other than the author of the change. This approval mechanism is part of how GitOps enables the idea of governance: it can be used to enforce rules about who can approve change, and under what circumstances. Thus `promote` doesn't cause deployments to happen: it creates Pull or Merge Requests. In due course these result in `git merge` operations within the associated repository, which may then cause deployment changes via automation.

### Types of promotion: 'Remote' vs 'Local'

`promote` currently handles two different scenarios. You may hear us refer to these in shorthand as 'remote' and 'local'. In the 'remote' scenario, a specific version of a service is deployed in environment A and we want to promote it into environment B. Thus we might say, "promote my-service from dev to staging." The same version of my-service will be running in both environments on the eventual completion of the whole promotion worklow - i.e. once the Pull Request has been merged and the implied deployment changes completed. We call this 'remote' because the promotion is happening between two Git repositories both remote from the machine running 'promote'. Both git repositories are referred to by https:// URLs.

We often think of promotion as a linear flow, such as 'from dev to staging to production.' The second scenario provides a means for services to enter this flow - it's about how services can get into 'dev' at all. So how can services get into 'dev'? One option is via [`odo pipelines service add`](https://github.com/rhd-gitops-example/docs/tree/master/commands/service). Under this model, having created a new service in a new source repository, a command is issued that generates a new local version of the GitOps repository, from which a pull request towards 'dev' may be manually prepared. Another option, which 'local promotion' is there to support, requires a new service, in its source repository, to contain a basic version of the kubernetes yaml that will be used to deploy it. 'Local' promotion can then be used by automation that is building the microservice, to automatically 'promote' it into the 'dev' environment. We call this 'local' promotion because the yaml in question is coming from a local directory, for example within a running Tekton build. 

### Using environments

If an `environments` folder exists in the GitOps repository you are promoting into, and that only has one folder, the files will be copied into the destination repository's `/environments/<the only folder>` directory.

Future support is planned for an `--env` like flag which will allow us to promote from/to different repositories with multiple environments.

## Example 1: Promote a service 'service-a' from 'dev' to 'staging'

You'll need a GitHub token to test this out.

```sh
export GITHUB_TOKEN=<paste in GitHub access token>
./services promote --from https://github.com/organisation/dev.git --to https://github.com/organisation/staging.git --service service-a --commit-name <User to commit as> --commit-email <Email to commit as>
```

If the `commit-name` and `commit-email` are not provided, it will attempt to find them in `~/.gitconfig`, otherwise it will fail.

This will _copy_ all files under `/environments/<env-name>/services/service-a/base/config/*` in `dev` to `staging`, commit and push, and open a PR for the change.

### Example 2: Promote (or 'publish') a microservice into 'dev'

```sh
services promote --from /workspace/service-b --to https://my.git/org/dev --service service-b --commit-message "Publish service-b into dev"
```

This will again raise a Pull Request, copying the local files `/workspace/service-b/config/*` to https://my.git/org/dev, into `/environments/<env-name>/services/service-b/base/config/*`

## CLI Reference

```sh
services promote --help
promote from one environment to another

Usage:
  services promote [flags]

Flags:
      --branch-name string       the name of the branch on the destination repository for the pull request
      --cache-dir string         where to cache Git checkouts (default "~/.promotion/cache")
      --commit-email string      the email to use for commits when creating branches
      --commit-message string    the msg to use on the resultant commit and pull request
      --commit-name string       the name to use for commits when creating branches
      --debug                    additional debug logging output
      --from string              source Git repository
      --from-branch string       branch on the source Git repository (default "master")
  -h, --help                     help for promote
      --insecure-skip-verify     Insecure skip verify TLS certificate
      --keep-cache               whether to retain the locally cloned repositories in the cache directory
      --repository-type string   the type of repository: github, gitlab or ghe (default "github")
      --service string           service name to promote
      --to string                destination Git repository
      --to-branch string         branch on the destination Git repository (default "master")
Global Flags:
      --github-token string   oauth access token to authenticate the request
```

This will _copy_ all files under `/services/service-a/base/config/*` in `first-environment` to `second-environment`, commit and push, and open a PR for the change. Any of these arguments may be provided as environment variables, using all upper case and replacing `-` with `_`. Hence you can set CACHE_DIR, COMMIT_EMAIL, etc.

- `--branch-name` : use this to override the branch name on the destination Git repository, which will otherwise be generated automatically.
- `--cache-dir` : path on the local filesystem in which Git checkouts will be cached.
- `--commit-email` : Git commits require an associated email address and username. This is the email address. May be set via ~/.gitconfig.
- `--commit-message` : use this to override the commit message which will otherwise be generated automatically.
- `--commit-name` : The other half of `commit-email`. Both must be set.
- `--debug` : prints extra debug output if true.
- `--from` : an https URL to a GitOps repository for 'remote' cases, or a path to a Git clone of a microservice for 'local' cases.
- `--from-branch` : use this to specify a branch on the source repository, instead of using the "master" branch.
- `--help`: prints the above text if true.
- `--insecure-skip-verify` : skip TLS cerificate verification if true. Do not set this to true unless you know what you are doing.
- `--keep-cache` : `cache-dir` is deleted unless this is set to true. Keeping the cache will often cause further promotion attempts to fail. This flag is mostly used along with `--debug` when investigating failure cases. 
- `--repository-type` : the type of repository: github, gitlab or ghe (default "github"). If `--from` is a Git URL, it must be of the same type as that specified via `--to`.
- `--service` : the destination path for promotion is `/environments/<env-name>/services/<service-name>/base/config/`. This argument defines `service-name` in that path.
- `--to`: an https URL to the destination GitOps repository.
- `--to-branch` : use this to specify a branch on the destination repository, instead of using the "master" branch.

### Troubleshooting

- Authentication and authorisation failures: ensure that GITHUB_TOKEN is set and has the necessary permissions.
- 'Failure to commit, Error 128'. Errors of this form are often caused by a failure to set the `commit-email` and `commmit-name` parameters.
- 'Nothing to commit'. If there's no difference between the state of `--from` and `--to` then there's no change to made, and no Git commit can be created. 
- Remote branch is created but no Pull Request. Again check GITHUB_TOKEN, and that `--repository-type` is set correctly.

## Experimental plugin section

Inside of the `plugin` folder you'll see documentation and other files related to using the `services` binary as a plugin to `oc`. This has been tested with the following version on OpenShift 4.3:

```yaml
Client Version: openshift-clients-4.3.13-202004121622
Server Version: 4.3.13
Kubernetes Version: v1.16.2
```
