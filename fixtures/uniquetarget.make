# Makefile fixture for testing the "uniquetargets" rule.

.PHONY: all clean test

all:
	@echo "Building project..."

clean:
	rm -rf build/

test:
	@echo "Running tests..."

all:
	@echo "Rebuilding project..."

test:
	@echo "Re-running tests..."

