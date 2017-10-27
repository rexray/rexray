# Storage Providers

Connecting storage and platforms...

---

| Provider              | Storage Platform  | <center>[Docker](https://docs.docker.com/engine/extend/plugins_volume/)</center> | <center>[CSI](https://github.com/container-storage-interface/spec)</center> | <center>Containerized</center> |
|-----------------------|----------------------|:---:|:---:|:---:|
| Amazon EC2 | [EBS](./storage-providers/aws.md#aws-ebs) | ✓ | ✓ | ✓  |
| | [EFS](./storage-providers/aws.md#aws-efs) | ✓ | ✓ | ✓ |
| | [S3FS](./storage-providers/aws.md#aws-s3fs) | ✓ | ✓ | ✓ |
| Ceph | [RBD](./storage-providers/ceph.md#ceph-rbd) | ✓ | ✓ | ✓ |
| Local | [CSI-BlockDevices](https://github.com/thecodeteam/csi-blockdevices) | | ✓ | ✓ |
| | [CSI-NFS](https://github.com/thecodeteam/csi-nfs) | ✓ | ✓ | ✓ |
| | [CSI-VFS](https://github.com/thecodeteam/csi-vfs) | | ✓ | ✓ |
| Dell EMC | [Isilon](./storage-providers/dellemc.md#dell-emc-isilon) | ✓ | ✓ | ✓ |
| | [ScaleIO](./storage-providers/dellemc.md#dell-emc-scaleio) | ✓ | ✓ | ✓ |
| DigitalOcean | [Block Storage](./storage-providers/digitalocean.md#do-block-storage) | ✓ | ✓ | ✓ |
| FittedCloud | [EBS Optimizer](./storage-providers/fittedcloud.md#ebs-optimizer) | ✓ | ✓ | |
| Google | [GCE Persistent Disk](./storage-providers/google.md#gce-persistent-disk) | ✓ | ✓ | ✓ |
| Microsoft | [Azure Unmanaged Disk](./storage-providers/microsoft.md#azure-ud) | ✓ | ✓ | |
| OpenStack | [Cinder](./storage-providers/openstack.md#cinder) | ✓ | ✓ | ✓ |
| VirtualBox | [Virtual Media](./storage-providers/virtualbox.md#virtualbox) | ✓ | ✓ | |
