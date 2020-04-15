# Tekton pipeline / task example  

A small example of using promote in a tekton pipeline to promote a service's config to a gitops repo.  Creation of the
pipelinerun (using servicepromotepipelinerun.yaml) will drive the pipeline to clone, build and push the service and then promote the config from the local clone into your staging/test gitops repo.

Creation of a taskrun (using the promoterun.yaml) will then further promote from one gitops repo to another, i.e from
staging to production.

## Files

- auth.yaml: (template) create secrets for github and docker registry and create the service account
- gitconfig: (template) data file for the gitconfig configmap.  The config map is created by `kubectl create configmap promoteconfigmap --from-file=gitconfig`

- promotesecret.yaml: (template) create an access token secret for the github repository
- resources.yaml: (template) create pipeline resources for github and docker repository

- build-task.yaml: create a build push task
- servicepromote.yaml: (template)create a promote from service repo to env repo task
- servicepromotepipeline.yaml: create a pipeline that executes build, push and promote
- servicepromotepipelinerun.yaml: create a pipelinerun that executes the servicepromotepipeline

- promote.yaml: (template)create a promote from one env repo to another env repo task
- promoterun.yaml: create a taskrun that execute promote task

## Build docker image with `service promote` command

- clone this repository
- run `docker build -t <image name> .` in repository root directory
- run `docker tag <image name> <your docker hub id>/<image name>` to tag the image
- run `docker login` to login to the docker hub
- run `docker push <your docker hub id>/<image name>` to push the image to the docker hub

## Create Tekton resource

- edit all yaml files marked as (template) in the section above and also the gitconfig file.  Entries of the form `<some property>` must be replaced with the real value, i.e at occurences such as `<image name>`, `<github org>/<github repo>` etc...
- create a new namespace e.g. `kubectl create ns promote`
- apply auth.yaml, promotesecret.yaml, resources.yaml, build-task.yaml, servicepromote.yaml, servicepromotepipeline.yaml and promote.yaml in the namespace e.g. `kubectl -n <namespace> apply -f <yaml file name>`
- create a configmap by `kubectl create configmap promoteconfigmap --from-file=gitconfig -n <namespace>`

## Execute pipeline

The servicepromotepipelinerun is designed to build your microservice from its development repository and then promote the new configuration to a gitops repository (some dev/staging/test environment).

- create the servicepromotepipelinerun using the servicepromotepipelinerun.yaml e.g. `kubectl create -n <namespace> -f servicepromotepipelinerun.yaml`

## Promote to Prod

You can use the promoterun taskrun to run a subsequent promote from one gitops repo to another, e.g staging to prod, after merging the pull request on your first gitops repository.

- create the promoterun taskrun by running `kubectl create -n <namespace> -f promoterun.yaml`
