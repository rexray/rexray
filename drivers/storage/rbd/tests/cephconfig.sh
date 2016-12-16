#!/bin/bash
set -x

NUM_NODES=$1

if [ -e /etc/ceph/ceph.conf ]; then
  echo "skipping Ceph config because it's been done before"
  exit 0
fi

tee ~/.ssh/config << EOF
Host *
   StrictHostKeyChecking no
   UserKnownHostsFile=/dev/null
EOF

chmod 0600 ~/.ssh/config

#Configure Ceph

if [ $NUM_NODES -ge 3 ]; then
	ceph-deploy new libstorage-rbd-test-server1 libstorage-rbd-test-server2 libstorage-rbd-test-server3
else
	ceph-deploy new libstorage-rbd-test-server1
fi

if [ $NUM_NODES == 1 ]; then
	tee -a ceph.conf << EOF
osd pool default size = 1
osd pool default min size = 1
osd crush chooseleaf type = 0
EOF

elif [ $NUM_NODES == 2 ]; then
	tee -a ceph.conf << EOF
osd pool default size = 2
osd pool default min size = 1
EOF

fi

ceph-deploy mon create-initial
for x in $(seq 1 $NUM_NODES); do
	ceph-deploy osd create --zap-disk libstorage-rbd-test-server$x:/dev/sdb
done
ceph-deploy admin localhost libstorage-rbd-test-client
