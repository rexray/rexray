#!/bin/bash

sudo tee /etc/rexray/config.yml << EOF
rexray:
  logLevel: warn
libstorage:
  service: rbd
  volume:
    mount:
      preempt: false
EOF
