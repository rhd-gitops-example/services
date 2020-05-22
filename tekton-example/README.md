# Tekton Pipeline/Task Example  

An example of using `promote` in a Tekton Pipeline to promote a service's config to a GitOps repository.  Creation of the PipelineRun (using `service-promote-pipeline-run.yaml`) will drive the pipeline to clone, build and push the service and then promote the config from the local clone into your staging/test GitOps repository.

Creation of a TaskRun (using `promote-run.yaml`) will then further promote from one GitOps repo to another, e.g. from staging to production.

## Template Files

- `auth.yaml`: Creates secrets for a GitHub repository and Docker registry, an access token for the GitHub repository and the ServiceAccount
- `resources.yaml`: Creates PipelineResources for GitHub and Docker repositories

## Other Files
- `service-promote.yaml`: This is the Tekton Task used for promoting from one repository to another. It creates a PullRequest and this represents the promotion from one environment to another (for example, from development to production - in this case represented as repositories)
- `service-promote-pipeline.yaml`: Creates a pipeline that executes `build-task.yaml` and `service-promote.yaml`
- `promote.yaml`: Creates a pull request from one repository to another repository
- `build-task.yaml`: This task builds a git source into a docker image and pushes to a docker registry

## Pre-requisites

- You will need two repositories for this example, one to promote from and one to promote to 
- For the repository to promote from, an example can be forked from here: https://github.com/akihikokuroda/promote-demo
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

- Apply the resources folder:
```shell 
kubectl apply -f resources
```

- Edit both files in the template folder to contain real values. Entries of the form `REPLACE_ME.x` must be replaced with the value you wish to use, i.e at occurences such as `REPLACE_ME.IMAGE_NAME`, `REPLACE_ME.GITHUB_ORG/REPLACE_ME.GITHUB_REPO` etc... There are eight instances to replace in this folder.

- Apply the templates folder:
```shell 
kubectl apply -f templates
```

## Execute Pipeline

The PipelineRun you will create is designed to build your microservice from its development repository and then promote the new configuration to a GitOps repository (representing a different environment, for example development, staging, test or production).

- To create the PipelineRun, use:
```shell
tkn pipeline start service-promote-pipeline --resource git-source=git-app-repo --resource docker-image=docker-app-image--param commitId=v1 --param github-secret=promote-secret --param commit-name=<yourgitname> --param commit-email=<yourgitemail> --param to=https://github.com/<github username>/<github repo>.git --param service=promote-demo --workspace name=repo-space,claimName=repopvc,subPath=dir -s demo --showlog
```

This example promotes from the `promote-demo` repository, containing a service with the same name.


- This creates a PipelineRun that executes the `service-promote-pipeline`, which will build the code and promote it to a repository you have specified
- The logs will be outputted to your console, and you can also view its progress in the Tekton Dashboard.

## Promote to Next Managed Environment

Optionally, you can run a subsequent promote from one GitOps repository to another (e.g staging to prod) after merging the pull request on your first GitOps repository. For this you will need a third repository, and for this you can fork: https://github.com/rhd-gitops-example/gitops-example-staging

-  To do this second promote, you will need to create a TaskRun that executes a task promoting from a testing repository to a production repository
- To create the TaskRun (again this uses a service called `promote-demo`), use:
```shell
tkn task start promote --param github-secret=promote-secret --param from=https://github.com/<yourorg>/<yourdevrepo>.git --param to=https://github.com/<yourorg>/<yourstagingrepo>.git --param commit-name=<yourgitname> --param commit-email=<yourgitemail> --param service=promote-demo -s demo --showlog
```
This will start the TaskRun and output its logs, and you can also view its progress in the Tekton Dashboard.

The TaskRun will clone the code from the initial repository locally, build it and promote it to the final repository. This will open a pull request which you will be able to view in the repository you chose to promote to.

