# This file tests how the parser handles empty and unknown lines
# Empty line follows

thisisnotarule

all:
	@echo all
.PHONY: all

clean:
	@echo clean
.PHONY: clean

test:
	@echo test
.PHONY: test


