include help.mk

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.DEFAULT_GOAL := start-docker

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

# K3d cluster management
IMAGE_NAME := video-to-podcast-service
IMAGE_VERSION := 1.0.0

.PHONY: create-cluster
create-cluster:
	@k3d cluster create --config ${ROOT_DIR}k3d/podcastcluster.yaml

.PHONY: build-and-push
build-and-push: ## build and push docker image to local registry
	@docker build -f ${ROOT_DIR}Dockerfile . -t ${IMAGE_NAME}
	@docker tag ${IMAGE_NAME} localhost:5000/${IMAGE_NAME}:${IMAGE_VERSION}
	@docker push localhost:5000/${IMAGE_NAME}:${IMAGE_VERSION}

.PHONY: deploy-helm
deploy-helm:
	@helm install videopodcast --set service.enabled=true \
	                           --set service.type=LoadBalancer \
	                           --set service.port=8081 \
	                           --set service.targetPort=8080 \
	                           --set image.repository=registry.localhost:5000/${IMAGE_NAME} \
	                           --set image.tag=${IMAGE_VERSION} \
	                           ${ROOT_DIR}charts/video-to-podcast-service

.PHONY: start-k3d
start-k3d: create-cluster build-and-push deploy-helm ## starts k3d and deploys local image with loadbalancer

.PHONY: stop-k3d
stop-k3d: ## stop K3d cluster
	@k3d cluster delete --config ${ROOT_DIR}k3d/podcastcluster.yaml

.PHONY: restart-k3d
restart-k3d: stop-k3d start-k3d ## restart the k3d cluster

.PHONY: start-docker
start-docker: ## start service
	docker-compose up video-to-podcast-service --build

.PHONY: generate-helm-docs
generate-helm-docs: ## generates helm docu in /helm folder 
	@docker run --rm --volume "$(ROOT_DIR)charts/video-to-podcast-service:/helm-docs" jnorwood/helm-docs:latest
