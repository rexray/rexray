# JSON Schema validator

[![GoDoc](https://godoc.org/github.com/cesanta/validate-json/schema?status.svg)](https://godoc.org/github.com/cesanta/validate-json/schema)

This binary is a command-line wrapper for a library that implements [JSON Schema
draft 04 specification](http://json-schema.org/documentation.html).
It passes all the tests from https://github.com/json-schema/JSON-Schema-Test-Suite
except for optional/bignum.json, but it doesn't mean that it's free of bugs,
especially in scope alteration and resolution, since that part is not entrirely
clear.

## Contributions

People who have agreed to the
[Cesanta CLA](http://cesanta.com/contributors_la.html)
can make contributions. Note that the CLA isn't a copyright
_assigment_ but rather a copyright _license_.
You retain the copyright on your contributions.

## Licensing

This software is released under commercial and
[GNU GPL v.2](http://www.gnu.org/licenses/old-licenses/gpl-2.0.html) open
source licenses. The GPLv2 open source License does not generally permit
incorporating this software into non-open source programs.
For those customers who do not wish to comply with the GPLv2 open
source license requirements,
[Cesanta](http://cesanta.com) offers a full,
royalty-free commercial license and professional support
without any of the GPL restrictions.
