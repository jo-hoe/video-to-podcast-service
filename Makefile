include help.mk

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.DEFAULT_GOAL := start

.PHONY: start-rebuild
start-rebuild: ## rebuild and start via docker
	python ${ROOT_DIR}start-docker-compose.py --rebuild

.PHONY: start
start: ## start via docker
	python ${ROOT_DIR}start-docker-compose.py