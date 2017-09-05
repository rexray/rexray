# Container Storage Client
The Container Storage Client (`csc`) is a command line interface (CLI) tool
that provides analogues for all of the CSI RPCs.

```bash
$ csc
usage: csc RPC [ARGS...]

       CONTROLLER RPCs
         createvolume (new, create)
         deletevolume (d, rm, del)
         controllerpublishvolume (att, attach)
         controllerunpublishvolume (det, detach)
         validatevolumecapabilities (v, validate)
         listvolumes (l, ls, list)
         getcapacity (getc, capacity)
         controllergetcapabilities (cget)

       IDENTITY RPCs
         getsupportedversions (gets)
         getplugininfo (getp)

       NODE RPCs
         nodepublishvolume (mnt, mount)
         nodeunpublishvolume (umount, unmount)
         getnodeid (id, getn, nodeid)
         probenode (p, probe)
         nodegetcapabilities (n, node)

Use the -? flag with an RPC for additional help.
```
