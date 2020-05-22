#!/bin/bash -ex

go build ./cmd/services
mv ./services ./oc-services
chmod +x ./oc-services
sudo mv ./oc-services /usr/local/bin/
oc services --help
