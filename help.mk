# Helper file for Makefile to generate documentation for make targets

MAKEFILE_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
ABSOLUT_MAKEFILE_PATH := $(MAKEFILE_DIR)Makefile

# Determine the operating system
ifeq ($(OS),Windows_NT)
    DETECT_OS := Windows
else
    DETECT_OS := $(shell uname -s 2>/dev/null || echo Unknown)
endif

.PHONY: help
help:
ifeq ($(DETECT_OS),Windows)
	@powershell -Command "& { \
		Get-Content $(ABSOLUT_MAKEFILE_PATH) | ForEach-Object { \
			if ($$_ -match '([a-zA-Z0-9-_]{0,}:).+##\s?(.+)') { \
				'{0,-25} {1}' -f $$Matches[1], $$Matches[2] \
			} \
		} \
	}"
else
	@sed -ne 's/^\([^[:space:]]*\):.*##/\1: /p' $(ABSOLUT_MAKEFILE_PATH) | awk '{printf "%-25s %s\n", $$1, substr($$0, index($$0,$$2))}'
endif