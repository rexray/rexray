#!/usr/bin/env bash

# This script runs tests

set -e
# set -x

: ${CF_STACK_NAME:="libstorage-efs-integration-test"}
: ${CF_EC2_USER:="ec2-user"}
: ${COVERPROFILE_NAME:="efs.test.out"}

TEST_BINARY="$1"

# Make sure that aws cli is installed
hash aws 2>/dev/null || {
  echo >&2 "Missing AWS command line. Please install aws cli: https://aws.amazon.com/cli/"
  exit 1
}

usage() {
  echo "Usage: ${0} test-binary"
  echo ""
  echo "   test-binary: Path to compiled and runnable golang binary"
}

if [ -z "${TEST_BINARY}" ]; then
  usage
  exit 1
fi

# Require valid test binary file
if [ ! -f "${TEST_BINARY}" ]; then
  echo >&2 "${TEST_BINARY} is not a valid file"
  exit 1
fi

echo "Waiting for CF stack to come up ..."

aws cloudformation wait stack-create-complete \
  --stack-name ${CF_STACK_NAME}

# Get IP address of EC2 machine where tests can be executed
EC2_IP_ADDRESS=$(aws cloudformation describe-stacks \
  --stack-name libstorage-efs-integration-test \
  --query 'Stacks[0].Outputs[?OutputKey==`Ec2IpAddress`].{OutputValue:OutputValue}[0].OutputValue' \
  --output text)

# Copy binary file to EC2 instance
scp $TEST_BINARY $CF_EC2_USER@$EC2_IP_ADDRESS:efs.test

# Run tests
CMD="./efs.test -test.coverprofile ${COVERPROFILE_NAME}"
ssh $CF_EC2_USER@$EC2_IP_ADDRESS $CMD

# Copy test coverage results
scp $CF_EC2_USER@$EC2_IP_ADDRESS:${COVERPROFILE_NAME} $(dirname $0)

echo "Tests passed and coverge results are available at $(dirname $0)/${COVERPROFILE_NAME}"