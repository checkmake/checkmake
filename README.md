# checkmake

[![Build Status](https://travis-ci.org/mrtazz/checkmake.svg?branch=master)](https://travis-ci.org/mrtazz/checkmake)
[![Coverage Status](https://coveralls.io/repos/mrtazz/checkmake/badge.svg?branch=master&service=github)](https://coveralls.io/github/mrtazz/checkmake?branch=master)
[![Packagecloud](https://img.shields.io/badge/packagecloud-available-brightgreen.svg)](https://packagecloud.io/mrtazz/checkmake)
[![MIT license](https://img.shields.io/badge/license-MIT-blue.svg)](http://opensource.org/licenses/MIT)

## Overview
checkmake is an experimental tool for linting and checking Makefiles. It may
not do what you want it to.

## Usage

```
% checkmake Makefile

% checkmake --help
checkmake.

Usage:
checkmake [--debug|--config=<configPath>] <makefile>
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

## Installation

There are packages for linux up [on
packagecloud.io](https://packagecloud.io/mrtazz/checkmake) but you can also
just clone the repo and build the binary and man page yourself.

```
go get github.com/mrtazz/checkmake
cd $GOPATH/src/github.com/mrtazz/checkmake
make
```

## Inspiration
This is totally inspired by an idea by [Dan
Buch](https://twitter.com/meatballhat/status/768112351924985856).
