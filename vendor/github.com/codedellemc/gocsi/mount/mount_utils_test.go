package mount

import (
	"strings"
	"testing"
)

func TestReadProcMountsFrom(t *testing.T) {
	r := strings.NewReader(procMountInfoData)
	mis, _, err := readProcMountsFrom(r, true, procMountsFields)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("len(mounts)=%d", len(mis))
	success1 := false
	success2 := false
	for _, mi := range mis {
		t.Logf("%+v", mi)
		if mi.Path == "/home/akutz/2" && mi.Source == "/home/akutz/1" {
			success1 = true
		}
		if mi.Path == "/home/akutz/travis-is-right" && mi.Source == "/dev/sda1" {
			success2 = true
		}
	}
	if !(success1 && success2) {
		t.FailNow()
	}
}

const procMountInfoData = `17 60 0:16 / /sys rw,nosuid,nodev,noexec,relatime shared:6 - sysfs sysfs rw,seclabel
18 60 0:3 / /proc rw,nosuid,nodev,noexec,relatime shared:5 - proc proc rw
19 60 0:5 / /dev rw,nosuid shared:2 - devtmpfs devtmpfs rw,seclabel,size=1930460k,nr_inodes=482615,mode=755
20 17 0:15 / /sys/kernel/security rw,nosuid,nodev,noexec,relatime shared:7 - securityfs securityfs rw
21 19 0:17 / /dev/shm rw,nosuid,nodev shared:3 - tmpfs tmpfs rw,seclabel
22 19 0:11 / /dev/pts rw,nosuid,noexec,relatime shared:4 - devpts devpts rw,seclabel,gid=5,mode=620,ptmxmode=000
23 60 0:18 / /run rw,nosuid,nodev shared:23 - tmpfs tmpfs rw,seclabel,mode=755
24 17 0:19 / /sys/fs/cgroup ro,nosuid,nodev,noexec shared:8 - tmpfs tmpfs ro,seclabel,mode=755
25 24 0:20 / /sys/fs/cgroup/systemd rw,nosuid,nodev,noexec,relatime shared:9 - cgroup cgroup rw,xattr,release_agent=/usr/lib/systemd/systemd-cgroups-agent,name=systemd
26 17 0:21 / /sys/fs/pstore rw,nosuid,nodev,noexec,relatime shared:20 - pstore pstore rw
27 24 0:22 / /sys/fs/cgroup/cpu,cpuacct rw,nosuid,nodev,noexec,relatime shared:10 - cgroup cgroup rw,cpuacct,cpu
28 24 0:23 / /sys/fs/cgroup/hugetlb rw,nosuid,nodev,noexec,relatime shared:11 - cgroup cgroup rw,hugetlb
29 24 0:24 / /sys/fs/cgroup/perf_event rw,nosuid,nodev,noexec,relatime shared:12 - cgroup cgroup rw,perf_event
30 24 0:25 / /sys/fs/cgroup/net_cls,net_prio rw,nosuid,nodev,noexec,relatime shared:13 - cgroup cgroup rw,net_prio,net_cls
31 24 0:26 / /sys/fs/cgroup/blkio rw,nosuid,nodev,noexec,relatime shared:14 - cgroup cgroup rw,blkio
32 24 0:27 / /sys/fs/cgroup/devices rw,nosuid,nodev,noexec,relatime shared:15 - cgroup cgroup rw,devices
33 24 0:28 / /sys/fs/cgroup/pids rw,nosuid,nodev,noexec,relatime shared:16 - cgroup cgroup rw,pids
34 24 0:29 / /sys/fs/cgroup/freezer rw,nosuid,nodev,noexec,relatime shared:17 - cgroup cgroup rw,freezer
35 24 0:30 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:18 - cgroup cgroup rw,memory
36 24 0:31 / /sys/fs/cgroup/cpuset rw,nosuid,nodev,noexec,relatime shared:19 - cgroup cgroup rw,cpuset
58 17 0:32 / /sys/kernel/config rw,relatime shared:21 - configfs configfs rw
60 1 253:0 / / rw,relatime shared:1 - xfs /dev/mapper/cl-root rw,seclabel,attr2,inode64,noquota
37 17 0:14 / /sys/fs/selinux rw,relatime shared:22 - selinuxfs selinuxfs rw
38 18 0:33 / /proc/sys/fs/binfmt_misc rw,relatime shared:24 - autofs systemd-1 rw,fd=25,pgrp=1,timeout=300,minproto=5,maxproto=5,direct
39 17 0:6 / /sys/kernel/debug rw,relatime shared:25 - debugfs debugfs rw
40 19 0:34 / /dev/hugepages rw,relatime shared:26 - hugetlbfs hugetlbfs rw,seclabel
41 19 0:13 / /dev/mqueue rw,relatime shared:27 - mqueue mqueue rw,seclabel
72 60 8:1 / /boot rw,relatime shared:28 - xfs /dev/sda1 rw,seclabel,attr2,inode64,noquota
74 60 253:2 / /home rw,relatime shared:29 - xfs /dev/mapper/cl-home rw,seclabel,attr2,inode64,noquota
150 60 253:0 /var/lib/docker/devicemapper /var/lib/docker/devicemapper rw,relatime - xfs /dev/mapper/cl-root rw,seclabel,attr2,inode64,noquota
109 23 0:35 / /run/user/1000 rw,nosuid,nodev,relatime shared:62 - tmpfs tmpfs rw,seclabel,size=388200k,mode=700,uid=1000,gid=1000
116 38 0:36 / /proc/sys/fs/binfmt_misc rw,relatime shared:66 - binfmt_misc binfmt_misc rw
113 17 0:37 / /sys/fs/fuse/connections rw,relatime shared:65 - fusectl fusectl rw
119 74 253:2 /akutz/1 /home/akutz/2 rw,relatime shared:29 - xfs /dev/mapper/cl-home rw,seclabel,attr2,inode64,noquota
119 74 0:5 /sda1 /home/akutz/travis-is-right rw,nosuid shared:2 - devtmpfs devtmpfs rw,seclabel,size=1930460k,nr_inodes=482615,mode=755
`
