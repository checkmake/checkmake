#
# some housekeeping tasks
#

export GO111MODULE = on
export GOFLAGS = -mod=mod

# variable definitions
NAME := checkmake
DESC := linter for Makefiles
PREFIX ?= usr/local
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDDATE := $(shell date -u +"%B %d, %Y")

BUILD_GOOS ?= $(shell go env GOOS)
BUILD_GOARCH ?= $(shell go env GOARCH)
BUILD_GOPATH :=$(shell go env GOPATH)

GOLANGCI_LINT_VERSION := latest
GOLANGCI_LINT_MAJOR_VER := v2
GOLANGCI_LINT_BIN:= $(BUILD_GOPATH)/bin/golangci-lint



IMAGE_REGISTRY ?= quay.io
REGISTRY_NAMESPACE ?= $(NAME)

IMAGE_NAME ?= checkmake

IMAGE_VERSION_TAG ?= latest

IMG ?= $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_VERSION_TAG)


ifeq ($(origin CONTAINER_CMD),undefined)
# try podman first
CONTAINER_CMD=$(shell podman version >/dev/null 2>&1 && echo podman)
ifeq ($(CONTAINER_CMD),)
#try docker if podman is not available
CONTAINER_CMD=$(shell docker version >/dev/null 2>&1 && echo docker)
endif
endif


ifeq ($(BUILD_GOOS),windows)
	EXTENSION := .exe
endif

RELEASE_ARTIFACTS_DIR := .release_artifacts
CHECKSUM_FILE := checksums.txt

$(RELEASE_ARTIFACTS_DIR):
	install -d $@

CHECKMAKE_RELEASE_BINARY = "$(RELEASE_ARTIFACTS_DIR)/checkmake-$(VERSION).$(BUILD_GOOS).$(BUILD_GOARCH)$(EXTENSION)"

BUILDER_NAME := $(if $(BUILDER_NAME),$(BUILDER_NAME),$(shell git config user.name))
ifndef BUILDER_NAME
$(error "You must set environment variable BUILDER_NAME or set a user.name in your git configuration.")
endif

EMAIL := $(if $(BUILDER_EMAIL),$(BUILDER_EMAIL),$(shell git config user.email))
ifndef EMAIL
$(error "You must set environment variable BUILDER_EMAIL or set a user.email in your git configuration.")
endif

BUILDER := $(shell echo "${BUILDER_NAME} <${EMAIL}>")

PKG_RELEASE ?= 1
PROJECT_URL := "https://github.com/checkmake/$(NAME)"
LDFLAGS := -X 'main.version=$(VERSION)' \
           -X 'main.buildTime=$(BUILDTIME)' \
           -X 'main.builder=$(BUILDER)' \
           -X 'main.goversion=$(GOVERSION)'

PACKAGES := $(shell find ./* -type d | grep -v vendor)
TEST_PKG ?= $(shell go list ./... | grep -v /vendor/)

CMD_SOURCES := $(shell find cmd -name main.go)
TARGETS := $(patsubst cmd/%/main.go,%,$(CMD_SOURCES))
MAN_SOURCES := $(shell find man -name "*.md")
MAN_TARGETS := $(patsubst man/man1/%.md,%,$(MAN_SOURCES))

INSTALLED_TARGETS = $(addprefix $(PREFIX)/bin/, $(TARGETS))
INSTALLED_MAN_TARGETS = $(addprefix $(PREFIX)/share/man/man1/, $(MAN_TARGETS))

%: cmd/%/main.go
	GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_GOARCH) go build -ldflags "$(LDFLAGS)" -o $@ $<

%.1: man/man1/%.1.md
	sed "s/REPLACE_DATE/$(BUILDDATE)/" $< | pandoc -s -t man -o $@

all: lint require $(TARGETS) $(MAN_TARGETS)
.DEFAULT_GOAL:=all

binaries: $(TARGETS)

require:
	@echo "Checking the programs required for the build are installed..."
	@pandoc --version >/dev/null 2>&1 || (echo "ERROR: pandoc is required."; exit 1)

# development tasks

test: lint
	go test -v $(TEST_PKG)

.PHONY: vet
vet:
	@go vet ./...

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT_BIN)
	@$(GOLANGCI_LINT_BIN) run

$(GOLANGCI_LINT_BIN):
	@go install github.com/golangci/golangci-lint/$(GOLANGCI_LINT_MAJOR_VER)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: lint
lint: vet golangci-lint




coverage:
	@echo "mode: set" > cover.out
	@for package in $(PACKAGES); do \
		if ls $${package}/*.go &> /dev/null; then  \
		go test -coverprofile=$${package}/profile.out $${package}; fi; \
		if test -f $${package}/profile.out; then \
	 	cat $${package}/profile.out | grep -v "mode: set" >> cover.out; fi; \
	done
	@-go tool cover -html=cover.out -o cover.html

benchmark:
	@echo "Running tests..."
	@go test -bench=. ${NAME}

# install tasks
$(PREFIX)/bin/%: %
	install -d $$(dirname $@)
	install -m 755 $< $@

$(PREFIX)/share/man/man1/%: %
	install -d $$(dirname $@)
	install -m 644 $< $@

install: $(INSTALLED_TARGETS) $(INSTALLED_MAN_TARGETS)

local-install:
	$(MAKE) install PREFIX=usr/local

# packaging tasks
packages: local-install rpm deb

deploy-packages: packages
	package_cloud push mrtazz/$(NAME)/rpm_any *.rpm
	package_cloud push mrtazz/$(NAME)/any/any *.deb


rpm: $(SOURCES)
	  fpm -t rpm -s dir \
    --name $(NAME) \
    --version $(VERSION) \
		--description "$(DESC)" \
    --iteration $(PKG_RELEASE) \
    --epoch 1 \
    --license MIT \
    --maintainer "Daniel Schauenberg <d@unwiredcouch.com>" \
    --url $(PROJECT_URL) \
    --vendor mrtazz \
    usr

deb: $(SOURCES)
	  fpm -t deb -s dir \
    --name $(NAME) \
    --version $(VERSION) \
		--description "$(DESC)" \
    --iteration $(PKG_RELEASE) \
    --epoch 1 \
    --license MIT \
    --maintainer "Daniel Schauenberg <d@unwiredcouch.com>" \
    --url $(PROJECT_URL) \
    --vendor mrtazz \
    usr

.PHONY: build-standalone
build-standalone: all $(RELEASE_ARTIFACTS_DIR)
	mv checkmake.1 $(RELEASE_ARTIFACTS_DIR)
	mv checkmake $(CHECKMAKE_RELEASE_BINARY)
	shasum -a 256 $(CHECKMAKE_RELEASE_BINARY) >> $(RELEASE_ARTIFACTS_DIR)/$(CHECKSUM_FILE)


.PHONY: clean-release-artifacts
clean-release-artifacts:
	rm -rf $(RELEASE_ARTIFACTS_DIR)
.PHONY: release-artifacts
release-artifacts: clean-release-artifacts
	hack/create-release-artifacts.sh
.PHONY: github-release
github-release: release-artifacts
	gh release create $(VERSION) --title 'Release $(VERSION)' --notes-file docs/releases/$(VERSION).md $(RELEASE_ARTIFACTS_DIR)/*


# clean up tasks
clean: clean-docs clean-deps
	$(RM) -r ./usr
	$(RM) $(TARGETS)

clean-docs:
	$(RM) $(MAN_TARGETS)

pizza: # ignore checkmake
	@echo ""
	@echo "🍕 🍕 🍕 🍕 🍕 🍕   make.pizza 🍕 🍕 🍕 🍕 🍕 🍕 "
	@echo ""
	@echo "https://twitter.com/mrb_bk/status/760636493710983168"
	@echo ""
	@echo ""

.PHONY: all test rpm deb install local-install packages coverage clean-deps clean clean-docs pizza binaries

.PHONY: image-build
image-build:
	$(CONTAINER_CMD) build --build-arg BUILDER_NAME='$(BUILDER_NAME)' --build-arg BUILDER_EMAIL='$(EMAIL)' -t $(IMG) -f Containerfile .

.PHONY: image-push
image-push:
	$(CONTAINER_CMD) push $(IMG)



