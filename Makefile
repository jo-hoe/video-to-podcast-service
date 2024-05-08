include help.mk

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))


.PHONY: start
start: ## rebuild and start via docker
	python start-docker-compose.py