#!/bin/bash

# Build a docker image for a Pando node.
# Usage: 
#    integration/docker/node/build.sh
#
# After the image is built, you can create a container by:
#    docker stop pando_node
#    docker rm pando_node
#    docker run -e Pando_CONFIG_PATH=/pando/integration/privatenet/node --name pando_node -it pando
set -e

SCRIPTPATH=$(dirname "$0")

echo $SCRIPTPATH

if [ "$1" =  "force" ] || [[ "$(docker images -q pando 2> /dev/null)" == "" ]]; then
    docker build -t pando -f $SCRIPTPATH/Dockerfile .
fi


