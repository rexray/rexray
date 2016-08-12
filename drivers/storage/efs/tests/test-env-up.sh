#!/usr/bin/env bash

# This script launches infrastructure required for testing the EFS storage driver.
# It spins up VPC and EC2 instance so AWS account owner will get charged for time
# that resources will be running.

set -e
# set -x

LAUNCH_KEY_NAME="$1"
: ${CF_STACK_NAME:="libstorage-efs-integration-test"}

# Make sure that aws cli is installed
hash aws 2>/dev/null || {
  echo >&2 "Missing AWS command line. Please install aws cli: https://aws.amazon.com/cli/"
  exit 1
}

usage() {
  echo "Usage: ${0} launch-key-name"
  echo ""
  echo "   launch-key-name: AWS key that will be used to launch EC2 instance"
}

template_path() {
  echo "$(dirname $0)/test-cf-template.json"
}

if [ -z "${LAUNCH_KEY_NAME}" ]; then
  usage
  exit 1
fi

# Launch CF stack
aws cloudformation create-stack \
  --stack-name ${CF_STACK_NAME} \
  --template-body file://$(template_path) \
  --parameter ParameterKey=KeyName,ParameterValue=${LAUNCH_KEY_NAME} \
  --capabilities CAPABILITY_IAM 1>/dev/null

echo "Environment launch started. It will take couple minutes to create whole environment..."