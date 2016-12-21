# Release Notes

Release early, release often

---
## Version 0.4.0 (TBD)
Another exciting new feature release, this update brings with it support for
the Ceph RBD platform.

### New Features
* Ceph RBD Support ([#347](https://github.com/codedellemc/libstorage/issues/347), [#367](https://github.com/codedellemc/libstorage/issues/367))

## Version 0.3.7 (2016/12/21)
This is a minor bugfix release that includes a fix for attachment querying.

### Bug Fixes
* EFS security group ID fix ([#369](https://github.com/codedellemc/libstorage/pull/369))

## Version 0.3.6 (2016/12/13)
This is a minor release to update the build process so that smaller binaries
for embedding projects, such as REX-Ray, is supported.

### Enhancements
* Do not build Darwin executor on Travis-CI ([#362](https://github.com/codedellemc/libstorage/issues/362))

## Version 0.3.5 (2016/12/07)
This build updates the libStorage model and EBS driver to function with a
custom encryption key for encrypting volumes as well as includes a fix for
determining an EFS instance's security groups.

### Enhancements
* Custom encryption key support ([#355](https://github.com/codedellemc/libstorage/issues/355), [#358](https://github.com/codedellemc/libstorage/issues/358))
* Support for build-tag driven driver inclusion ([#356](https://github.com/codedellemc/libstorage/issues/356))

### Bug Fixes
* EFS security group ID fix ([#354](https://github.com/codedellemc/libstorage/pull/354))

## Version 0.3.4 (2016/12/02)
This is a minor release that restricts some initialization logging so
that it only appears if the environment variable `LIBSTORAGE_DEBUG` is set to a
truthy value.

### Bug Fixes
* Fix for path initialization logging ([#349](https://github.com/codedellemc/libstorage/pull/349))

### Updates
* Updated build matrix ([#350](https://github.com/codedellemc/libstorage/pull/350))

## Version 0.3.3 (2016/11/29)
This release includes some minor fixes as well as a new way to query
attachment information about one or more volumes.

### Enhancements
* Enhanced attachment querying ([#313](https://github.com/codedellemc/libstorage/pull/313), [#316](https://github.com/codedellemc/libstorage/pull/316), [#319](https://github.com/codedellemc/libstorage/pull/319), [#330](https://github.com/codedellemc/libstorage/pull/330), [#331](https://github.com/codedellemc/libstorage/pull/331), [#332](https://github.com/codedellemc/libstorage/pull/332), [#334](https://github.com/codedellemc/libstorage/pull/334),
[#335](https://github.com/codedellemc/libstorage/pull/335), [#336](https://github.com/codedellemc/libstorage/pull/336), [#343](https://github.com/codedellemc/libstorage/pull/343))

### Bug Fixes
* AWS Config Support ([#314](https://github.com/codedellemc/libstorage/pull/314))
* VirtualBox Executor Fix ([#325](https://github.com/codedellemc/libstorage/pull/325))

## Version 0.3.2 (2016/10/18)
This release updates the project to reflect its new location at
github.com/codedellemc.

### Enhancements
* Relocated to codedellemc ([#307](https://github.com/codedellemc/libstorage/pull/307))

## Version 0.3.1 (2016/10/18)
This is a minor update that includes support for ScaleIO 2.0.1.

### Enhancements
* Support for ScaleIO 2.0.1 ([#303](https://github.com/codedellemc/libstorage/issues/303))

## Version 0.3.0 (2016/10/16)
This release introduces the Elastic Block Storage (EBS) driver, formerly known
as the EC2 driver in REX-Ray <=0.3.x.

### Enhancements
* Amazon Elastic Block Storage (EBS) Support ([#248](https://github.com/codedellemc/libstorage/issues/248), [#279](https://github.com/codedellemc/libstorage/issues/279))
* Build with Docker ([#274](https://github.com/codedellemc/libstorage/issues/274), [#281](https://github.com/codedellemc/libstorage/issues/281))
* Documentation updates ([#298](https://github.com/codedellemc/libstorage/issues/298))

### Bug Fixes
* Volume Removal Instance ID Fix ([#292](https://github.com/codedellemc/libstorage/issues/292))
* Avoid Client Failure when Server Driver not Supported ([#296](https://github.com/codedellemc/libstorage/issues/296), [#297](https://github.com/codedellemc/libstorage/issues/297), [#299](https://github.com/codedellemc/libstorage/issues/299), [#300](https://github.com/codedellemc/libstorage/issues/300))

### Thank You
  Name | Blame  
-------|------
[Proud Heng](https://github.com/proudh) | So long Proud, and thanks for all the fish. EBS is now part of a tagged release!
[Aaron Spiegel](https://github.com/spiegela) | Aaron, you may be a new contributor, but I feel like we've known each other since we were kids, running around the front-yard on a summer's dusky-eve, catching fireflies and speaking of the day we'd be patching Markdown documentation together.
[Travis Rhoden](https://github.com/codenrhoden) | While we've been colleagues a while, I'm thrilled you're finally working with the rest of the nerdiest of nerds, on libStorage and the secret holographic unicorn fight club we run on Thursday nights.

## Version 0.2.1 (2016/09/14)
This is a minor release that includes a fix for the EFS storage driver as well
as improvements to the build process. For example, Travis-CI now builds
libStorage against multiple versions of Golang and both Linux and Darwin.

### Bug Fixes
* EFS Volume / Tag Creation Bug ([#261](https://github.com/codedellemc/libstorage/issues/261))
* Next Device Fix ([#268](https://github.com/codedellemc/libstorage/issues/268))

### Enhancements
* Build Matrix Support ([#263](https://github.com/codedellemc/libstorage/issues/263))
* Glide 12 Support ([#265](https://github.com/codedellemc/libstorage/issues/265))

## Version 0.2.0 (2016/09/07)
Beginning with this release, libStorage's versions will increment the MINOR
component with the introduction of a new storage driver in concert with the
[guidelines](http://semver.org) set forth by semantic versioning.

### New Features
* Amazon Elastic File System (EFS) Support ([#231](https://github.com/codedellemc/libstorage/issues/231))

### Enhancements
* Support for Go 1.7 ([#251](https://github.com/codedellemc/libstorage/issues/251))

### Bug Fixes
* Isilon Export Permissions ([#252](https://github.com/codedellemc/libstorage/issues/252), [#257](https://github.com/codedellemc/libstorage/issues/257))
* Isilon Volume Removal ([#253](https://github.com/codedellemc/libstorage/issues/253))

### Thank You
  Name | Blame  
-------|------
[Chris Duchesne](https://github.com/cduchesne) | Chris not only took on the role of project manager for libStorage and REX-Ray, he still provides ongoing test plan execution and release validation. Thank you Chris!
[Kenny Cole](https://github.com/kacole2) | Kenny's tireless effort to support users and triage submitted issues is such a cornerstone to libStorage that I'm not sure what this project would do without him!
[Martin Hrabovcin](https://github.com/mhrabovcin) | Martin, along with Kasisnu, definitely win the "Community Members of the Month" award! Their hard work and dedication resulted in the introduction of the Amazon EFS storage driver. Thank you Martin & Kasisnu!
[Kasisnu Singh](https://github.com/kasisnu) | Have I mentioned we have the best community around? Seriously, thank you again Kasisnu! Your work, along with Martin's, is a milestone in the growth of libStorage.

## Version 0.1.5 (2016/07/12)
This release comes hot on the heels of the last, but some dynamite bug fixes
have improved the performance of the server by leaps and bounds. Operations
that were taking minutes now take seconds or less. Memory consumption that
could exceed 50GB is now kept neat and tidy.

### Bug Fixes
* Task service memory fix ([#225](https://github.com/codedellemc/libstorage/issues/225))
* Context logger optimizations ([#224](https://github.com/codedellemc/libstorage/issues/224))

### Enhancements
* Improved volume path caching ([#227](https://github.com/codedellemc/libstorage/issues/227))
* Make Gometalinter optional ([#223](https://github.com/codedellemc/libstorage/issues/223))


## Version 0.1.4 (2016/07/08)
This update provides a major performance improvement as well as a few other,
minor bug fixes and enhancements.

### Bug Fixes
* Performance degradation bug ([#218](https://github.com/codedellemc/libstorage/issues/218))
* Close bug in ScaleIO driver ([#213](https://github.com/codedellemc/libstorage/issues/213))
* Panic when checking attached instances with Isilon driver ([#211](https://github.com/codedellemc/libstorage/issues/211))

### Enhancements
* Improved build process ([#220](https://github.com/codedellemc/libstorage/issues/220))
* Improved executor logging ([#217](https://github.com/codedellemc/libstorage/issues/217))
* Log timestamps in ms ([#219](https://github.com/codedellemc/libstorage/issues/219))
* Updated ScaleIO docs ([#214](https://github.com/codedellemc/libstorage/issues/214))

## Version 0.1.3 (2016/06/14)
This is a minor update to support the release of REX-Ray 0.4.0.

### Enhancements
* Marshal to YAML Enhancements ([#203](https://github.com/codedellemc/libstorage/issues/203))

## Version 0.1.2 (2016/06/13)
This release updates the default VirtualBox endpoint to `http://10.0.2.2:18083`.

## Version 0.1.1 (2016/06/10)
This is the initial GA release of libStorage.

### Features
libStorage is an open source, platform agnostic, storage provisioning and
orchestration framework, model, and API. Features include:

- A standardized storage orchestration
  [model and API](http://docs.libstorage.apiary.io/)
- A lightweight, reference client implementation with a minimal dependency
  footprint
- The ability to embed both the libStorage client and server, creating native
  application integration opportunities

### Operations
`libStorage` supports the following operations:

Resource Type | Operation | Description
--------------|-----------|------------
Volume | List / Inspect | Get detailed information about one to many volumes
       | Create / Remote | Manage the volume lifecycle
       | Attach / Detach | Provision volumes to a client
       | Mount / Unmount | Make attached volumes ready-to-use, local file systems
Snapshot | | Coming soon
Storage Pool | | Coming soon

### Getting Started
Using libStorage can be broken down into several, distinct steps:

1. Configuring [libStorage](/user-guide/config.md)
2. Understanding the [API](http://docs.libstorage.apiary.io)
3. Identifying a production server and client implementation, such as
   [REX-Ray](https://rexray.rtfd.org)

### Thank You
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

And many more...
