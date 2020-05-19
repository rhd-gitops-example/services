# Automerge

The `promote` task initiates change by creating Pull Requests. Sometimes a team will want to merge these manually, and sometimes automation can assist. A development team may want changes to an early, shared 'dev' environment to be merged automatically, whereas changes into 'prod' will more likely require manual approval. This directory contains sample Tekton tasks that demonstrate ways to automatically merge Pull Requests. 

We have two subdirectories: 'standalone' and 'webhooks'. In the former case, a Tekton PipelineRun must be created manually. In the latter, the [Tekton Dashboard Webhooks Extension](https://github.com/tektoncd/experimental/tree/master/webhooks-extension) is used to start an automerge PipelineRun in response to a PullRequest arriving at a GitOps repository.

## Prerequisites

This example is for more advanced users. Start with the [tekton-example](../tekton-example/README.md) if you're new to the tool. You should have a good understanding of the tool, its purposes and syntax before working through the 'automerge' topic. In addition to the `services` binary you will need the following tools installed:

- [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [`tkn`](https://github.com/tektoncd/cli) if you're not using webhooks
- [`docker`](https://docs.docker.com/get-docker/)

*Note* The 'standalone' code was developed on Docker Desktop and does not yet include the Role-Based Access Control configuration necessary for it to run on OpenShift or other locked-down environments. The 'webhook' code was developed on OpenShift but used a ServiceAccount that had a generous Role attached to it. Full RBAC support should be added to this example under https://github.com/rhd-gitops-example/services/issues/77.

## Setup - both cases

Our samples currently work with GitHub. They use the `hub` CLI to create a merge commit that when pushed, will merge the associated Pull Request. See [here](https://hub.github.com/hub-merge.1.html) for more details.

## gitconfig

Since we are creating git commits within the Tekton tasks, we need to know the username and email address to associate with them, as in our ['tekton-example'](../tekton-example/README.md). So, edit `gitconfig` and

```sh
kubectl create configmap promoteconfigmap --from-file=gitconfig
```

## Dockerfile

The 'hub' CLI runs in a Docker container within a Tekton Pipeline. We've provided a sample Dockerfile. If you are taking the 'webhooks' path you can comment out the section that installs `yq`. Then, 

```sh
docker login
docker build -t YOUR_DOCKER_HUB_ID/hub-test .
docker push YOUR_DOCKER_HUB_ID/hub-test
```

## Create a Pull Request

1. Fork our example gitops repository, https://github.com/rhd-gitops-example/gitops-example-dev. 
1. Create a 'promotion' Pull Request either manually or using our ['tekton-example'](../tekton-example/README.md). For example you can use code of the form, 

```sh
services promote --from promote-demo --to https://github.com/YOUR_GITHUB_ID/gitops-example-dev.git --service promote-demo
```

## Generate a GitHub token

You will need a GitHub token with repository access in order to merge Pull Requests. 

## Standalone case

Edit the files in the 'standalone/templates' directory.

- In `automerge-task.yaml` replace `YOUR_DOCKER_HUB_ID` with your DockerHub id.
- In `git-resource.yaml` replace `YOUR_GITHUB_ID` with your GitHub id.
- In `github-secret.yaml` replace `[your-github-token-with-repo-access]` with your GitHub token.
- In pull-request.yaml replace YOUR_PULL_REQUEST_URL with your pull request URL, e.g. `https://github.com/mnuttall/gitops-example-dev/pull/22`

Apply all the Tekton resources:

```sh
kubectl apply -f standalone/resources
kubectl apply -f standalone/templates
```

Finally start the Tekton pipeline:

```sh
tkn pipeline start automerge-pipeline -r source-repo=gitops-repo -r pr=pull-request -p github-config=promoteconfigmap -p github-secret=github-secret --showlog
```

The pipeline will do a dry run to test that the yaml in the Pull Request is good, then merge the Pull Request and delete the branch associated with it.

## Tekton Dashboard Webhooks Extension

See the [Getting Started](https://github.com/tektoncd/experimental/blob/master/webhooks-extension/docs/GettingStarted.md) guide for setup guidelines. Our example uses GitHub Enterprise (GHE) and expects webhooks to be delivered to a cluster that is routeable from GHE.

### Secrets and Service Accounts

Set up your Service Account and Secret as per the guide above. In this case you should have,

- A ServiceAccount configured for use by Tekton.
- A Tekton-compatible Secret patched onto that that ServiceAccount containing your GitHub token.

This secret is used in two related ways. We check the source repository out using a Tekton Git PipelineResource. This sets up `~/.gitconfig` with the credentials needed for `git push` to work. It gets these credentials from the `accessToken` field of the relevant secret patched onto the ServiceAccount running the Tekton Task. We then extract the same field and export it into the `GITHUB_TOKEN` environment variable to make `hub merge` work. Instructions for creating this secret are in the Getting Started document linked above. You should have resources of the form, 

```yaml
---
apiVersion: v1
data:
  password: [base64-encoded token]
  username: [base64-encoded email address]
kind: Secret
metadata:
  annotations:
    tekton.dev/git-0: https://github.ibm.com  # For example
  labels:
    serviceAccount: test-sa                   # As configured in the Webhooks Extension
  name: github-repo-access-secret

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-sa
secrets:
- name: github-repo-access-secret
```

### Edit templates

Next edit the `webhooks/templates/*` files.

- In automerge-task.yaml,
  - replace `YOUR_DOCKER_HUB_ID` with your DockerHub id.
  - replace `YOUR_GHE` with your GitHub Enterprise domain.
- In automerge-tt.yaml, replace YOUR_TEKTON_SERVICE_ACCOUNT with the name of your ServiceAccount used by Tekton.

### Apply configuration

Apply your Tekton config as usual: 

```sh
kubectl apply -f webhooks/resources
kubectl apply -f webhooks/templates
```

### Set up webhook, and test

Using the Tekton Dashboard webhooks extension, associate the `automerge-pipeline` with your GitOps repository. Now when a PR is raised against that repo you should see three PipelineRuns created for `automerge-pipeline`. 

- The first is triggered when the branch for the PullRequest is created. This runs the `echo "do nothing"` section in `automerge-task`.
- The second run executes the bulk of `automerge-task`: a merge commit is created and pushed, and its associated branch deleted.

TODO:
A discrepancy occurs between running this task against GitHub and GitHub Enterprise. On GHE, while the commit is merged the PR remains open whereas on GitHub, it is shown as merged. A commented out section in `automerge-task` notes this, which will be investigated under https://github.com/rhd-gitops-example/services/issues/76.

- Finally the third run executes `echo "kubectl apply -k env"`. Were you to remove the `echo` then this would result in the updated configuration being deployed.