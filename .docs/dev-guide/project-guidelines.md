# Project Guidelines

These are important.

---

People contributing code to this project must adhere to the following rules.
These standards are in place to keep code clean, consistent, and stable.

## Documentation
There are two types of documentation: source and markdown.

### Source Code
All source code should be documented in accordance with the
[Go's documentation rules](http://blog.golang.org/godoc-documenting-go-code).

### Markdown
When creating or modifying the project's `README.md` file or any of the
documentation in the `.docs` directory, please keep the following rules in
mind:

1. All markdown should be limited to a width of 80 characters. This makes
the document easier to read in text editors. GitHub and ReadTheDocs still
produces the proper result when parsing the markdown.
2. All links to internal resources should be relative.
3. All links to markdown files should include the file extension.

For example, the below link points to the anchor `basic-configuration` on the
`Configuration` page:

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
[/user-guide/config#basic-configuration](/user-guide/config#basic-configuration)

However, when the above link is followed when viewing this page directly from
the Github repository instead of the generated site documentation, the link
will return a 404.

While it's recommended that users view the generated site documentation instead
of the source Markdown directly, we can still fix it so that the above link
will work regardless. To fix the link, simply make it relative and add the
Markdown file extension:

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
[../user-guide/config.md#basic-configuration](../user-guide/config.md#basic-configuration)

Now the link will work regardless from where it's viewed.

## Style & Syntax
All source files should be processed by the following tools prior to being
committed. Any errors or warnings produced by the tools should be corrected
before the source is committed.

Tool | Description
-----|------------
[gofmt](https://golang.org/cmd/gofmt/) | A golang source formatting tool
[golint](https://github.com/golang/lint) | A golang linter
[govet](https://golang.org/cmd/vet/) | A golang source optimization tool
[gocyclo](https://github.com/fzipp/gocyclo) | A golang cyclomatic complexity detection tool. No function should have a score above 0.15

If [Atom](https://atom.io/) is your IDE of choice, install the
[go-plus](https://atom.io/packages/go-plus) package, and it will execute all of
the tools above less gocyclo upon saving a file.

In lieu of using Atom as the IDE, the project's `Makefile` automatically
executes the above tools as part of the build process and will fail the build
if problems are discovered.

Another option is to use a client-side, pre-commit hook to ensure that the
sources meet the required standards. For example, in the project's `.git/hooks`
directory create a file called `pre-commit` and mark it as executable. Then
paste the following content inside the file:

```sh
#!/bin/sh
make fmt 1> /dev/null
```

The above script will execute prior to a Git commit operation, prior to even
the commit message dialog. The script will invoke the `Makefile`'s `fmt`
target, formatting the sources. If the command returns a non-zero exit code,
the commit operation will abort with the error.

## Code Coverage
All new work submitted to the project should have associated tests where
applicable. If there is ever a question of whether or not a test is applicable
then the answer is likely yes.

This project uses
[Coveralls](https://coveralls.io/github/rexray/rexray) for code coverage, and
all pull requests are processed just as a build from `master`. If a pull request
decreases the project's code coverage, the pull request will be declined until
such time that testing is added or enhanced to compensate.

It's also possible to test the project locally while outputting the code
coverage. On the command line, from the project's root directory, execute the
following:

```sh
$ .build/test.sh
ok  	github.com/AVENTER-UG/rexray/rexray/cli	0.039s	coverage: 33.6% of statements
ok  	github.com/AVENTER-UG/rexray/test	0.080s	coverage: 94.0% of statements in github.com/AVENTER-UG/rexray, github.com/AVENTER-UG/rexray/core
...
ok  	github.com/AVENTER-UG/rexray/util	0.024s	coverage: 100.0% of statements
[0]akutz@pax:rexray$
```

The file `test.sh` in the `.build` directory is the same script executed during
the project's [automated build system](travis-ci.org/rexray/rexray). The only
difference is when executed locally the results are not submitted to Coveralls.
Still, using the `test.sh` file one can easily determine if a package's coverage
has decreased and if additional testing is necessary.

## Commit Messages
Commit messages should follow the guide [5 Useful Tips For a Better Commit
Message](https://robots.thoughtbot.com/5-useful-tips-for-a-better-commit-message).
The two primary rules to which to adhere are:

  1. Commit message subjects should not exceed 50 characters in total and
     should be followed by a blank line.

  2. The commit message's body should not have a width that exceeds 72
     characters.

For example, the following commit has a very useful message that is succinct
without losing utility.

```text
commit e80c696939a03f26cd180934ba642a729b0d2941
Author: akutz <sakutz@gmail.com>
Date:   Tue Oct 20 23:47:36 2015 -0500

    Added --format,-f option for CLI

    This patch adds the flag '--format' or '-f' for the
    following CLI commands:

        * adapter instances
        * device [get]
        * snapshot [get]
        * snapshot copy
        * snapshot create
        * volume [get]
        * volume attach
        * volume create
        * volume map
        * volume mount
        * volume path

    The user can specify either '--format=yml|yaml|json' or
    '-f yml|yaml|json' in order to influence how the resulting,
    structured data is marshaled prior to being emitted to the console.
```

Please note that the output above is the full output for viewing a commit.
However, because the above message adheres to the commit message rules, it's
quite easy to show just the commit's subject:

```sh
$ git show e80c696939a03f26cd180934ba642a729b0d2941 --format="%s" -s
Added --format,-f option for CLI
```

It's also equally simple to print the commit's subject and body together:

```sh
$ git show e80c696939a03f26cd180934ba642a729b0d2941 --format="%s%n%n%b" -s
Added --format,-f option for CLI

This patch adds the flag '--format' or '-f' for the
following CLI commands:

    * adapter instances
    * device [get]
    * snapshot [get]
    * snapshot copy
    * snapshot create
    * volume [get]
    * volume attach
    * volume create
    * volume map
    * volume mount
    * volume path

The user can specify either '--format=yml|yaml|json' or
'-f yml|yaml|json' in order to influence how the resulting,
structured data is marshaled prior to being emitted to the console.
```

## Submitting Changes
All developers are required to follow the
[GitHub Flow model](https://guides.github.com/introduction/flow/) when
proposing new features or even submitting fixes.

Please note that although not explicitly stated in the referenced GitHub Flow
model, all work should occur on a __fork__ of this project, not from within a
branch of this project itself.

Pull requests submitted to this project should adhere to the following
guidelines:

  * Branches should be rebased off of the upstream master prior to being
    opened as pull requests and again prior to merge. This is to ensure that
    the build system accounts for any changes that may only be detected during
    the build and test phase.

  * Unless granted an exception a pull request should contain only a single
    commit. This is because features and patches should be atomic -- wholly
    shippable items that are either included in a release, or not. Please
    squash commits on a branch before opening a pull request. It is not a
    deal-breaker otherwise, but please be prepared to add a comment or
    explanation as to why you feel multiple commits are required.
