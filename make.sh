#!/bin/sh

# this file exists because of a requirement from Travis-CI regarding
# how a deploy script must be called

if [ "$1" = "push-docker-plugin" ]; then
    echo "pushing docker plugin"
    DOCKER_PLUGIN_TYPE=$2 exec make push-docker-plugin
fi

exec make "$@"
