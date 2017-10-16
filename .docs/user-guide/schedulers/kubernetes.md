# Kubernetes

Containers cubed

---

## Overview
REX-Ray can be integrated with [Kubernetes](https://kubernetes.io/) allowing
pods to consume data stored on volumes that are orchestrated by REX-Ray. Using
Kubernetes' [FlexVolume](https://kubernetes.io/docs/user-guide/volumes/#flexvolume)
plug-in, REX-Ray can provide uniform access to storage operations such as attach,
mount, detach, and unmount for any configured storage provider.  REX-Ray provides an
adapter script called `FlexREX` which integrates with the FlexVolume to interact
with the backing storage system.

### Pre-Requisites
- [Kubernetes](https://kubernetes.io/) 1.5 or higher
- REX-Ray 0.7 or higher
- [jq binary](https://stedolan.github.io/jq/)
- Kubernetes kubelets must be running with `enable-controller-attach-detach` disabled

### Installation
It is assumed that you have a Kubernetes cluster at your disposal. On each
Kubernetes node (running the kubelet), do the followings:

- Install and configure the REX-Ray binary as prescribed in the
[*Installation*](../installation.md) section.  
- Next, validate the REX-Ray installation by running `rexray volume ls`
as shown in the the following:

```
# rexray volume ls
ID                Name   Status     Size
925def7200000006  vol01  available  32
925def7100000005  vol02  available  32
```

If there is no issue, you should see an output, similar to above, which shows
a list of previously created volumes. If instead you get an error,  
ensure that REX-Ray is properly configured for the intended storage system.

Next, using the REX-Ray binary,  install the `FlexREX` adapter script on the node
as shown below.  

```
# rexray flexrex install
```

This should produce the following output showing that the FlexREX script is
installed successfully:

```
Path                                                                        Installed  Modified
/usr/libexec/kubernetes/kubelet-plug-ins/volume/exec/rexray~flexrex/flexrex  true       false
```

The path shown above is the default location where the FlexVolume plug-in will
expect to find its integration code.  If you are not using the default location
with FlexVolume, you can install the  `FlexREX` in an arbitrary location using:

```
# rexray flexrex install --path /opt/plug-ins/rexray~flexrex/flexrex
```

!!! note
    FlexREX requires that the `enable-controller-attach-detach` flag for the
    kubelet is set to False.

Next, restart the kubelet process on the node:

```
# systemctl restart kubelet
```

You can validate that the `FlexREX` script has been started successfully by searching
the kubelet log for an entry similar to the following:

```
I0208 10:56:57.412207    5348 plug-ins.go:350] Loaded volume plug-in "rexray/flexrex"
```

### Pods and Persistent Volumes
You can now deploy pods and persistent volumes that use storage systems orchestrated
by REX-Ray.  It is worth pointing out that the Kubernetes FlexVolume plug-in can only
attach volumes that already exist in the storage system.  Any volume that is to be used
by a Kubernetes resource must be listed in a `rexray volume ls` command.

#### Pod with REX-Ray volume
The following YAML file shows the definition of a pod that uses `FlexREX` to attach a volume
to be used by the pod.

```
apiVersion: v1
kind: Pod
metadata:
  name: pod-0
spec:
  containers:
  - image: gcr.io/google_containers/test-webserver
    name: pod-0
    volumeMounts:
    - mountPath: /test-pd
      name: vol-0
  volumes:
  - name: vol-0
    flexVolume:
      driver: rexray/flexrex
      fsType: ext4
      options:
        volumeID: test-vol-1
        forceAttach: "true"
        forceAttachDelay: "15"
```
Notice in the section under `flexVolume` the name of the driver attribute
`driver: rexray/flexrex`. This is used by the FlexVolume plug-in to launch REX-Ray.
Additional options can be provided in the `options:` as follows:

Option|Desription
------|----------
volumeID|Reference name of the volume in REX-Ray (Required)
forceAttach|When true ensures the volume is available before attaching (optional, defaults to false)
forceAttachDelay|Total amount of time (in sec) to attempt attachment with 5 sec interval between tries (optional)

#### REX-Ray PersistentVolume
The next example shows a YAML definition of Persistent Volume (PV) managed
by REX-Ray.

```
apiVersion: v1
kind: PersistentVolume
metadata:
  name: vol01
spec:
  capacity:
    storage: 32Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  flexVolume:
    driver: rexray/flexrex
    fsType: xfs
    options:
      volumeID: redis01
      forceAttach: "true"
      forceAttachDelay: "15"
```

The next YAML shows a `Persistent Volume Claim` (PVC) that carves out `10Gi` out of
the PV defined above.

```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: vol01
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

The claim can then be used by a pod in a YAML definition as shown below:

```
apiVersion: v1
kind: Pod
metadata:
  name: pod-1
spec:
  containers:
  - image: gcr.io/google_containers/test-webserver
    name: pod-1
    volumeMounts:
    - mountPath: /test-pd
      name: vol01
  volumes:
  - name: vol01
    persistentVolumeClaim:
      claimName: vol01
```
