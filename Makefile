include help.mk

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.DEFAULT_GOAL := start

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

.PHONY: lint
lint: ## run golangci-lint
	golangci-lint run ${ROOT_DIR}...

.PHONY: install-hooks
install-hooks: ## install git hooks
	@echo Installing git hooks...
	@go run -C .githooks install.go

.PHONY: start-docker
start: ## start service
	docker-compose up video-to-podcast-service --build
