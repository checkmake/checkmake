define MULTILINE_STRING

some-test:
1
2
3
endef
export MULTILINE_STRING

.PHONY: test
test:
	@echo "$$MULTILINE_STRING"

.PHONY: all
all: test
	@echo all

.PHONY: clean
clean:
	@echo clean
