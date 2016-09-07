# Release Notes

Release early, release often

---

## Version 0.2.0 (2016/09/07)
Beginning with this release, libStorage's versions will increment the MINOR
component with the introduction of a new storage driver in concert with the
[guidelines](http://semver.org) set forth by semantic versioning.

### New Features
* Amazon Elastic File System (EFS) Support ([#231](https://github.com/emccode/libstorage/issues/231))

### Enhancements
* Support for Go 1.7 ([#251](https://github.com/emccode/libstorage/issues/251))

### Bug Fixes
* Isilon Export Permissions ([#252](https://github.com/emccode/libstorage/issues/252), [#257](https://github.com/emccode/libstorage/issues/257))
* Isilon Volume Removal ([#253](https://github.com/emccode/libstorage/issues/253))

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
* Task service memory fix ([#225](https://github.com/emccode/libstorage/issues/225))
* Context logger optimizations ([#224](https://github.com/emccode/libstorage/issues/224))

### Enhancements
* Improved volume path caching ([#227](https://github.com/emccode/libstorage/issues/227))
* Make Gometalinter optional ([#223](https://github.com/emccode/libstorage/issues/223))


## Version 0.1.4 (2016/07/08)
This update provides a major performance improvement as well as a few other,
minor bug fixes and enhancements.

### Bug Fixes
* Performance degradation bug ([#218](https://github.com/emccode/libstorage/issues/218))
* Close bug in ScaleIO driver ([#213](https://github.com/emccode/libstorage/issues/213))
* Panic when checking attached instances with Isilon driver ([#211](https://github.com/emccode/libstorage/issues/211))

### Enhancements
* Improved build process ([#220](https://github.com/emccode/libstorage/issues/220))
* Improved executor logging ([#217](https://github.com/emccode/libstorage/issues/217))
* Log timestamps in ms ([#219](https://github.com/emccode/libstorage/issues/219))
* Updated ScaleIO docs ([#214](https://github.com/emccode/libstorage/issues/214))

## Version 0.1.3 (2016/06/14)
This is a minor update to support the release of REX-Ray 0.4.0.

### Enhancements
* Marshal to YAML Enhancements ([#203](https://github.com/emccode/libstorage/issues/203))

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
