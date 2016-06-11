# Release Notes

Release early, release often

---

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
[Chris Duchesne](https://github.com/cduchesne) | His short time with the time is in complete opposition to the value he has added to this project
[David vonThenen](https://github.com/dvonthenen) | David has been a go-to guy for debugging the most difficult of issues
[Steve Wong](https://github.com/cantbewong) | Steve stays on top of the things and keeps use cases in sync with industry needs
[Travis Rhoden](https://github.com/codenrhoden) | Another keen mind, Travis is also a great font of technical know-how
[Peter Blum](https://github.com/oskoss) | Absent Peter, the EMC World demo would not have been ready
[Megan Hyland](https://github.com/meganmurawski) | And absent Megan, Peter's work would only have taken things halfway there
[Eugene Chupriyanov](https://github.com/echupriyanov) | For helping with the EC2 planning
[Matt Farina](https://github.com/mattfarina) | Without Glide, it all comes crashing down
Josh Bernstein | The shadowy figure behind the curtain...

And many more...
