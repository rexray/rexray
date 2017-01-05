#!/bin/bash

# Show everything
set -x

# install jq to make json parsing easy
yum install -y epel-release && yum install -y jq

## These tests assume that the tests for node1 have already been run
# Check that volume is listed as unavailable
ATTACH_STATE=$(rexray volume get gce-test-vol --format json | jq -r '.[0].attachmentState')
if [ ! ${ATTACH_STATE} -eq 4 ]; then
	exit 1
fi

# Make sure that we cannot attach volume to ourselves without force
rexray volume attach gce-test-vol
if [ $? -eq 0 ]; then
	exit 1
fi

# Make sure we can attach with force
rexray volume attach --force gce-test-vol
if [ $? -ne 0 ]; then
	exit 1
fi

rexray volume rm gce-test-vol
