#!/usr/bin/env bash


        BUILD_GOARCH=amd64 BUILD_GOOS=freebsd make build-standalone
	BUILD_GOARCH=amd64 BUILD_GOOS=linux make build-standalone
	BUILD_GOARCH=arm64 BUILD_GOOS=linux make build-standalone
	BUILD_GOARCH=amd64 BUILD_GOOS=darwin make build-standalone
	BUILD_GOARCH=arm64 BUILD_GOOS=darwin make build-standalone
	BUILD_GOARCH=amd64 BUILD_GOOS=windows make build-standalone
	BUILD_GOARCH=arm64 BUILD_GOOS=windows make build-standalone
