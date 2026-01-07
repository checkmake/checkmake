# Changelog

## v0.3.0 (8th January 2026)
- no longer experimental
- project moved to github.com/checkmake/checkmake organization
- CLI rewrite: migrated from docopt to cobra
- JSON formatter for CI/tooling integration
- rule additions:
  - uniquetargets
- parser improvements: better rule detection, inline recipe parsing, PHONY handling
- config-aware `list-rules` command
- global `~/checkmake.ini` fallback configuration
- configurable minphony rule
- platform support: Windows, darwin.arm64, linux.arm64
- container images published to quay.io
- updated to Go 1.25

## v0.2.2 (11th April 2023)
- support for multiple makefiles
- pre-commit hook improvements
- CI integration documentation

## v0.2.1 (18th February 2022)
- build with Go 1.17
- standalone binaries on release
- container builds on releases only

## v0.2.0 (30th September 2021)
- custom formatter
- lots of build improvements
- addition of a docker build
- rule additions
  - maxbodylength
  - timestampexpanded

## v0.1.0 (30th August 2016)
- initial release
- rules:
  - phonydeclared
  - minphony
