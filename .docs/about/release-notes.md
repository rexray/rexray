# Release Notes

---

## Upgrading

To upgrade REX-Ray to the latest version, use `curl install`:

    curl -sSL https://dl.bintray.com/emccode/rexray/install | sh

Use `rexray version` to determine the currently installed version of REX-Ray:

```bash
$ rexray version
REX-Ray
-------
Binary: /usr/bin/rexray
Flavor: client+agent+controller
SemVer: 0.10.0-rc2
OsArch: Linux-x86_64
Commit: 5b1c7431012f28f72d36d6788e204d7e78811168
Formed: Thu, 07 Sep 2017 17:49:48 CDT
```

## Version 0.10.0 (2017/09/11)
Hi, REX here. Look, I know it's been a while since the last release. Three
whole months. I'm not going to apologize either. I'm an anthropomorphic dog
who chases bugs instead of bones for a living. That's pretty impressive in
and of itself!

However, it's been worth the wait. This is a major release, with changes
that include:

### New Features
* Support for the Container Storage Interface ([CSI](https://github.com/container-storage-interface/spec)) specification ([`8dbb732`](https://github.com/container-storage-interface/spec/commit/8dbb73222cdb63ce583e5c0abeafdf96748bf4f5)) ([\#878](https://github.com/codedellemc/rexray/issues/878))
* The [CSI-BlockDevices](https://github.com/codedellemc/csi-blockdevices) plug-in provides local block device support ([\#961](https://github.com/codedellemc/rexray/issues/961))
* The [CSI-NFS](https://github.com/codedellemc/csi-nfs) plug-in provides NFS support ([\#962](https://github.com/codedellemc/rexray/issues/962))
* The [CSI-VFS](https://github.com/codedellemc/csi-vfs) plug-in provides virtual filesystem (VFS) support ([\#878](https://github.com/codedellemc/rexray/issues/878))
* Raw device support for all existing, block-based storage platforms ([\#998](https://github.com/codedellemc/rexray/issues/998))

### Enhancements
* All existing, managed Docker plug-ins now support CSI ([\#1012](https://github.com/codedellemc/rexray/issues/1012))
* S3FS supports custom endpoints ([\#986](https://github.com/codedellemc/rexray/issues/986))
* EBS supports large-device range option ([\#996](https://github.com/codedellemc/rexray/issues/996))
* Support for DigitalOcean environment variables with underscores ([\#1004](https://github.com/codedellemc/rexray/issues/1004))
* Use Golang dependency management tool [`dep`](https://github.com/golang/dep) ([\#993](https://github.com/codedellemc/rexray/issues/993))
* Merged libStorage into REX-Ray ([\#932](https://github.com/codedellemc/rexray/issues/932))
* Simplified build process ([\#937](https://github.com/codedellemc/rexray/issues/937))
* Versioned `vendor` directory ([\#930](https://github.com/codedellemc/rexray/issues/930))
* REX-Ray now built with Go 1.8.3 ([\#918](https://github.com/codedellemc/rexray/issues/918))

### Bug Fixes
* Fixed required permissions for S3FS Docker plug-in ([\#893](https://github.com/codedellemc/rexray/issues/893))
* Fixed Cinder support ([\#935](https://github.com/codedellemc/rexray/issues/935))
* Fixed volume mount path retrieval in FlexREX `DetachVolume` ([\#914](https://github.com/codedellemc/rexray/issues/914))
* Sirupsen to sirupsen ([\#995](https://github.com/codedellemc/rexray/issues/995))

## Version 0.9.2 (2017/06/28)
This is a minor release that introduces a Docker managed plug-in for Ceph RBD,
and fixes a regression with the S3FS Docker plug-in.

### New Features
* Ceph RBD Docker managed plug-in ([#898](https://github.com/codedellemc/rexray/pull/898))

### Bug Fixes
* Fix S3FS Docker plug-in permissions issue ([#891](https://github.com/codedellemc/rexray/issues/891))

### Enhancements
* [libStorage 0.6.2](https://github.com/codedellemc/libstorage/releases/tag/v0.6.2)
* REX-Ray documentation has a new look, and is now searchable ([#889](https://github.com/codedellemc/rexray/pull/889))
* Be more verbose on CLI about embedded errors from libStorage ([#899](https://github.com/codedellemc/rexray/pull/899))

## Version 0.9.1 (2017/06/09)
This release is primarily a bug-fix release, and also introduces two new Docker
managed plug-ins: Digital Ocean (dobs) and OpenStack Cinder (cinder). This
release includes libStorage 0.6.1

### New Features
* OpenStack Cinder Docker plugin ([#853](https://github.com/codedellemc/rexray/issues/853))
* Digital Ocean Docker plugin ([#816](https://github.com/codedellemc/rexray/issues/816))

### Bug Fixes
* Fix handling of white space in Ceph config file for monitor hosts ([#811](https://github.com/codedellemc/rexray/issues/811))
* Fix volume create for Isilon storage ([#556](https://github.com/codedellemc/libstorage/issues/556))

### Enhancements
* No longer require `libstorage.host` config option when `host` is defined in a module ([#827](https://github.com/codedellemc/rexray/issues/827))
* Add additional env var options to S3FS Docker plugin ([#809](https://github.com/codedellemc/rexray/issues/809))
* Documentation updates around TLS and REX-Ray configuration ([#888](https://github.com/codedellemc/rexray/pull/888), [#831](https://github.com/codedellemc/rexray/pull/831))
* Build with Go 1.8.1 ([#844](https://github.com/codedellemc/rexray/pull/844))
* ARM build support ([#865](https://github.com/codedellemc/rexray/pull/865))

### Thank You
  Name | Blame  
-------|------
[Sébastien GLON](https://github.com/sebglon) | Lord GLON is heir to the Dukes of Brittany and currently resides Château des Ducs de Bretagne along the Loire River. One can only imagine the vistas to which Lord GLON is privilege as he gazes out the arrow-loops carved into the stone exterior of his office wall. When not creating pull requests for REX-Ray Lord GLON is rumored to be carrying on a torrid affair with Mayor Rolland, the first female, and certainly most stunning, to hold the office in Nantes. We submit the most humble of gratitudes to Johanna for the pittance of time she allowed Lord GLON to separate from her and attend to the Cinder plug-in. Many thanks to you both.

## Version 0.9.0 (2017/05/03)
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

## Version 0.8.2 (2017/03/28)
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
* Update organization text ([#774](https://github.com/codedellemc/rexray/issues/774))

## Version 0.8.1 (2017/02/24)
This is a minor release that reintroduces support for Go1.6 via
libStorage 0.5.1.

### Bug Fixes
* Go1.6 support ([#444](https://github.com/codedellemc/libstorage/issues/444))

## Version 0.8.0 (2017/02/24)
This is one of the largest releases in a while, including support for five new
storage platforms!

### New Features
* Amazon Simple Storage Service FUSE (S3FS) support ([#397](https://github.com/codedellemc/libstorage/issues/397), [#409](https://github.com/codedellemc/libstorage/issues/409))
* Google Compute Engine Persistent Disk (GCEPD) support ([#394](https://github.com/codedellemc/libstorage/issues/394), [#416](https://github.com/codedellemc/libstorage/issues/416))
* DigitalOcean support ([#392](https://github.com/codedellemc/libstorage/issues/392))
* Microsoft Azure unmanaged disk support ([#421](https://github.com/codedellemc/libstorage/issues/421))
* FittedCloud support ([#408](https://github.com/codedellemc/libstorage/issues/408))
* Docker Volume Plug-in for EBS ([#720](https://github.com/codedellemc/rexray/issues/720))
* Docker Volume Plug-in for EFS ([#729](https://github.com/codedellemc/rexray/issues/729))
* Docker Volume Plug-in for Isilon ([#727](https://github.com/codedellemc/rexray/issues/727))
* Docker Volume Plug-in for S3FS ([#724](https://github.com/codedellemc/rexray/issues/724))
* Docker Volume Plug-in for ScaleIO ([#725](https://github.com/codedellemc/rexray/issues/725))
* REX-Ray on Alpine Linux support ([#724](https://github.com/codedellemc/rexray/issues/724))
* Storage-platform specific mount/unmount support ([#399](https://github.com/codedellemc/libstorage/issues/399))
* The ScaleIO tool `drv_cfg` is now an optional client-side dependency instead of required ([#414](https://github.com/codedellemc/libstorage/issues/414))
* Multi-cluster support for ScaleIO ([#420](https://github.com/codedellemc/libstorage/issues/420))
* Forced volume remove support ([#717](https://github.com/codedellemc/rexray/issues/717))

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
[Vladimir Vivien](https://github.com/vladimirvivien) | A little known fact about Vladimir is that he's been seeded in the top 10 of the last US Opens, but has had to withdrawal at the last minute before each of those tournaments due to other responsibilities. What those are? Who can say? Are they contracts on people's lives? Perhaps. Are they appearances for Make a Wish? Probably. The only thing we know for sure is that when he is seen again, Vladimir seems rejuvenated and ready to conquer the tennis world yet again.
[Steve Wong](https://github.com/cantbewong) | I've known Steve for a very long time, and in that time I can say I've never once seen him in the same room as President Barack Obama. Now, does that mean that I can definitively state that Steve and President Obama are in fact the same person. No, of course not. There are obvious differences. The most glaring of course being that Steve wears glasses and President Obama does not. However, other than that the two men are nearly identical. I guess we'll never know if Steve Wong lives a double life as the 44th President of these United States, but I personally would like to think that yeah, he does.
[Dan Norris](https://github.com/protochron) | Dan "The Man" Norris is well known in the underground street-swimming circuit. Last year he tied Michael Phelps in the Santa Monica Sewer 120 meter medley. He would have won if not for stopping to create the DigitalOcean driver for libStorage.
[Alexey Morlang](https://github.com/alexey-mr) | As a third-chair oboe player in the Moscow orchestra it is surprising that Alexey still finds time to contribute to the project, but coming from a long line of oboligarchs (oboe playing oligarchs), it's just in his nature. As is creating storage drivers. That, and, well, playing the oboe.
[Andrey Pavlov](https://github.com/Andrey-mp) | There is no Andrey. You have not met him. He does not exist. Don't look behind you. He is not there. He is writing storage drivers. Then just like that, he's vanished.
[Lax Kota](https://github.com/Lax77) | Lax is a rock star in the Slack channel, helping others by answering their questions before the project's developers can take a stab. We do not want to upset him. It's rumored he beats those who upset him in order to provide inspiration for his true passion -- corporal poetry. Every punch thrown is another verse towards his masterpiece.
[Jack Huang](https://github.com/jack-fittedcloud) | Jack is not his job. Jack is not the amount of money he has in the bank. Jack is not the car he drives. Jack is not the clothes he wears. Jack is a supernova, accelerating at the speed of light beyond the bounds of quantifiable space and time. Jack is not the stuff above. Jack is not the stuff below. Jack is not the stuff in between. Jack is not the empty void. Jack. just. is.


## Version 0.7.0 (2017/01/23)
This feature release includes support for libStorage 0.4.0 and the Ceph RBD
storage platform.

### Enhancements
* [libStorage 0.4.0](https://github.com/codedellemc/libstorage/releases/tag/v0.4.0)
* Ceph/RBD storage platform ([#347](https://github.com/codedellemc/libstorage/pull/347))

### Bug Fixes
* Prevent unnecessary removal of directory by FlexREX ([#699](https://github.com/codedellemc/rexray/pull/699))
* Update `volume attach` to check for `--force` flag ([#696](https://github.com/codedellemc/rexray/pull/696))
* Fix installer to correctly parse new Bintray HTML ([#687](https://github.com/codedellemc/rexray/pull/687))

## Version 0.6.4 (2017/01/05)
This release includes the new script manager and FlexVol REX-Ray plug-in.

### Enhancements
* [libStorage 0.3.8](https://github.com/codedellemc/libstorage/releases/tag/v0.3.8)
* Script manager ([#669](https://github.com/codedellemc/rexray/pull/669))
* FlexVol plug-in for Kubernetes ([#641](https://github.com/codedellemc/rexray/pull/641))

### Bug Fixes
* Panic on `$ rexray-client volume mount` ([#673](https://github.com/codedellemc/rexray/pull/673))

## Version 0.6.3 (2016/12/07)
This release includes the ability to specify a custom encryption key when
creating volumes and makes the `volume attach` command idempotent.

### Enhancements
* [libStorage 0.3.5](https://github.com/codedellemc/libstorage/releases/tag/v0.3.5)
* Support for creating encrypted volumes ([#649](https://github.com/codedellemc/rexray/pull/649), [#652](https://github.com/codedellemc/rexray/pull/652))
* Idempotent volume attach command ([#651](https://github.com/codedellemc/rexray/pull/651))

### Bug Fixes
* Fix volume status for detach op ([#654](https://github.com/codedellemc/rexray/pull/654))

## Version 0.6.2 (2016/12/05)
While a patch release, this new version includes some much-requested features
and updates.

### Enhancements
* [libStorage 0.3.4](https://github.com/codedellemc/libstorage/pull/351)
* Auto-detect running service ([#642](https://github.com/codedellemc/rexray/pull/642))
* Prettier error messages ([#645](https://github.com/codedellemc/rexray/pull/645))

### Bug Fixes
* Graceful exit with SystemD ([#644](https://github.com/codedellemc/rexray/pull/644))

## Version 0.6.1 (2016/12/01)
This release includes some minor fixes as well as a new and improved version of
the `volume ls` command.

### Enhancements
* [libStorage 0.3.3](https://github.com/codedellemc/libstorage/pull/348)
* Enhanced `volume ls` command ([#634](https://github.com/codedellemc/rexray/pull/634))

### Bug Fixes
* EFS Mounting Issues ([#609](https://github.com/codedellemc/rexray/pull/609))
* VirtualBox Attach Issues ([#610](https://github.com/codedellemc/rexray/pull/610))
* Installer upgrade fix ([#637](https://github.com/codedellemc/rexray/pull/637))
* Build deployment fix ([#638](https://github.com/codedellemc/rexray/pull/638))

## Version 0.6.0 (2016/10/20)
This release reintroduces the Elastic Block Storage (EBS) driver, formerly known
as the EC2 driver. All vestigial EC2 configuration properties are still
supported.

### Enhancements
* libStorage 0.3.0 ([#docs](http://libstorage.readthedocs.io/en/v0.3.0))
* Amazon Elastic Block Storage (EBS) Support ([#522](https://github.com/codedellemc/rexray/issues/522))
* New CLI Output ([#579](https://github.com/codedellemc/rexray/issues/579), [#603](https://github.com/codedellemc/rexray/issues/603), [#606](https://github.com/codedellemc/rexray/issues/606))
* Support for ScaleIO 2.0.1 ([#599](https://github.com/codedellemc/rexray/issues/599))

### Bug Fixes
* Handle phantom mounts for EBS (formerly EC2) ([#410](https://github.com/codedellemc/rexray/issues/410))

## Version 0.5.1 (2016/09/14)
This is a minor release, but includes a few important patches.

### Enhancements
* libStorage 0.2.1 ([#docs](http://libstorage.readthedocs.io/en/v0.2.1))
* ScaleIO 2.0.0.2 Support ([#555](https://github.com/codedellemc/rexray/issues/555))

### Bug Fixes
* EFS Volume / Tag Creation Bug ([#261](https://github.com/codedellemc/libstorage/issues/261))

## Version 0.5.0 (2016/09/07)
Beginning with this release, REX-Ray's versions will increment the MINOR
component with the introduction of a new storage driver via libStorage in
concert with the [guidelines](http://semver.org) set forth by semantic
versioning.

### New Features
* Amazon Elastic File System (EFS) Support ([#525](https://github.com/codedellemc/rexray/issues/525))

### Enhancements
* Support for Go 1.7 ([#541](https://github.com/codedellemc/rexray/issues/541))
* Enhanced Isilon Support ([#520](https://github.com/codedellemc/rexray/issues/520), [#521](https://github.com/codedellemc/rexray/issues/521))

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
* libStorage 0.1.5 ([#TBA](https://github.com/codedellemc/rexray/issues/TBA))
* Improved volume path caching ([#500](https://github.com/codedellemc/rexray/issues/500))

## Version 0.4.1 (2016/07/08)
Although a minor release, 0.4.1 provides some meaningful and useful enhancements
and fixes, further strengthening the foundation of the REX-Ray platform.

### Enhancements
* Improved build process ([#474](https://github.com/codedellemc/rexray/issues/474), [#492](https://github.com/codedellemc/rexray/issues/492))
* [libStorage](http://libstorage.readthedocs.io) 0.1.4 ([#493](https://github.com/codedellemc/rexray/issues/493))
* Removed Docker spec file ([#486](https://github.com/codedellemc/rexray/issues/486))
* Improved REX-Ray 0.3.3 Config Backwards Compatibility ([#481](https://github.com/codedellemc/rexray/issues/481))
* Improved install script ([#439](https://github.com/codedellemc/rexray/issues/439), [#495](https://github.com/codedellemc/rexray/issues/495))

### Bug Fixes
* Fixed input validation bug when creating volume sans name ([#478](https://github.com/codedellemc/rexray/issues/478))

## Version 0.4.0 (2016/06/20)
REX-Ray 0.4.0 introduces centralized configuration and control along with
a new client/server architecture -- features made possible by
[libStorage](http://libstorage.readthedocs.io). Users are no longer
required to configure storage drivers or store privileged information on all
systems running the REX-Ray client. The new client delegates storage-platform
related operations to a remote, libStorage-compatible server such as REX-Ray
or [Poly](https://github.com/codedellemc/polly).

Please note that the initial release of REX-Ray 0.4 includes support for only
the following storage platforms:

* ScaleIO
* VirtualBox

Support for the full compliment of drivers present in earlier versions of
REX-Ray will be reintroduced over the course of several, incremental updates,
beginning with 0.4.1.

### New Features
* Distributed architecture ([#399](https://github.com/codedellemc/rexray/issues/399), [#401](https://github.com/codedellemc/rexray/issues/401), [#411](https://github.com/codedellemc/rexray/issues/411), [#417](https://github.com/codedellemc/rexray/issues/417), [#418](https://github.com/codedellemc/rexray/issues/418), [#419](https://github.com/codedellemc/rexray/issues/419), [#420](https://github.com/codedellemc/rexray/issues/420), [#423](https://github.com/codedellemc/rexray/issues/423))
* Volume locking mechanism ([#171](https://github.com/codedellemc/rexray/issues/171))
* Volume creation with initial data ([#169](https://github.com/codedellemc/rexray/issues/169))

### Enhancements
* Improved storage driver logging ([#396](https://github.com/codedellemc/rexray/issues/396))
* Docker mount path ([#403](https://github.com/codedellemc/rexray/issues/403))

### Bug Fixes
* Fixed issue with install script ([#409](https://github.com/codedellemc/rexray/issues/409))
* Fixed volume ls filter ([#400](https://github.com/codedellemc/rexray/issues/400))
* Fixed panic during access attempt of offline REX-Ray daemon ([#148](https://github.com/codedellemc/rexray/issues/148))

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
* ScaleIO v2 support ([#355](https://github.com/codedellemc/rexray/issues/355))
* EC2 Tags added to Volumes & Snapshots ([#314](https://github.com/codedellemc/rexray/issues/314))

### Enhancements
* Use of official Amazon EC2 SDK ([#359](https://github.com/codedellemc/rexray/issues/359))
* Added a disable feature for create/remove volume ([#366](https://github.com/codedellemc/rexray/issues/366))
* Added ScaleIO troubleshooting information ([#367](https://github.com/codedellemc/rexray/issues/367))

### Bug Fixes
* Fixes URLs for documentation when viewed via Github ([#337](https://github.com/codedellemc/rexray/issues/337))
* Fixes logging bug on Ubuntu 14.04 ([#377](https://github.com/codedellemc/rexray/issues/377))
* Fixes module start timeout error ([#376](https://github.com/codedellemc/rexray/issues/376))
* Fixes ScaleIO authentication loop bug ([#375](https://github.com/codedellemc/rexray/issues/375))

### Thank You
* [Philipp Franke](https://github.com/philippfranke)
* [Eugene Chupriyanov](https://github.com/echupriyanov)
* [Peter Blum](https://github.com/oskoss)
* [Megan Hyland](https://github.com/meganmurawski)

## Version 0.3.2 (2016-03-04)

### New Features
* Support for Docker 1.10 and Volume Plugin Interface 1.2 ([#273](https://github.com/codedellemc/rexray/issues/273))
* Stale PID File Prevents Service Start ([#258](https://github.com/codedellemc/rexray/issues/258))
* Module/Personality Support ([#275](https://github.com/codedellemc/rexray/issues/275))
* Isilon Preemption ([#231](https://github.com/codedellemc/rexray/issues/231))
* Isilon Snapshots ([#260](https://github.com/codedellemc/rexray/issues/260))
* boot2Docker Support ([#263](https://github.com/codedellemc/rexray/issues/263))
* ScaleIO Dynamic Storage Pool Support ([#267](https://github.com/codedellemc/rexray/issues/267))

### Enhancements
* Improved installation documentation ([#331](https://github.com/codedellemc/rexray/issues/331))
* ScaleIO volume name limitation ([#304](https://github.com/codedellemc/rexray/issues/304))
* Docker cache volumes for path operations ([#306](https://github.com/codedellemc/rexray/issues/306))
* Config file validation ([#312](https://github.com/codedellemc/rexray/pull/312))
* Better logging ([#296](https://github.com/codedellemc/rexray/pull/296))
* Documentation Updates ([#285](https://github.com/codedellemc/rexray/issues/285))

### Bug Fixes
* Fixes issue with daemon process getting cleaned as part of SystemD Cgroup ([#327](https://github.com/codedellemc/rexray/issues/327))
* Fixes regression in 0.3.2 RC3/RC4 resulting in no log file ([#319](https://github.com/codedellemc/rexray/issues/319))
* Fixes no volumes returned on empty list ([#322](https://github.com/codedellemc/rexray/issues/322))
* Fixes "Unsupported FS" when mounting/unmounting with EC2 ([#321](https://github.com/codedellemc/rexray/issues/321))
* ScaleIO re-authentication issue ([#303](https://github.com/codedellemc/rexray/issues/303))
* Docker XtremIO create volume issue ([#307](https://github.com/codedellemc/rexray/issues/307))
* Service status is reported correctly ([#310](https://github.com/codedellemc/rexray/pull/310))

### Updates
* <del>Go 1.6 ([#308](https://github.com/codedellemc/rexray/pull/308))</del>

### Thank You
* Dan Forrest
* Kapil Jain
* Alex Kamalov


## Version 0.3.1 (2015-12-30)

### New Features
* Support for VirtualBox ([#209](https://github.com/codedellemc/rexray/issues/209))
* Added Developer's Guide ([#226](https://github.com/codedellemc/rexray/issues/226))

### Enhancements
* Mount/Unmount Accounting ([#212](https://github.com/codedellemc/rexray/issues/212))
* Support for Sub-Path Volume Mounts / Permissions ([#215](https://github.com/codedellemc/rexray/issues/215))

### Milestone Issues
This release also includes many other small enhancements and bug fixes. For a
complete list click [here](https://github.com/codedellemc/rexray/pulls?q=is%3Apr+is%3Aclosed+milestone%3A0.3.1).

### Downloads
Click [here](https://dl.bintray.com/emccode/rexray/stable/0.3.1/) for the 0.3.1
binaries.

## Version 0.3.0 (2015-12-08)

### New Features
* Pre-Emption support ([#190](https://github.com/codedellemc/rexray/issues/190))
* Support for VMAX ([#197](https://github.com/codedellemc/rexray/issues/197))
* Support for Isilon ([#198](https://github.com/codedellemc/rexray/issues/198))
* Support for Google Compute Engine (GCE) ([#194](https://github.com/codedellemc/rexray/issues/194))

### Enhancements
* Added driver example configurations ([#201](https://github.com/codedellemc/rexray/issues/201))
* New configuration file format ([#188](https://github.com/codedellemc/rexray/issues/188))

### Tweaks
* Chopped flags `--rexrayLogLevel` becomes `logLevel` ([#196](https://github.com/codedellemc/rexray/issues/196))

### Pre-Emption Support
Pre-Emption is an important feature when using persistent volumes and container
schedulers.  Without pre-emption, the default behavior of the storage drivers is
to deny the attaching operation if the volume is already mounted elsewhere.  
If it is desired that a host should be able to pre-empt from other hosts, then
this feature can be used to enable any host to pre-empt from another.

### Milestone Issues
This release also includes many other small enhancements and bug fixes. For a
complete list click [here](https://github.com/codedellemc/rexray/pulls?q=is%3Apr+is%3Aclosed+milestone%3A0.3.0).

### Downloads
Click [here](https://dl.bintray.com/emccode/rexray/stable/0.3.0/) for the 0.3.0
binaries.

## Version 0.2.1 (2015-10-27)
REX-Ray release 0.2.1 includes OpenStack support, vastly improved documentation,
and continued foundation changes for future features.

### New Features
* Support for OpenStack ([#111](https://github.com/codedellemc/rexray/issues/111))
* Create volume from volume using existing settings ([#129](https://github.com/codedellemc/rexray/issues/129))

### Enhancements
* A+ [GoReport Card](http://goreportcard.com/report/emccode/rexray)
* A+ [Code Coverage](https://coveralls.io/github/emccode/rexray?branch=master)
* [GoDoc Support](https://godoc.org/github.com/codedellemc/rexray)
* Ability to load REX-Ray as an independent storage platform ([#127](https://github.com/codedellemc/rexray/issues/127))
* New documentation at http://rexray.readthedocs.org ([#145](https://github.com/codedellemc/rexray/issues/145))
* More foundation updates

### Tweaks
* Command aliases for `get` and `delete` - `ls` and `rm` ([#107](https://github.com/codedellemc/rexray/issues/107))

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
[REX-Ray website](http://github.com/codedellemc/rexray) for information on how to
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
