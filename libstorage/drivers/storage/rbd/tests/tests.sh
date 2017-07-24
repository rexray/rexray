#!/bin/bash

# Show everything
set -x

# Exit on error
set -e

# install jq to make json parsing easy
sudo yum install -y jq

# start with a clean slate
EXISTING=$(docker ps --filter 'name=rbdtest' -q)
if [ ! ${EXISTING} == "" ]; then
	docker stop ${EXISTING} && docker rm ${EXISTING}
fi
EXISTING=$(docker volume ls -q)
if [ ! ${EXISTING} == "" ]; then
	docker volume rm ${EXISTING}
fi

# Make sure we can create a volume via docker plugin
docker volume create --driver rexray --name rbd_test

# Make sure we can inspect it
docker volume inspect rbd_test

docker volume rm rbd_test

# Create a volume with custom size and make sure its correct
docker volume create --driver rexray --name rbd_test --opt size=4

SIZE=$(docker volume inspect rbd_test -f '{{ .Status.size }}')
if [ ! ${SIZE} -eq 4 ]; then
	exit 1
fi

docker volume rm rbd_test

# Create an implicit volume and make sure it shows up as attached in rexray
docker run -d --volume-driver rexray -v rbd_test:/test --name rbdtest busybox tail -f /dev/null
docker volume inspect rbd_test
ATTACH_STATE=$(sudo rexray volume get rbd_test --format json | jq -r '.[0].attachmentState')
if [ ! ${ATTACH_STATE} -eq 2 ]; then
	exit 1
fi

# Check that volume is listed as unavailabe on remote hostmanager
ATTACH_STATE=$(ssh libstorage-rbd-test-client sudo rexray volume get rbd_test --format json | jq -r '.[0].attachmentState')
if [ ! ${ATTACH_STATE} -eq 4 ]; then
	exit 1
fi

# Stop docker container and verify volume is now available
docker stop rbdtest && docker rm rbdtest
ATTACH_STATE=$(sudo rexray volume get rbd_test --format json | jq -r '.[0].attachmentState')
if [ ! ${ATTACH_STATE} -eq 3 ]; then
	exit 1
fi

docker volume rm rbd_test
