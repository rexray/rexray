# Contributing to REX-Ray
The REX-Ray project welcomes, and depends, on contributions from developers and
users in the open source community. Contributions can be made in a number of
ways, a few examples are:

- Code patches via pull requests
- Documentation improvements
- Bug reports and patch reviews
- OS, Storage, and Volume Drivers
- A distributed server/client model with profile support

## Reporting an Issue
Please include as much detail as you can. This includes:

  * The OS type and version
  * The REX-Ray version
  * The storage system in question
  * A set of logs with debug-logging enabled that show the problem

## Testing the Development Version
If you want to just install and try out the latest development version of
REX-Ray you can do so with the following command. This can be useful if you
want to provide feedback for a new feature or want to confirm if a bug you
have encountered is fixed in the git master. It is **strongly** recommended
that you do this within a virtual environment.

```bash
go get github.com/AVENTER-UG/rexray
```

## Installing for Development
First you'll need to fork and clone the repository. Once you have a local
copy, run the following command.

```bash
go get github.com/AVENTER-UG/rexray
```

This will install REX-Ray into your `GOPATH` and you'll be able to make changes
locally, test them, and commit ideas and fixes back to your fork of the
repository.

## Running the tests
To run the tests, run the following commands:

```bash
make test
```

## Submitting Pull Requests
Once you are happy with your changes or you are ready for some feedback, push
it to your fork and send a pull request. For a change to be accepted it will
most likely need to have tests and documentation if it is a new feature.
