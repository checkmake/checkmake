.PHONY: poetry-publish-test
poetry-publish-test: ARTIFACT_VERSION=$(ARTIFACT_TEST_VERSION)
poetry-publish-test: PUBLISH_SOURCE=$(PROJECT_ID)
poetry-publish-test: build
	@echo "Published $(ARTIFACT_VERSION) to $(PUBLISH_SOURCE)"

build.o: CFLAGS += -O2
build.o: build.c
	cc $(CFLAGS) -c build.c

.PHONY: all clean test build
all:
clean:
test:
build:

poetry-publish-test: PUBLISH_SOURCE := $(PROJECT_ID)
build.o: CFLAGS ?= -O0
%.o: override CFLAGS ?= -g
prog: export PATH = /bin

foo: ; echo a=b
bar: baz
	echo hello
