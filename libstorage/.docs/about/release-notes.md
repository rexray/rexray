# Release Notes

Release early, release often

---
## Version 0.6.3 (2017/07/22)
This is the last public, tagged release of libStorage. The project will be
absorbed into [REX-Ray](https://github.com/thecodeteam/rexray).

## Version 0.6.2 (2017/06/28)
This is a minor release that improves volume lookup response for GCEPD
and  DOBS, with minor enhancements to the EFS, RBD, and Integration
drivers.

### Bug Fixes
* Cinder driver no longer complains about invalid local devices ([#578](https://github.com/codedellemc/libstorage/pull/578))

### Enhancements
* libStorage documentation is now searchable ([#574](https://github.com/codedellemc/libstorage/pull/574))
* Add config option for force remove of volume from integration driver ([#577](https://github.com/codedellemc/libstorage/pull/577))
* Add ability to disable `modprobe` call from RBD driver ([#576](https://github.com/codedellemc/libstorage/pull/576))
* Add support for VolumeInsepctByName to GCEPD and DOBS drivers ([#579](https://github.com/codedellemc/libstorage/pull/579), [#581](https://github.com/codedellemc/libstorage/pull/581))
* Add config option to translate volume names containing underscores to dashes for GCEPD and DOBS ([#580](https://github.com/codedellemc/libstorage/pull/580), [#582](https://github.com/codedellemc/libstorage/pull/582))

## Version 0.6.1 (2017/06/09)
This is a minor release that includes bug fixes for RBD and Isilon, and some
minor enhancements.

### Bug Fixes
* Fix handling of white space in Ceph config file for monitor hosts ([#551](https://github.com/codedellemc/libstorage/issues/551))
* Fix volume create for Isilon storage ([#556](https://github.com/codedellemc/libstorage/issues/556))

### Enhancements
* Introduce ability for storage driver to implement VolumeInspectByName ([#560](https://github.com/codedellemc/libstorage/issues/560))
* Add ability to enable TLS over UNIX Sockets ([#546](https://github.com/codedellemc/libstorage/pull/546))
* ARM build support ([#553](https://github.com/codedellemc/libstorage/pull/553))

## Version 0.6.0 (2017/05/03)
This release introduces support for the Cinder storage driver and
multiple security-related enhancements, including default-to-TLS for
libStorage client/server communications, and service-scoped
authentication!

### New Features
* Client Token Authentication ([#475](https://github.com/codedellemc/libstorage/issues/475))
* Cinder storage driver ([#182](https://github.com/codedellemc/libstorage/issues/182))
* Allow customization of default paths ([#509](https://github.com/codedellemc/libstorage/pull/509))
* TLS Known Hosts support ([#510](https://github.com/codedellemc/libstorage/pull/510))

### Bug Fixes
* Return HTTP status 400 instead of 500 when attachment mask requires InstanceID or LocalDevices header and it is missing ([#352](https://github.com/codedellemc/libstorage/issues/352))
* Make sure all drivers return error if VolumeInspect doesn't find volume ([#396](https://github.com/codedellemc/libstorage/issues/396))
* Ensure all drivers reject size 0 volume creation ([#459](https://github.com/codedellemc/libstorage/issues/459))
* Prevent possible endless loops in drivers when underlying API does not respond ([#480](https://github.com/codedellemc/libstorage/issues/480))
* Standardize log levels across libStorage client and server ([#521](https://github.com/codedellemc/libstorage/pull/521))

### Enhancements
* Digital Ocean Block Storage driver now supports client/server topology ([#432](https://github.com/codedellemc/libstorage/issues/432))
* Improve error reporting ([#504](https://github.com/codedellemc/libstorage/pull/504), [#128](https://github.com/codedellemc/libstorage/issues/128))
* Improve driver config examples ([#531](https://github.com/codedellemc/libstorage/issues/531))

### Thank You
  Name | Blame  
-------|------
[Mathieu Velten](https://github.com/MatMaul) | Mr. Velten, as his people alert you to the fact that he insists on being addressed, is a dubious individual. It's apparent he's old money, but it's also not exactly clear from where his fortune originated. There are rumors in the back rooms of the shadiest gambling parlors of Monte Carlo that Mr. Velten was once an employee of an unnamed wing of a shadow government. A "cleaner" if you will. Maybe it was these experiences that make Mr. Velten so apt at slicing up Git commits. Is there really any difference between slicing up a full-grown man and hash series of changes? Mr. Velten is proof there isn't.
[Joe Topjian](https://github.com/jtopjian) | Joe insisted that we omit this pithy attempt at showing gratitude, but we simply could not do that. Not when Mr. Velten insisted it would be in our best interest to include Joe. Is this okay Mr. Velten? Can our families come home now? We did what you asked. Joe is awesome. We like Joe. See? We're cooperating. Please Mr. Velten, just let them come home!


## Version 0.5.2 (2017/03/28)
This is a minor release with some bug fixes, enhancements, and simplified
support for TLS.

### New Features
* TLS Support ([#447](https://github.com/codedellemc/libstorage/issues/447))

### Bug Fixes
* Handle varying `rbd` output format ([#451](https://github.com/codedellemc/libstorage/issues/451))
* Fix ScaleIO missing `/dev/disk/by-id` ([#466](https://github.com/codedellemc/libstorage/issues/466))
* Fix Linux integration driver's encryption omission ([#481](https://github.com/codedellemc/libstorage/issues/481))
* Document `Volume.AttachmentState` ([#483](https://github.com/codedellemc/libstorage/issues/483))

### Enhancements
* Embedded API documentation ([#487](https://github.com/codedellemc/libstorage/issues/487))
* Update organization text ([#472](https://github.com/codedellemc/libstorage/issues/472))

## Version 0.5.1 (2017/02/24)
This is a minor release to ensure Go1.6 compatibility.

### Bug Fixes
* FittedCloud Go1.6 support ([#444](https://github.com/codedellemc/libstorage/pull/444))

## Version 0.5.0 (2017/02/24)
This is one of the largest releases in a while, including support for new
storage platforms, client enhancements, and more!

### New Features
* Amazon Simple Storage Service FUSE (S3FS) support ([#397](https://github.com/codedellemc/libstorage/issues/397), [#409](https://github.com/codedellemc/libstorage/issues/409))
* Google Compute Engine Persistent Disk (GCEPD) support ([#394](https://github.com/codedellemc/libstorage/issues/394), [#416](https://github.com/codedellemc/libstorage/issues/416))
* DigitalOcean support ([#392](https://github.com/codedellemc/libstorage/issues/392))
* Microsoft Azure unmanaged disk support ([#421](https://github.com/codedellemc/libstorage/issues/421))
* FittedCloud support ([#408](https://github.com/codedellemc/libstorage/issues/408))
* Storage-platform specific mount/unmount support ([#399](https://github.com/codedellemc/libstorage/issues/399))
* The ScaleIO tool `drv_cfg` is now an optional client-side dependency instead of required ([#414](https://github.com/codedellemc/libstorage/issues/414))
* Multi-cluster support for ScaleIO ([#420](https://github.com/codedellemc/libstorage/issues/420))

### Bug Fixes
* Preemption fix ([#413](https://github.com/codedellemc/libstorage/issues/413))
* Ceph RBD monitored IP fix ([#412](https://github.com/codedellemc/libstorage/issues/412), [#424](https://github.com/codedellemc/libstorage/issues/424))
* Ceph RBD dashes in names fix ([#425](https://github.com/codedellemc/libstorage/issues/425))
* Fix for `lsx-OS wait` argument count ([#401](https://github.com/codedellemc/libstorage/issues/401))
* Build fixes ([#403](https://github.com/codedellemc/libstorage/issues/403))

### Thank You
  Name | Blame  
-------|------
[Chris Duchesne](https://github.com/cduchesne) | Chris is my partner in crime when it comes to libStorage and REX-Ray. Without him I would have absolutely no one to take the fall for the heist I'm planning. So is Chris invaluable? Yeah, in that way, as the patsy who will do at least a dime while I'm on the beach sipping my drink, yeah, he's invaluable.
[Travis Rhoden](https://github.com/codenrhoden) | Travis, or as I call him, T-Dawg, is essential to "taking care of business." He comes to work to chew bubblegum and kick butt, and he leaves the gum at home!
[Dan Norris](https://github.com/protochron) | Dan "The Man" Norris is well known in the underground street-swimming circuit. Last year he tied Michael Phelps in the Santa Monica Sewer 120 meter medley. He would have won if not for stopping to create the DigitalOcean driver for libStorage.
[Alexey Morlang](https://github.com/alexey-mr) | As a third-chair oboe player in the Moscow orchestra it is surprising that Alexey still finds time to contribute to the project, but coming from a long line of oboligarchs (oboe playing oligarchs), it's just in his nature. As is creating storage drivers. That, and, well, playing the oboe.
[Andrey Pavlov](https://github.com/Andrey-mp) | There is no Andrey. You have not met him. He does not exist. Don't look behind you. He is not there. He is writing storage drivers. Then just like that, he's vanished.
[Lax Kota](https://github.com/Lax77) | Lax is a rock star in the Slack channel, helping others by answering their questions before the project's developers can take a stab. We do not want to upset him. It's rumored he beats those who upset him in order to provide inspiration for his true passion -- corporal poetry. Every punch thrown is another verse towards his masterpiece.
[Jack Huang](https://github.com/jack-fittedcloud) | Jack is not his job. Jack is not the amount of money he has in the bank. Jack is not the car he drives. Jack is not the clothes he wears. Jack is a supernova, accelerating at the speed of light beyond the bounds of quantifiable space and time. Jack is not the stuff above. Jack is not the stuff below. Jack is not the stuff in between. Jack is not the empty void. Jack. just. is.


## Version 0.4.0 (2017/01/20)
Another exciting new feature release, this update brings with it support for
the Ceph RBD platform.

### New Features
* Ceph RBD Support ([#347](https://github.com/codedellemc/libstorage/issues/347), [#367](https://github.com/codedellemc/libstorage/issues/367))

### Bug Fixes
* Fix Linux integration driver preemption ([#391](https://github.com/codedellemc/libstorage/issues/391))

## Version 0.3.8 (2017/01/05)
This is a minor bugfix release that includes a fix for volume filtering.

### Bug Fixes
* Fix for attachment filtering on unavailable volumes ([#383](https://github.com/codedellemc/libstorage/pull/383))

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
