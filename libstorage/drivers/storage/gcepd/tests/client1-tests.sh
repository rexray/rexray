#!/bin/bash

# Show everything
set -x

# Exit on error
set -e

# install jq to make json parsing easy
yum install -y epel-release && yum install -y jq

# start with a clean slate
EXISTING=$(docker ps --filter 'name=gcetest' -q)
if [ ! -n ${EXISTING} ]; then
	docker stop ${EXISTING} && docker rm ${EXISTING}
fi
EXISTING=$(docker volume ls -q)
if [ ! -n ${EXISTING} ]; then
	docker volume rm ${EXISTING}
fi

# Make sure we can create a volume via docker plugin
docker volume create --driver rexray --name gce-test-vol

# Make sure we can inspect it
docker volume inspect gce-test-vol

docker volume rm gce-test-vol

# Create a volume with custom size and make sure its correct
docker volume create --driver rexray --name gce-test-vol --opt size=20

SIZE=$(docker volume inspect gce-test-vol -f '{{ .Status.size }}')
if [ ! ${SIZE} -eq 20 ]; then
	exit 1
fi

docker volume rm gce-test-vol

# Create an implicit volume and make sure it shows up as attached in rexray
docker run -d --volume-driver rexray -v gce-test-vol:/test --name gcetest busybox tail -f /dev/null
docker volume inspect gce-test-vol
ATTACH_STATE=$(rexray volume get gce-test-vol --format json | jq -r '.[0].attachmentState')
if [ ! ${ATTACH_STATE} -eq 2 ]; then
	exit 1
fi

# Stop docker container and verify volume is now available
docker stop gcetest && docker rm gcetest
ATTACH_STATE=$(rexray volume get gce-test-vol --format json | jq -r '.[0].attachmentState')
if [ ! ${ATTACH_STATE} -eq 3 ]; then
	exit 1
fi

# Attach volume and leave it, so that we can try to pre-empt it from node0
rexray volume attach gce-test-vol
