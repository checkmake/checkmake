# checkmake

[![Build Status](https://github.com/mrtazz/checkmake/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/mrtazz/checkmake/actions)
[![Coverage Status](https://coveralls.io/repos/github/mrtazz/checkmake/badge.svg?branch=master)](https://coveralls.io/github/mrtazz/checkmake?branch=master)
[![Code Climate](https://codeclimate.com/github/mrtazz/checkmake/badges/gpa.svg)](https://codeclimate.com/github/mrtazz/checkmake)
[![Packagecloud](https://img.shields.io/badge/packagecloud-available-brightgreen.svg)](https://packagecloud.io/mrtazz/checkmake)
[![MIT license](https://img.shields.io/badge/license-MIT-blue.svg)](http://opensource.org/licenses/MIT)

## Overview
checkmake is a tool for linting and checking Makefiles. It may
not do what you want it to.

## Usage

```
% checkmake Makefile

% checkmake Makefile foo.mk bar.mk baz.mk

% checkmake --help
checkmake.

Usage:
checkmake [--debug|--config=<configPath>] <makefile>...
checkmake -h | --help
checkmake --version

Options:
-h --help               Show this screen.
--version               Show version.
--debug                 Enable debug mode
--config=<configPath>   Configuration file to read
--list-rules            List registered rules

% checkmake fixtures/missing_phony.make

      RULE                 DESCRIPTION             LINE NUMBER

  minphony        Missing required phony target    0
                  "all"
  minphony        Missing required phony target    0
                  "test"
  phonydeclared   Target "all" should be           18
                  declared PHONY.

```

## Docker usage
Build the image, or pull it:
```sh
docker build --build-arg BUILDER_NAME='Your Name' --build-arg BUILDER_EMAIL=your.name@example.com . -t checker
```

Then run it with your Makefile attached, below is an example of it assuming the Makefile is in your current working directory:
```sh
docker run -v "$PWD"/Makefile:/Makefile checker
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

```sh
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

### Requirements
The [pandoc](https://pandoc.org/) document converter utility is required to run checkmake. You can find out if you have it via `which pandoc`. [Install pandoc](https://pandoc.org/installing.html) if the command was not found.

## With Go

With `go` 1.16 or higher:

```sh
go install github.com/checkmake/checkmake/cmd/checkmake@latest
checkmake Makefile
```

Or alternatively, run it directly:

```sh
go run github.com/checkmake/checkmake/cmd/checkmake@latest Makefile
```

### Packages
There are packages for linux up [on packagecloud.io](https://packagecloud.io/mrtazz/checkmake) or build it yourself with the steps below.

### Build
To build checkmake you will need to have [golang](https://golang.org/) installed. Once you have Go installed, you can simply clone the repo and build the binary and man page yourself with the following commands.

```sh
git clone https://github.com/checkmake/checkmake
cd checkmake
make
```

## Use in CI

### MegaLinter

checkmake is [natively embedded](https://oxsecurity.github.io/megalinter/latest/descriptors/makefile_checkmake/) within [MegaLinter](https://github.com/oxsecurity/megalinter)

To install it, run `npx mega-linter-runner --install` (requires Node.js)

## Inspiration
This is totally inspired by an idea by [Dan
Buch](https://web.archive.org/web/20200916193234/https://twitter.com/meatballhat/status/768112351924985856).
