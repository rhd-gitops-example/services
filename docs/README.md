# Tekton pipeline / task example

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
- run `docker tag <image name> <your docker hub id>/<image name>` toi tag the image
- run `docker login` to login to the docker hub
- run `docker push <your docker hub id>/<image name>` to push the image to the docker hub

## Create Tekton resource

- edit all yaml files marked as (template) and gitconfig file.  `<xxx>` must be replaced with the real value
- create a new namespace e.g. `kubectl create ns promote`
- apply auth.yaml, promotesecret.yaml, resources.yaml, build-task.yaml, servicepromote.yaml and servicepromotepipeline.yaml in the namespace e.g. `kubectl apply -n <namespace> -f <yaml file name>`
- create a configmap by `kubectl create configmap promoteconfigmap --from-file=gitconfig -n <namespace>`

## Execute pipeline

- create the servicepromotepipelinerun using the servicepromotepipelinerun.yaml e.g. `kubectl create -n <namespace> -f servicepromotepipelinerun.yaml`
