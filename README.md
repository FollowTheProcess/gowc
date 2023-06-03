# gowc

[![License](https://img.shields.io/github/license/FollowTheProcess/gowc)](https://github.com/FollowTheProcess/gowc)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/gowc)](https://goreportcard.com/report/github.com/FollowTheProcess/gowc)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/gowc?logo=github&sort=semver)](https://github.com/FollowTheProcess/gowc)
[![CI](https://github.com/FollowTheProcess/gowc/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/gowc/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/FollowTheProcess/gowc/branch/main/graph/badge.svg)](https://codecov.io/gh/FollowTheProcess/gowc)

Toy clone of [coreutils] [wc] in Go

## Project Description

`gowc` is a toy reimplementation of [wc] in Go, mainly written for fun 😃. It's perfectly functional, well tested and correct but there's no real
benefit over using it vs the original (aside from maybe the JSON flag).

The main reason I chose to write it was that I discovered you can (sort of) abuse the [io.Writer] interface to count lines, words etc. The primary benefit being
you can then leverage [io.Copy] from either files or stdin (both of which implement [io.Reader]).

Using [io.Copy] means large files automatically get chunked into 32kb blocks and streamed through your program so `gowc` works seamlessly on enormous files!

So this was a fun experiment to see how far you can take it.

## Installation

Compiled binaries for all supported platforms can be found in the [GitHub release]. There is also a [homebrew] tap:

```shell
brew install FollowTheProcess/homebrew-tap/gowc
```

## Quickstart

### Pipe from stdin

```shell
gowc < moby_dick.txt

# Or
cat moby_dick.txt | gowc
```

```plain
File          Bytes   Chars   Lines Words
moby_dick.txt 1232922 1232922 23243 214132
```

### Read from file

```shell
gowc moby_dick.txt
```

```plain
File          Bytes   Chars   Lines Words
moby_dick.txt 1232922 1232922 23243 214132
```

### Multiple files

Multiple files are counted concurrently using a worker pool 🚀

```shell
gowc myfiles/*
```

```plain
File                   Bytes    Chars   Lines Words
.myfiles/onemore.txt   460      460     2     63
.myfiles/another.txt   608      608     2     80
.myfiles/moby_dick.txt 1232922  1232922 23243 214132
```

### JSON

```shell
gowc -json moby_dick.txt | jq
```

```json
{
  "name": "moby_dick.txt",
  "lines": 23243,
  "bytes": 1232922,
  "words": 214132,
  "chars": 1232922
}
```

You can also do multiple files in JSON:

```shell
gowc -json myfiles/*
```

```json
[
  {
    "name": "myfiles/onemore.txt",
    "lines": 2,
    "bytes": 460,
    "words": 63,
    "chars": 460
  },
  {
    "name": "myfiles/another.txt",
    "lines": 2,
    "bytes": 608,
    "words": 80,
    "chars": 608
  },
  {
    "name": "myfiles/moby_dick.txt",
    "lines": 23243,
    "bytes": 1232922,
    "words": 214132,
    "chars": 1232922
  }
]
```

### Credits

This package was created with [copier] and the [FollowTheProcess/go_copier] project template.

[copier]: https://copier.readthedocs.io/en/stable/
[FollowTheProcess/go_copier]: https://github.com/FollowTheProcess/go_copier
[GitHub release]: https://github.com/FollowTheProcess/gowc/releases
[homebrew]: https://brew.sh
[coreutils]: https://www.gnu.org/software/coreutils/manual/
[wc]: https://www.gnu.org/software/coreutils/manual/html_node/wc-invocation.html#wc-invocation
[io.Writer]: https://pkg.go.dev/io#Writer
[io.Copy]: https://pkg.go.dev/io#Copy
[io.Reader]: https://pkg.go.dev/io#Reader
