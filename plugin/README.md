This folder contains an example script that will compile the services binary and name it in such a way that it can be used as a plugin to `oc` - using `oc services promote ...` for example.

A Dockerfile is also provided that will pull in `oc` and copy the built plugin binary into that image. You could then push that image to your own testing place on DOckerhub for use later - be it as a standalone container or as part of a Tekton Task.

Build it with the following command from the main `services` folder:

`docker build -f plugin/Dockerfile -t <dockerusername>/oc-services-plugin-experiment .` 

You can then do

`docker run <dockerusername>/oc-services-plugin-experiment:latest services promote`

and that's using `promote` but from being an `oc plugin` - useful for seeing if anything breaks.