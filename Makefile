include help.mk

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.DEFAULT_GOAL := start-service

.PHONY: build
build: ## build binary
	go build ${ROOT_DIR}

.PHONY: update
update: ## pulls git repo
	@git -C ${ROOT_DIR} pull
	go mod tidy

.PHONY: test
test: ## run golang test (including integration tests)
	go test -timeout 0  ./...

.PHONY: start
start: ## start via docker
	docker build . -t v2p
	docker run --rm -p 8080:8080 v2p

.PHONY: start-service
start-service: ## start service
	docker-compose up video-to-podcast-service

.PHONY: start-service-rebuild
start-service-rebuild: ## rebuild and start service
	docker-compose up --build video-to-podcast-service

.PHONY: start-services-rebuild
start-services-rebuild: ## start service with webhook
	docker-compose up --build video-to-podcast-service mail-webhook-service
