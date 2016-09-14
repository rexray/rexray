# Release Notes

---

## Upgrading

To upgrade REX-Ray to the latest version, use `curl install`:

    curl -sSL https://dl.bintray.com/emccode/rexray/install | sh

Use `rexray version` to determine the currently installed version of REX-Ray:

    $ rexray version
    REX-Ray
    -------
    Binary: /Users/akutz/Projects/go/bin/rexray
    SemVer: 0.4.0
    OsArch: Linux-x86_64
    Branch: v0.4.0
    Commit: c83f0237e60792cfe89c4255d7149b5670965539
    Formed: Mon, 20 Jun 2016 20:56:48 CDT

    libStorage
    ----------
    SemVer: 0.1.3
    OsArch: Linux-x86_64
    Branch: v0.1.3
    Commit: 182a626937677a081b89651598ee2eac839308e7
    Formed: Wed, 15 Jun 2016 16:27:36 CDT

## Version 0.5.1 (2016/09/14)
This is a minor release, but includes a few important patches.

### Enhancements
* libStorage 0.2.1 ([#docs](http://libstorage.readthedocs.io/en/v0.2.1))
* ScaleIO 2.0.0.2 Support ([#555](https://github.com/emccode/rexray/issues/555))

### Bug Fixes
* EFS Volume / Tag Creation Bug ([#261](https://github.com/emccode/libstorage/issues/261))

## Version 0.5.0 (2016/09/07)
Beginning with this release, REX-Ray's versions will increment the MINOR
component with the introduction of a new storage driver via libStorage in
concert with the [guidelines](http://semver.org) set forth by semantic
versioning.

### New Features
* Amazon Elastic File System (EFS) Support ([#525](https://github.com/emccode/rexray/issues/525))

### Enhancements
* Support for Go 1.7 ([#541](https://github.com/emccode/rexray/issues/541))
* Enhanced Isilon Support ([#520](https://github.com/emccode/rexray/issues/520), [#521](https://github.com/emccode/rexray/issues/521))

### Thank You
  Name | Blame  
-------|------
[Chris Duchesne](https://github.com/cduchesne) | Chris not only took on the role of project manager for libStorage and REX-Ray, he still provides ongoing test plan execution and release validation. Thank you Chris!
[Kenny Cole](https://github.com/kacole2) | Kenny's tireless effort to support users and triage submitted issues is such a cornerstone to libStorage and REX-Ray that I'm not sure what this project would do without him!
[Martin Hrabovcin](https://github.com/mhrabovcin) | Martin, along with Kasisnu, definitely win the "Community Members of the Month" award! Their hard work and dedication resulted in the introduction of the Amazon EFS storage driver. Thank you Martin & Kasisnu!
[Kasisnu Singh](https://github.com/kasisnu) | Have I mentioned we have the best community around? Seriously, thank you again Kasisnu! Your work, along with Martin's, is a milestone in the growth of libStorage and REX-Ray.

## Version 0.4.2 (2016/07/12)
This minor update represents a *major* performance boost for REX-Ray.
Operations that use to take up to minutes now take seconds or less. The memory
footprint has been reduced from the magnitude of phenomenal cosmic powers to
the size of an itty bitty living space!

### Enhancements
* libStorage 0.1.5 ([#TBA](https://github.com/emccode/rexray/issues/TBA))
* Improved volume path caching ([#500](https://github.com/emccode/rexray/issues/500))

## Version 0.4.1 (2016/07/08)
Although a minor release, 0.4.1 provides some meaningful and useful enhancements
and fixes, further strengthening the foundation of the REX-Ray platform.

### Enhancements
* Improved build process ([#474](https://github.com/emccode/rexray/issues/474), [#492](https://github.com/emccode/rexray/issues/492))
* [libStorage](http://libstorage.readthedocs.io) 0.1.4 ([#493](https://github.com/emccode/rexray/issues/493))
* Removed Docker spec file ([#486](https://github.com/emccode/rexray/issues/486))
* Improved REX-Ray 0.3.3 Config Backwards Compatibility ([#481](https://github.com/emccode/rexray/issues/481))
* Improved install script ([#439](https://github.com/emccode/rexray/issues/439), [#495](https://github.com/emccode/rexray/issues/495))

### Bug Fixes
* Fixed input validation bug when creating volume sans name ([#478](https://github.com/emccode/rexray/issues/478))

## Version 0.4.0 (2016/06/20)
REX-Ray 0.4.0 introduces centralized configuration and control along with
a new client/server architecture -- features made possible by
[libStorage](http://libstorage.readthedocs.io). Users are no longer
required to configure storage drivers or store privileged information on all
systems running the REX-Ray client. The new client delegates storage-platform
related operations to a remote, libStorage-compatible server such as REX-Ray
or [Poly](https://github.com/emccode/polly).

Please note that the initial release of REX-Ray 0.4 includes support for only
the following storage platforms:

* ScaleIO
* VirtualBox

Support for the full compliment of drivers present in earlier versions of
REX-Ray will be reintroduced over the course of several, incremental updates,
beginning with 0.4.1.

### New Features
* Distributed architecture ([#399](https://github.com/emccode/rexray/issues/399), [#401](https://github.com/emccode/rexray/issues/401), [#411](https://github.com/emccode/rexray/issues/411), [#417](https://github.com/emccode/rexray/issues/417), [#418](https://github.com/emccode/rexray/issues/418), [#419](https://github.com/emccode/rexray/issues/419), [#420](https://github.com/emccode/rexray/issues/420), [#423](https://github.com/emccode/rexray/issues/423))
* Volume locking mechanism ([#171](https://github.com/emccode/rexray/issues/171))
* Volume creation with initial data ([#169](https://github.com/emccode/rexray/issues/169))

### Enhancements
* Improved storage driver logging ([#396](https://github.com/emccode/rexray/issues/396))
* Docker mount path ([#403](https://github.com/emccode/rexray/issues/403))

### Bug Fixes
* Fixed issue with install script ([#409](https://github.com/emccode/rexray/issues/409))
* Fixed volume ls filter ([#400](https://github.com/emccode/rexray/issues/400))
* Fixed panic during access attempt of offline REX-Ray daemon ([#148](https://github.com/emccode/rexray/issues/148))

### Thank You
Yes, the author is so lazy as to blatantly
[copy](http://libstorage.readthedocs.io/en/stable/about/release-notes/#version-011-20160610)
this section. So sue me :)

  Name | Blame  
-------|------
[Clint Kitson](https://github.com/clintonskitson) | His vision come to fruition. That's __his__ vision, thus please assign __all__ bugs to Clint :)
[Vladimir Vivien](https://github.com/vladimirvivien) | A nascent player, Vlad had to hit the ground running and has been a key contributor
[Kenny Coleman](https://github.com/kacole2) | While some come close, none are comparable to Kenny's handlebar
[Jonas Rosland](https://github.com/jonasrosland) | Always good for a sanity check and keeping things on the straight and narrow
[Steph Carlson](https://github.com/stephcarlson) | Steph keeps the convention train chugging along...
[Amanda Katona](https://github.com/amandakatona) | And Amanda is the one keeping the locomotive from going off the rails
[Drew Smith](https://github.com/mux23) | Drew is always ready to lend a hand, no matter the problem
[Chris Duchesne](https://github.com/cduchesne) | His short time with the team is in complete opposition to the value he has added to this project
[David vonThenen](https://github.com/dvonthenen) | David has been a go-to guy for debugging the most difficult of issues
[Steve Wong](https://github.com/cantbewong) | Steve stays on top of the things and keeps use cases in sync with industry needs
[Travis Rhoden](https://github.com/codenrhoden) | Another keen mind, Travis is also a great font of technical know-how
[Peter Blum](https://github.com/oskoss) | Absent Peter, the EMC World demo would not have been ready
[Megan Hyland](https://github.com/meganmurawski) | And absent Megan, Peter's work would only have taken things halfway there
[Eugene Chupriyanov](https://github.com/echupriyanov) | For helping with the EC2 planning
[Matt Farina](https://github.com/mattfarina) | Without Glide, it all comes crashing down
Josh Bernstein | The shadowy figure behind the curtain...

## Version 0.3.3 (2016/04/21)

### New Features
* ScaleIO v2 support ([#355](https://github.com/emccode/rexray/issues/355))
* EC2 Tags added to Volumes & Snapshots ([#314](https://github.com/emccode/rexray/issues/314))

### Enhancements
* Use of official Amazon EC2 SDK ([#359](https://github.com/emccode/rexray/issues/359))
* Added a disable feature for create/remove volume ([#366](https://github.com/emccode/rexray/issues/366))
* Added ScaleIO troubleshooting information ([#367](https://github.com/emccode/rexray/issues/367))

### Bug Fixes
* Fixes URLs for documentation when viewed via Github ([#337](https://github.com/emccode/rexray/issues/337))
* Fixes logging bug on Ubuntu 14.04 ([#377](https://github.com/emccode/rexray/issues/377))
* Fixes module start timeout error ([#376](https://github.com/emccode/rexray/issues/376))
* Fixes ScaleIO authentication loop bug ([#375](https://github.com/emccode/rexray/issues/375))

### Thank You
* [Philipp Franke](https://github.com/philippfranke)
* [Eugene Chupriyanov](https://github.com/echupriyanov)
* [Peter Blum](https://github.com/oskoss)
* [Megan Hyland](https://github.com/meganmurawski)

## Version 0.3.2 (2016-03-04)

### New Features
* Support for Docker 1.10 and Volume Plugin Interface 1.2 ([#273](https://github.com/emccode/rexray/issues/273))
* Stale PID File Prevents Service Start ([#258](https://github.com/emccode/rexray/issues/258))
* Module/Personality Support ([#275](https://github.com/emccode/rexray/issues/275))
* Isilon Preemption ([#231](https://github.com/emccode/rexray/issues/231))
* Isilon Snapshots ([#260](https://github.com/emccode/rexray/issues/260))
* boot2Docker Support ([#263](https://github.com/emccode/rexray/issues/263))
* ScaleIO Dynamic Storage Pool Support ([#267](https://github.com/emccode/rexray/issues/267))

### Enhancements
* Improved installation documentation ([#331](https://github.com/emccode/rexray/issues/331))
* ScaleIO volume name limitation ([#304](https://github.com/emccode/rexray/issues/304))
* Docker cache volumes for path operations ([#306](https://github.com/emccode/rexray/issues/306))
* Config file validation ([#312](https://github.com/emccode/rexray/pull/312))
* Better logging ([#296](https://github.com/emccode/rexray/pull/296))
* Documentation Updates ([#285](https://github.com/emccode/rexray/issues/285))

### Bug Fixes
* Fixes issue with daemon process getting cleaned as part of SystemD Cgroup ([#327](https://github.com/emccode/rexray/issues/327))
* Fixes regression in 0.3.2 RC3/RC4 resulting in no log file ([#319](https://github.com/emccode/rexray/issues/319))
* Fixes no volumes returned on empty list ([#322](https://github.com/emccode/rexray/issues/322))
* Fixes "Unsupported FS" when mounting/unmounting with EC2 ([#321](https://github.com/emccode/rexray/issues/321))
* ScaleIO re-authentication issue ([#303](https://github.com/emccode/rexray/issues/303))
* Docker XtremIO create volume issue ([#307](https://github.com/emccode/rexray/issues/307))
* Service status is reported correctly ([#310](https://github.com/emccode/rexray/pull/310))

### Updates
* <del>Go 1.6 ([#308](https://github.com/emccode/rexray/pull/308))</del>

### Thank You
* Dan Forrest
* Kapil Jain
* Alex Kamalov


## Version 0.3.1 (2015-12-30)

### New Features
* Support for VirtualBox ([#209](https://github.com/emccode/rexray/issues/209))
* Added Developer's Guide ([#226](https://github.com/emccode/rexray/issues/226))

### Enhancements
* Mount/Unmount Accounting ([#212](https://github.com/emccode/rexray/issues/212))
* Support for Sub-Path Volume Mounts / Permissions ([#215](https://github.com/emccode/rexray/issues/215))

### Milestone Issues
This release also includes many other small enhancements and bug fixes. For a
complete list click [here](https://github.com/emccode/rexray/pulls?q=is%3Apr+is%3Aclosed+milestone%3A0.3.1).

### Downloads
Click [here](https://dl.bintray.com/emccode/rexray/stable/0.3.1/) for the 0.3.1
binaries.

## Version 0.3.0 (2015-12-08)

### New Features
* Pre-Emption support ([#190](https://github.com/emccode/rexray/issues/190))
* Support for VMAX ([#197](https://github.com/emccode/rexray/issues/197))
* Support for Isilon ([#198](https://github.com/emccode/rexray/issues/198))
* Support for Google Compute Engine (GCE) ([#194](https://github.com/emccode/rexray/issues/194))

### Enhancements
* Added driver example configurations ([#201](https://github.com/emccode/rexray/issues/201))
* New configuration file format ([#188](https://github.com/emccode/rexray/issues/188))

### Tweaks
* Chopped flags `--rexrayLogLevel` becomes `logLevel` ([#196](https://github.com/emccode/rexray/issues/196))

### Pre-Emption Support
Pre-Emption is an important feature when using persistent volumes and container
schedulers.  Without pre-emption, the default behavior of the storage drivers is
to deny the attaching operation if the volume is already mounted elsewhere.  
If it is desired that a host should be able to pre-empt from other hosts, then
this feature can be used to enable any host to pre-empt from another.

### Milestone Issues
This release also includes many other small enhancements and bug fixes. For a
complete list click [here](https://github.com/emccode/rexray/pulls?q=is%3Apr+is%3Aclosed+milestone%3A0.3.0).

### Downloads
Click [here](https://dl.bintray.com/emccode/rexray/stable/0.3.0/) for the 0.3.0
binaries.

## Version 0.2.1 (2015-10-27)
REX-Ray release 0.2.1 includes OpenStack support, vastly improved documentation,
and continued foundation changes for future features.

### New Features
* Support for OpenStack ([#111](https://github.com/emccode/rexray/issues/111))
* Create volume from volume using existing settings ([#129](https://github.com/emccode/rexray/issues/129))

### Enhancements
* A+ [GoReport Card](http://goreportcard.com/report/emccode/rexray)
* A+ [Code Coverage](https://coveralls.io/github/emccode/rexray?branch=master)
* [GoDoc Support](https://godoc.org/github.com/emccode/rexray)
* Ability to load REX-Ray as an independent storage platform ([#127](https://github.com/emccode/rexray/issues/127))
* New documentation at http://rexray.readthedocs.org ([#145](https://github.com/emccode/rexray/issues/145))
* More foundation updates

### Tweaks
* Command aliases for `get` and `delete` - `ls` and `rm` ([#107](https://github.com/emccode/rexray/issues/107))

## Version 0.2.0 (2015-09-30)

### Installation, SysV, SystemD Support
REX-Ray now includes built-in support for installing itself as a service on
Linux distributions that support either SystemV or SystemD initialization
systems. This feature has been tested successfully on both CentOS 7 Minimal
(SystemD) and Ubuntu 14.04 Server (SystemV) distributions.

To install REX-Ray on a supported Linux distribution, all that is required
now is to download the binary and execute:

    sudo ./rexray service install

What does that do? In short the above command will determine if the Linux
distribution uses systemctl, update-rc.d, or chkconfig to manage system
services. After that the following steps occur:

 1. The path /opt/rexray is created and chowned to root:root with permissions
 set to 0755.
 2. The binary is copied to /opt/rexray/rexray and chowned to root:root with
 permissions set to 4755. This is important, because this means that any
 non-privileged user can execute the rexray binary as root without requiring
 sudo privileges. For more information on this feature, please read about the
 [Linux kernel's super-user ID (SUID) bit](http://www.tldp.org/HOWTO/Security-HOWTO/file-security.html).

 Because the REX-Ray binary can now be executed with root privileges by
 non-root users, the binary can be used by non-root users to easily attach
 and mount external storage.

 3. The directory /etc/rexray is created and chowned to root:root.

The next steps depends on the type of Linux distribution. However, it's
important to know that the new version of the REX-Ray binary now supports
managing its own PID (at `/var/run/rexray.pid`) when run as a service as well
as supports the standard SysV control commands such as `start`, `stop`,
`status`, and `restart`.

For SysV Linux distributions that use `chkconfig` or `update-rc.d`, a symlink
of the REX-Ray binary is created in `/etc/init.d` and then either
`chkconfig rexray on` or `update-rc.d rexray defaults` is executed.

Modern Linux distributions have moved to SystemD for controlling services.
If the `systemctl` command is detected when installing REX-Ray then a unit
file is written to `/etc/systemd/system/rexray.servic`e with the following
contents:

    [Unit]
    Description=rexray
    Before=docker.service

    [Service]
    EnvironmentFile=/etc/rexray/rexray.env
    ExecStart=/usr/local/bin/rexray start -f
    ExecReload=/bin/kill -HUP $MAINPID
    KillMode=process

    [Install]
    WantedBy=docker.service

The REX-Ray service is not started immediately upon installation. The install
command completes by informing the users that they should visit the
[REX-Ray website](http://github.com/emccode/rexray) for information on how to
configure REX-Ray's storage drivers. The text to the users also explains how
to start the REX-Ray service once it's configured using the service command
particular to the Linux distribution.

### Single Service
This release also removes the need for REX-Ray to be configured as multiple
service instances in order to provide multiple end-points to such consumers
such as `Docker`. REX-Ray's backend now supports an internal, modular design
which enables it to host multiple module instances of any module, such as the
DockerVolumeDriverModule. In fact, one of the default, included modules is...

### Admin Module & HTTP JSON API
The AdminModule enables an HTTP JSON API for managing REX-Ray's module system
as well as provides a UI to view the currently running modules. Simply start
the REX-Ray server and then visit the URL http://localhost:7979 in your favorite
browser to see what's loaded. Or you can access either of the currently
supported REST URLs:

    http://localhost:7979/r/module/types

and

    http://localhost:7979/r/module/instances

Actually, those aren't the *only* two URLs, but the others are for internal
users as of this point. However, the source *is* open, so... :)

If you want to know what modules are available by using the CLI, after starting
the REX-Ray service simply type:

    [0]akutz@poppy:rexray$ rexray service module types
    [
      {
        "id": 2,
        "name": "DockerVolumeDriverModule",
        "addresses": [
          "unix:///run/docker/plugins/rexray.sock",
          "tcp://:7980"
        ]
      },
      {
        "id": 1,
        "name": "AdminModule",
        "addresses": [
          "tcp://:7979"
        ]
      }
    ]
    [0]akutz@poppy:rexray$

To get a list of the *running* modules you would type:

    [0]akutz@poppy:rexray$ rexray service module instance get
    [
      {
        "id": 1,
        "typeId": 1,
        "name": "AdminModule",
        "address": "tcp://:7979",
        "description": "The REX-Ray admin module",
        "started": true
      },
      {
        "id": 2,
        "typeId": 2,
        "name": "DockerVolumeDriverModule",
        "address": "unix:///run/docker/plugins/rexray.sock",
        "description": "The REX-Ray Docker VolumeDriver module",
        "started": true
      },
      {
        "id": 3,
        "typeId": 2,
        "name": "DockerVolumeDriverModule",
        "address": "tcp://:7980",
        "description": "The REX-Ray Docker VolumeDriver module",
        "started": true
      }
    ]
    [0]akutz@poppy:rexray$

Hmmm, you know, the REX-Ray CLI looks a little different in the above examples,
doesn't it? About that...

### Command Line Interface
The CLI has also been enhanced to present a more simplified view up front to
users. The commands are now categorized into logical groups:

    [0]akutz@pax:~$ rexray
    REX-Ray:
      A guest-based storage introspection tool that enables local
      visibility and management from cloud and storage platforms.

    Usage:
      rexray [flags]
      rexray [command]

    Available Commands:
      volume      The volume manager
      snapshot    The snapshot manager
      device      The device manager
      adapter     The adapter manager
      service     The service controller
      version     Print the version
      help        Help about any command

    Global Flags:
      -c, --config="/Users/akutz/.rexray/config.yaml": The REX-Ray configuration file
      -?, --help[=false]: Help for rexray
      -h, --host="tcp://:7979": The REX-Ray service address
      -l, --logLevel="info": The log level (panic, fatal, error, warn, info, debug)
      -v, --verbose[=false]: Print verbose help information

    Use "rexray [command] --help" for more information about a command.

### Travis-CI Support
REX-Ray now supports Travis-CI builds either from the primary REX-Ray repository
or via a fork. All builds should be executed through the Makefile, which is a
Travis-CI default. For the Travis-CI settings please be sure to set the
environment variable `GO15VENDOREXPERIMENT` to `1`.
