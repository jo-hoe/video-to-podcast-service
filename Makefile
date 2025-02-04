include help.mk

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.DEFAULT_GOAL := start-service

.PHONY: update
update: ## pulls git repo
	@git -C ${ROOT_DIR} pull
	go mod tidy

.PHONY: test
test: ## run golang test
	go test ./...

.PHONY: start
start: ## start via docker
	docker build . -t v2p
	docker run --rm -p 8080:8080 v2p

.PHONY: start-service
start-service: ## start service
	python ${ROOT_DIR}start-docker-compose.py --services video-to-podcast-service

.PHONY: start-service-rebuild
start-service-rebuild: ## rebuild and start service
	python ${ROOT_DIR}start-docker-compose.py --rebuild --services video-to-podcast-service

.PHONY: start-services-rebuild
start-services-rebuild: ## start service with webhook
	python ${ROOT_DIR}start-docker-compose.py --rebuild --services video-to-podcast-service mail-webhook-service
