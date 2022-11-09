#!/bin/bash

# Usage: 
#    integration/build/build.sh
#    integration/build/build.sh force # Always recreate docker image and container.
set -e

SCRIPTPATH=$(dirname "$0")

echo $SCRIPTPATH

if [ "$1" =  "force" ] || [[ "$(docker images -q pando_builder_image 2> /dev/null)" == "" ]]; then
    docker build -t pando_builder_image $SCRIPTPATH
fi

set +e
docker stop pando_builder
docker rm pando_builder
set -e

docker run --name pando_builder -it -v "$GOPATH:/go" pando_builder_image
