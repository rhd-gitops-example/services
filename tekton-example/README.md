# Tekton Pipeline/Task Example  

An example of using `promote` in a Tekton Pipeline to promote a service's config to a GitOps repository.  Creation of the PipelineRun will drive the pipeline to clone, build and push the service and then promote the config from the local clone into your staging/test GitOps repository.

## Tekton APIs: v1alpha1 and v1beta1

Tekton Pipelines introduced new [v1beta1](https://github.com/tektoncd/pipeline/blob/master/docs/migrating-v1alpha1-to-v1beta1.md) APIs with its 0.11.x release. We first developed this sample against v1alpha1 APIs, but now include both v1alpha1 and v1beta1 versions of this sample. The v1beta1 version uses no Tekton PipelineResources, in the spirit of [this](https://github.com/tektoncd/pipeline/blob/master/docs/migrating-v1alpha1-to-v1beta1.md#replacing-pipelineresources-with-tasks) section of the migration document. You should use Tekton Pipelines v0.12.1 or higher and the v1beta1 samples unless you are unable to do so, or wish to use the code in the final section, 'Promote to Next Managed Environment'. This has only been implemented under the v1alpha1 sample. We're still working through how promotion between GitOps repositories should work, so this last section is fairly experimental.

## Template Files

- `auth.yaml`: Creates secrets for a GitHub repository and image registry, an access token for the GitHub repository and the ServiceAccount
- `resources.yaml`: Creates PipelineResources for GitHub and Docker repositories. This file only exists in the v1alpha1 sample.

## Other Files

- `service-promote.yaml`: This is the Tekton Task used for promoting from one repository to another. It creates a PullRequest and this represents the promotion from one environment to another (for example, from development to production - in this case represented as repositories)
- `service-promote-pipeline.yaml`: Creates a Pipeline that executes `build-task.yaml` and `service-promote.yaml`.
- `promote.yaml`: Creates a pull request from one repository to another repository.
- `build-task.yaml`: This task builds a Git source into a container image image and pushes to an image registry
- `git-clone.yaml`: A copy of the Tekton Catalog [git-clone](https://github.com/tektoncd/catalog/blob/v1beta1/git/git-clone.yaml) Task, replacing the Git PipelineResource used in v1alpha1.

## Pre-requisites

- You will need two repositories for this example, one to promote from and one to promote to 
- For the repository to promote from, an example can be forked from here: https://github.com/rhd-gitops-example/promote-demo
- For the repository to promote to, an example can be forked from here:
 https://github.com/rhd-gitops-example/gitops-example-dev
- You will also need to have the latest release of the Tekton CLI, which can be downloaded from here: https://github.com/tektoncd/cli

## Create Tekton Resources

- Create a new namespace:

```shell
kubectl create ns <namespace>
```

- Modify your Kubernetes context to use this namespace by default:

```shell
kubectl config set-context --current --namespace=<namespace>
```

- Choose whether to use the `v1alpha1` or `v1beta1` resources. As per the introduction, we recommend Tekton 0.12.x+ and the `v1beta1` path unless you have clear reasons to choose `v1alpha1`.

```sh
cd v1beta1
# OR
cd v1alpha1
```

- Apply the resources folder:

```shell
kubectl apply -f resources
```

- Edit the files in the template folder to contain real values. Entries of the form `REPLACE_ME.x` must be replaced with the value you wish to use, i.e at occurences such as `REPLACE_ME.IMAGE_NAME`, `REPLACE_ME.GITHUB_ORG/REPLACE_ME.GITHUB_REPO` etc... There are eight instances to replace in the v1alpha1 templates/ folder and nine in the v1beta1 templates/ folder.

- If you are using v1alpha1 you can apply the templates folder:

```shell
kubectl apply -f templates
```

If you are using v1beta1 then only apply templates/auth.yaml. The templates/ folder also contains a PipelineRun which we don't want to run yet.

```sh
kubectl apply -f templates/auth.yaml
```

## Execute Pipeline

The PipelineRun you will create is designed to build your microservice from its development repository and then promote the new configuration to a GitOps repository (representing a different environment, for example development, staging, test or production).

This example promotes from the `promote-demo` repository, containing a service with the same name.

- This creates a PipelineRun that executes the `service-promote-pipeline`, which will build the code and promote it to a repository you have specified
- The logs will be outputted to your console, and you can also view its progress in the Tekton Dashboard.

The exact steps to create a PipelineRun depend on whether you are using the v1alpha1 or v1beta1 APIs.

### v1alpha1

To create the PipelineRun using v1alpha1, use:

```shell
tkn pipeline start service-promote-pipeline --resource git-source=git-app-repo --resource docker-image=docker-app-image --param commitId=v1 --param github-secret=promote-secret --param commit-name=<yourgitname> --param commit-email=<yourgitemail> --param to=https://github.com/<github username>/<github repo>.git --param service=promote-demo --workspace name=repo-space,claimName=repopvc,subPath=dir -s demo --showlog
```

### v1beta1

One of the main differences between our v1alpha1 and v1beta1 samples is that v1beta1 does not use a Tekton Git PipelineResource. Instead we use Tekton workspaces, which must be backed by persistent storage since they contain more than one task. Persistent storage requires a PersistentVolumeClaim. These can be dynamically generated using the `volumeClaimTemplate` stanza, but this is not supported in `tkn` until https://github.com/tektoncd/cli/issues/1006 is resolved. In the meantime we provide a PipelineRun:

```sh
kubectl create -f templates/promote-pipelinerun.yaml
```

You will need to locate and tail the logs of this PipelineRun for yourself. For example you can use, 

```sh
kubectl get pipelineruns
tkn pipelinerun logs [pipelinerun] -f
```

## Promote to Next Managed Environment

Optionally, you can run a subsequent promote from one GitOps repository to another (e.g staging to prod) after merging the pull request on your first GitOps repository. For this you will need a third repository, and for this you can fork: https://github.com/rhd-gitops-example/gitops-example-staging

- To do this second promote, you will need to create a TaskRun that executes a task promoting from a testing repository to a production repository
- To create the TaskRun (again this uses a service called `promote-demo`), use:

```shell
tkn task start promote --param github-secret=promote-secret --param from=https://github.com/<yourorg>/<yourdevrepo>.git --param to=https://github.com/<yourorg>/<yourstagingrepo>.git --param commit-name=<yourgitname> --param commit-email=<yourgitemail> --param service=promote-demo -s demo --showlog
```

This will start the TaskRun and output its logs, and you can also view its progress in the Tekton Dashboard.

The TaskRun will clone the code from the initial repository locally, build it and promote it to the final repository. This will open a pull request which you will be able to view in the repository you chose to promote to.
