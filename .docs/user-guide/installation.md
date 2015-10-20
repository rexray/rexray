# Installation

Getting the bits, bit by bit

---

## Install via curl
The following command will download the most recent, stable build of `REX-Ray` and install it to `/usr/bin/rexray.` On Linux systems `REX-Ray` will also be registered as either a SystemD or SystemV service.

```bash
curl -sSL https://dl.bintray.com/emccode/rexray/install | sh -
```

## Install a pre-built binary
There are also pre-built binaries available for the various release types.

### [ ![Download](https://api.bintray.com/packages/emccode/rexray/unstable/images/download.svg) ](https://dl.bintray.com/emccode/rexray/unstable/latest/) Unstable
The most up-to-date, bleeding-edge, and often unstable REX-Ray binaries.

### [ ![Download](https://api.bintray.com/packages/emccode/rexray/staged/images/download.svg) ](https://dl.bintray.com/emccode/rexray/staged/latest/) Staged
The most up-to-date, release candidate REX-Ray binaries.

### [ ![Download](https://api.bintray.com/packages/emccode/rexray/stable/images/download.svg) ](https://dl.bintray.com/emccode/rexray/stable/latest/) Stable
The most up-to-date, stable REX-Ray binaries.

## Build and install from source
`REX-Ray` is also fairly simple to build from source, especially if you have `Docker` installed:

```bash
SRC=$(mktemp -d 2> /dev/null || mktemp -d -t rexray 2> /dev/null) && cd $SRC && docker run --rm -it -v $SRC:/usr/src/rexray -w /usr/src/rexray golang:1.5.1 bash -c "git clone https://github.com/emccode/rexray.git . && make build-all”
```

If you'd prefer to not use `Docker` to build `REX-Ray` then all you need is Go 1.5:

```bash
# clone the rexray repo
git clone https://github.com/emccode/rexray.git

# change directories into the freshly-cloned repo
cd rexray

# build rexray
make build-all
```

After either of the above methods for building `REX-Ray` there should be a `.bin` directory in the current directory, and inside `.bin` will be binaries for Linux-i386, Linux-x86-64,
and Darwin-x86-64.

```bash
[0]akutz@poppy:tmp.SJxsykQwp7$ ls .bin/*/rexray
-rwxr-xr-x. 1 root 14M Sep 17 10:36 .bin/Darwin-x86_64/rexray*
-rwxr-xr-x. 1 root 12M Sep 17 10:36 .bin/Linux-i386/rexray*
-rwxr-xr-x. 1 root 14M Sep 17 10:36 .bin/Linux-x86_64/rexray*
```
