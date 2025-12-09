# checkmake

[![Build Status](https://github.com/checkmake/checkmake/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/checkmake/checkmake/actions)
[![Coverage Status](https://coveralls.io/repos/github/mrtazz/checkmake/badge.svg?branch=master)](https://coveralls.io/github/mrtazz/checkmake?branch=master)
[![Code Climate](https://codeclimate.com/github/mrtazz/checkmake/badges/gpa.svg)](https://codeclimate.com/github/mrtazz/checkmake)
[![Go Report Card](https://goreportcard.com/badge/github.com/checkmake/checkmake)](https://goreportcard.com/report/github.com/checkmake/checkmake)
[![Packagecloud](https://img.shields.io/badge/packagecloud-available-brightgreen.svg)](https://packagecloud.io/mrtazz/checkmake)
[![MIT license](https://img.shields.io/badge/license-MIT-blue.svg)](http://opensource.org/licenses/MIT)

## Overview
**checkmake** is a linter for Makefiles. It scans Makefiles for potential issues based on configurable rules.

## Usage

```console
% checkmake Makefile
% checkmake Makefile foo.mk bar.mk baz.mk
```

checkmake analyzes one or more Makefiles and reports potential issues according to configurable rules.


### Command-line options
```console
Usage:
  checkmake [flags] [makefile...]
  checkmake [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  list-rules  List registered rules

Flags:
      --config string   Configuration file to read (default "checkmake.ini")
      --debug           Enable debug mode
      --format string   Custom Go template for text output (ignored in JSON mode)
  -h, --help            help for checkmake
  -o, --output string   Output format: 'text' (default) or 'json' (mutually exclusive with --format) (default "text")
  -v, --version         version for checkmake

Use "checkmake [command] --help" for more information about a command.
```

### Example output
```console
% checkmake fixtures/missing_phony.make
      RULE                 DESCRIPTION                      FILE NAME            LINE NUMBER
  minphony        Missing required phony target    fixtures/missing_phony.make   21
                  "all"
  minphony        Missing required phony target    fixtures/missing_phony.make   21
                  "test"
  phonydeclared   Target "all" should be           fixtures/missing_phony.make   16
                  declared PHONY.
```

## Container  usage

building or running a container image can be done with docker and podman.


# building  an image


```console
docker build --build-arg BUILDER_NAME='Your Name' --build-arg BUILDER_EMAIL=your.name@example.com . -t checker
```

Alternatively, the image can be built with the make target `image-build` :

```console
$ BUILDER_NAME='Your Name' BUILDER_EMAIL='your@mail' image-build
```

By default, the image tag is constructed as `IMAGE_REGISTRY/checkmake/checkmake:IMAGE_VERSION_TAG`

The image registry defaults to `quay.io` but can be overridden by the `IMAGE_BUILD make variable.

The image version tag defaults to `latest` and can be overridden with the `IMAGE_VERSION_TAG`.

The container command used for building (docker or podman) is auto-detected with a preference for podman but can be overridden by the make variable `CONTAINER_CMD`.

# publishing an image

The locally built image can be published with a `make image-push`command corresponding to the previously described `make image-build`command or alrenatively directly using `docker push` or `podman push`  


# running checkmake in container


Then checkmake can be run in a contaner based on a locally built or pulled image with a  Makefile attached. below is an example of it assuming the Makefile is in the  current working directory:
```console
docker run -v "$PWD"/Makefile:/Makefile quay.io/checkmake/checkmake:latest 
```

## `pre-commit` usage

This repo includes a `pre-commit` hook, which you may choose to use in your own
repos. Simply add a `.pre-commit-config.yaml` to your repo's top-level directory

```yaml
repos:
-   repo: https://github.com/checkmake/checkmake.git
    # Or another commit hash or version
    rev: 0.2.2
    hooks:
    # Use this hook to let pre-commit build checkmake in its sandbox
    -   id: checkmake
    # OR Use this hook to use a pre-installed checkmake executable
    # -   id: checkmake-system
```

There are two hooks available:

- `checkmake` (Recommended)

   pre-commit will set up a Go environment from scratch to compile and run checkmake.
   See the [pre-commit `golang` plugin docs](https://pre-commit.com/#golang) for more information.

- `checkmake-system`

   pre-commit will look for `checkmake` on your `PATH`.
   This hook requires you to install `checkmake` separately, e.g. with your package manager or [a prebuilt binary release](https://github.com/checkmake/checkmake/releases).
   Only recommended if it's permissible to require all repository users install `checkmake` manually.

Then, run `pre-commit` as usual as a part of `git commit` or explicitly, for example:

```console
pre-commit run --all-files
```

### pre-commit in GitHub Actions

You may also choose to run this as a GitHub Actions workflow. To do this, add a
`.github/workflows/pre-commit.yml` workflow to your repo:

```yaml
name: pre-commit

on:
  pull_request:
    branches:
      - master
      - main
    paths:
      - '.pre-commit-config.yaml'
      - '.pre-commit-hooks.yaml'
      - 'Makefile'
      - 'makefile'
      - 'GNUmakefile'
      - '**.mk'
      - '**.make'
  push:
    paths:
      - '.pre-commit-config.yaml'
      - '.pre-commit-hooks.yaml'
      - 'Makefile'
      - 'makefile'
      - 'GNUmakefile'
      - '**.mk'
      - '**.make'

jobs:
  pre-commit:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-python@v3
    - name: Set up Go 1.17
      uses: actions/setup-go@v2
      with:
        go-version: 1.17
      id: go
    - uses: pre-commit/action@v2.0.3
```

## Installation

### With Go

With `go` 1.16 or higher:

```console
go install github.com/checkmake/checkmake/cmd/checkmake@latest
checkmake Makefile
```

Or alternatively, run it directly:

```console
go run github.com/checkmake/checkmake/cmd/checkmake@latest Makefile
```

### From Packages
checkmake is available in many Linux distributions and package managers. See [Repology](https://repology.org/project/checkmake/versions) for full list:

[![Repology](https://repology.org/badge/vertical-allrepos/checkmake.svg?exclude_unsupported=1)](https://repology.org/project/checkmake/versions)

Packages are also available [on packagecloud.io](https://packagecloud.io/mrtazz/checkmake).

### Build
You'll need [Go](https://golang.org/) installed.

```console
git clone https://github.com/checkmake/checkmake
cd checkmake
make checkmake
```

To build the man page (optional), install [pandoc](https://pandoc.org/installing.html) and run:

```console
make checkmake.1
```

## Use in CI

### MegaLinter

checkmake is [natively embedded](https://oxsecurity.github.io/megalinter/latest/descriptors/makefile_checkmake/) within [MegaLinter](https://github.com/oxsecurity/megalinter)

To install it, run `npx mega-linter-runner --install` (requires Node.js)

## Inspiration
This is totally inspired by an idea by [Dan
Buch](https://web.archive.org/web/20200916193234/https://twitter.com/meatballhat/status/768112351924985856).
