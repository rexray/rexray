#!/usr/bin/env bash
set -e

hash terraform 2>/dev/null || {
  echo >&2 "You need to install terraform: https://www.terraform.io/downloads.html"
  exit 1
}

SSH_KEY_ID=$1

usage() {
  echo "Usage: ${0}" ssh-key-id
  echo ""
  echo "Requires that the DIGITALOCEAN_ACCESS_TOKEN environment variable is set\nA new server will be started in sfo2"
}

if [ -z "$DIGITALOCEAN_TOKEN" ] || [ -z "$SSH_KEY_ID" ]; then
  usage
  exit 1
fi

cd terraform && terraform destroy -force -var "ssh_key=$SSH_KEY_ID"
