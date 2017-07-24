#!/usr/bin/env bash

# This script cleans up the infrastructure used for tests

set -e
# set -x

: ${CF_STACK_NAME:="libstorage-efs-integration-test"}

# Make sure that aws cli is installed
hash aws 2>/dev/null || {
  echo >&2 "Missing AWS command line. Please install aws cli: https://aws.amazon.com/cli/"
  exit 1
}

# Delete file system by cleaning up mount targets
delete_file_system() {
  local EFS_ID_DELETE=$1
  local MOUNT_TARGETS=$(aws efs describe-mount-targets \
    --file-system-id $EFS_ID_DELETE \
    --query "MountTargets[*].MountTargetId" \
    --output text)

  if [ ! -z "${MOUNT_TARGETS}" ]; then
    echo "Cleaning mount targets for EFS: ${EFS_ID_DELETE}"

    for MOUNT_TARGET in $MOUNT_TARGETS; do
      aws efs delete-mount-target \
        --mount-target-id $MOUNT_TARGET
    done

    # Wait for mount targets to clean up
    sleep 60
  fi

  echo "Deleting EFS: ${EFS_ID_DELETE}"
  aws efs delete-file-system \
    --file-system-id $EFS_ID_DELETE
}

# Get possible EFS tests leaks, i.e. binary crash and clean up remainng EFS volumes
EFS_TAG=$(aws cloudformation describe-stacks \
  --stack-name ${CF_STACK_NAME} \
  --output text \
  --query 'Stacks[0].Parameters[?ParameterKey==`EfsTag`].ParameterValue')

EFS_IDS_TO_CLEAN=$(aws efs describe-file-systems \
  --output text \
  --query "FileSystems[?starts_with(Name, '${EFS_TAG}')].FileSystemId")

# Remove any possibly uncleaned EFS filesystems by crashed tests
if [ ! -z "$EFS_IDS_TO_CLEAN" ]; then
  echo " --> Removing test EFS instance leaks"
  echo ""

  for EFS_ID in $EFS_IDS_TO_CLEAN; do
    delete_file_system $EFS_ID
  done
fi

# Get stack ID
CF_STACK_ID=$(aws cloudformation describe-stacks \
  --stack-name ${CF_STACK_NAME} \
  --output text \
  --query 'Stacks[0].StackId')

# Delete cloud formation stack
aws cloudformation delete-stack \
  --stack-name ${CF_STACK_NAME}

echo "Waiting for CF stack to get deleted ..."

aws cloudformation wait stack-delete-complete \
  --stack-name ${CF_STACK_ID}

echo "Stack has been deleted"