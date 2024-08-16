include help.mk

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.DEFAULT_GOAL := start-services

.PHONY: update
update: ## pulls git repo
	@git -C ${ROOT_DIR} pull

.PHONY: test
test: ## run golang test
	go test ./...

.PHONY: start
start: ## start via docker
	docker build . -t v2p
	docker run --rm -p 8080:8080 v2p

.PHONY: start-services-rebuild
start-services-rebuild: ## rebuild and start service with webhook
	python ${ROOT_DIR}start-docker-compose.py --rebuild

.PHONY: start-services
start-services: ## start with webhook
	python ${ROOT_DIR}start-docker-compose.py