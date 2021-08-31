# Release Process

How to release REX-Ray

---

## Project Stages
This project has three parallels stages of release:

Name | Description
-----|------------
`unstable` | The tip or HEAD of the `master` branch is referred to as `unstable`
`staged` | A commit tagged with the suffix `-rc\d+` such as `v0.3.1-rc2` is a `staged` release. These are release candidates.
`stable` | A commit tagged with a version sans `-rc\d+` suffix such as `v0.3.1` is a `stable` release.

There are no steps necessary to create an `unstable` release as that happens
automatically whenever an untagged commit is pushed to `master`. However, the
following workflow should be used when tagging a `staged` release candidate
or `stable` release.

  1. Review outstanding issues & pull requests
  2. Prepare release notes
  3. Update the version file
  4. Commit & pull request
  5. Tag the release

## Review Issues & Pull Requests
The first step to a release is to review the outstanding
[issues](https://github.com/AVENTER-UG/rexray/issues) and
[pull requests](https://github.com/AVENTER-UG/rexray/pulls) that are tagged for
the release in question.

If there are outstanding issues requiring changes or pending pull requests to
be merged, handle those prior to tagging any commit as a release candidate or
release.

It is __highly__ recommended that pull requests be merged synchronously after
rebasing each subsequent one off of the new tip of `master`. Remember, while
GitHub will update a pull request as in conflict if a change to `master`
results in a merge conflict with the pull request, GitHub will __not__ force a
new build to spawn unless the pull request is actually updated.

At the very minimum a pull request's build should be re-executed prior to the
pull request being merged if `master` has changed since the pull request was
opened.

## Prepare Release Notes
Update the release notes at `.docs/about/release-notes.md`. This file is the
project's authoritative changelog and should reflect new features, fixes, and
any significant changes.

The most recent, `stable` version of the release notes are always available
online at
[REX-Ray's documentation site](http://rexray.rtfd.org/en/stable/about/release-notes/).

## Update Version File
The `VERSION` file exists at the root of the project and should be updated to
reflect the value of the intended release.

For example, if creating the first release candidate for version 0.3.1, the
contents of the `VERSION` file should be a single line `0.3.1-rc1` followed by
a newline character:

```sh
$ cat VERSION
0.3.1-rc1
```

If releasing version 0.3.1 proper then the contents of the `VERSION` file
should be `0.3.1` followed by a newline character:

```sh
$ cat VERSION
0.3.1
```

## Commit & Pull Request
Once all outstanding issues and pull requests are handled, the release notes
and version are updated, it's time to create a commit.

Please make sure that the changes to the release notes and version files are
a part of the same commit. This makes identifying the aspects of a release,
staged or otherwise, far easier for future developers.

A release's commit message can either be a reflection of the release notes or
something simple. Either way the commit message should have the following
subject format and first line in its body:

```text
Release Candidate 0.3.1-rc1

This patch marks release candidate 0.3.1-rc1.
```

If the commit message is longer it should simply reflect the same information
from the release notes.

Once committed push the change to a fork and open a pull request. Even though
this commit marks a staged or official release, the pull request system is still
used to assure that the build completes successfully and there are no unforeseen
errors.

## Tag the Release
Once the pull request marking the `staged` or `stable` release has been merged
into `upstream`'s `master` it's time to tag the release.

### Tag Format
The release tag should follow a prescribed format depending upon the release
type:

Release Type | Tag Format | Example
--------|---------|---------
`staged`  | vMAJOR.MINOR.PATCH-rc[0-9] | v0.3.1-rc1
`stable`  | vMAJOR.MINOR.PATCH | v0.3.1

### Tag Methods
There are two ways to tag a release:

  1. [GitHub Releases](https://github.com/AVENTER-UG/rexray/releases/new)
  2. Command Line

### Command Line
If tagging a release via the command line be sure to fetch the latest changes
from `upstream`'s `master` and either merge them into your local copy of
`master` or reset the local copy to reflect `upstream` prior to creating
any tags.

The following combination of commands can be used to create a tag for
0.3.1 Release Candidate 1:

```sh
git fetch upstream && \
  git checkout master && \
  git reset --hard upstream/master && \
  git tag -a -m v0.3.1-rc1 v0.3.1-rc1
```

The above example combines a few operations:

  1. The first command fetches the `upstream` changes
  2. The local `master` branch is checked out
  3. The local `master` branch is hard reset to `upstream/master`
  4. An annotated tag is created on `master` for `v0.3.1-rc1`, or 0.3.1 Release
     Candidate 1, with a tag message of `v0.3.1-rc1`.

Please note that the third step will erase any changes that exist only in the
local `master` branch that do not also exist in the remote, upstream copy.
However, if the two branches are not equal this method should not be used to
create a tag anyway.

The above steps do not actually push the tag upstream. This is to allow for one
final review of all the changes before doing so since the appearance of a new,
annotated tag in the repository will cause the project's build system to
automatically kick off a build that will result in the release of a `staged` or
`stable` release. For `stable` releases the project's documentation will also be
updated.

Once positive everything looks good simply execute the following command to
push the tag to the `upstream` repository:

```sh
git push upstream v0.3.1-rc1
```
