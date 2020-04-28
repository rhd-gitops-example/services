# Tekton Pipeline/Task Example  

An example of using `promote` in a Tekton Pipeline to promote a service's config to a GitOps repository.  Creation of the PipelineRun (using `service-promote-pipeline-run.yaml`) will drive the pipeline to clone, build and push the service and then promote the config from the local clone into your staging/test GitOps repository.

Creation of a TaskRun (using `promote-run.yaml`) will then further promote from one GitOps repo to another, e.g. from staging to production.

## Template Files

- `auth.yaml`: Creates secrets for a GitHub repository and Docker registry, an access token for the GitHub repository and the ServiceAccount
- `resources.yaml`: Creates PipelineResources for GitHub and Docker repositories

## Other Files
- `service-promote.yaml`: This is the Tekton Task used for promoting from one repository to another. It creates a PullRequest and this represents the promotion from one environment to another (for example, from development to production - in this case represented as repositories)
- `service-promote-pipeline.yaml`: Creates a pipeline that executes `build-task.yaml` and `service-promote.yaml`
- `service-promote-pipeline-run.yaml`: Creates a PipelineRun that executes the `service-promote-pipeline` - This will build the code and promote it to a repository you have specified
- `promote.yaml`: Creates a pull request from one repository to another repository
- `promote-run.yaml`: Creates a TaskRun that executes a task promoting from a testing repository to a production repository
- `build-task.yaml`: This task builds a git source into a docker image and pushes to a docker registry
- `gitconfig`: Data file for the configmap - includes a GitHub user ID and email address

## Pre-requisites

- You will need two repositories for this example, one to promote from and one to promote to 
- For the repository to promote from, an example can be forked from here: https://github.com/akihikokuroda/promote-demo
- For the repository to promote to, an example can be forked from here:
https://github.com/akihikokuroda/gitops-test

## Create Tekton Resources

- Create a new namespace:
```shell
kubectl create ns <namespace>
```

- Modify your Kubernetes context to use this namespace by default:
```shell 
kubectl config set-context --current --namespace=<namespace>
```

- Apply the resources folder:
```shell 
kubectl apply -f resources
```

- Edit all files in the template folder to contain real values. Entries of the form `REPLACE_ME.x` must be replaced with the value you wish to use, i.e at occurences such as `REPLACE_ME.IMAGE_NAME`, `REPLACE_ME.GITHUB_ORG/REPLACE_ME.GITHUB_REPO` etc... There are eight instances to replace in this folder.

- Apply the templates folder:
```shell 
kubectl apply -f templates
```
- Edit `gitconfig` to contain your GitHub username and email address

- Create a configmap:
```shell
kubectl create configmap promote-configmap --from-file=gitconfig
```
This will store your GitHub username and email address in key-value pairs that can be used in the PipelineRun. 

- Edit `service-promote-pipeline-run.yaml` to contain the name of the repository you want to promote to

## Execute Pipeline

`service-promote-pipeline-run` is designed to build your microservice from its development repository and then promote the new configuration to a GitOps repository (representing a different environment, for example development, staging, test or production).

- Create `service-promote-pipeline-run`:
```shell
kubectl create -f service-promote-pipeline-run.yaml
```

## Promote to Prod

You can use `promote-run.yaml` to run a subsequent promote from one GitOps repository to another (e.g staging to prod) after merging the pull request on your first GitOps repository. For this you will need a third repository, and for this you can clone: https://github.com/a-roberts/gitops-repo-testing

- Edit `promote-run.yaml` to contain the URL of the repository you want to promote from, and the URL of the repository you want to promote to

- Create the `promote-run` TaskRun:
```shell
kubectl create -f promote-run.yaml
```

You can view the progress of your PipelineRun using: 
```shell
kubectl get pod <name of pod>
```
Or alternatively view its progress in the Tekton Dashboard.

The PipelineRun will clone the code from the inital repository locally, build it and promote it to the final repository. This will open a pull request which you will be able to view in the repository you chose to promote to.

