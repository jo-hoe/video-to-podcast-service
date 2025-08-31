include help.mk

# Minimal Makefile for video-to-podcast-service
# Cross-platform, no bash code, no OS-dependent commands

.DEFAULT_GOAL := help

# Configuration
CLUSTER_NAME := video-podcast-cluster
NAMESPACE := video-to-podcast
HELM_CHART := ./charts/video-to-podcast
K3D_VALUES := $(HELM_CHART)/values.yaml
ROOT_DIR ?= $(CURDIR)

# =============================================================================
# Essential Targets
# =============================================================================

.PHONY: test
test: ## run Go tests
	go test -timeout 300s ./...

.PHONY: lint
lint: ## run golangci-lint
	golangci-lint run ./...

.PHONY: start-k3d
start-k3d: ## create k3d cluster with registry and push images
	k3d cluster create --config k3d/$(CLUSTER_NAME).yaml
	kubectl wait --for=condition=Ready nodes --all --timeout=300s
	docker build -f Dockerfile.api -t localhost:5000/video-to-podcast-api:latest .
	docker build -f Dockerfile.ui -t localhost:5000/video-to-podcast-ui:latest .
	docker push localhost:5000/video-to-podcast-api:latest
	docker push localhost:5000/video-to-podcast-ui:latest
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	helm upgrade --install $(CLUSTER_NAME) $(HELM_CHART) --namespace $(NAMESPACE) --values $(K3D_VALUES) --wait --timeout=300s
	kubectl apply -f k3d/service.yaml -n $(NAMESPACE)

.PHONY: stop-k3d
stop-k3d: ## destroy k3d cluster
	k3d cluster delete $(CLUSTER_NAME)

.PHONY: restart-k3d
restart-k3d: stop-k3d start-k3d ## restart k3d cluster

.PHONY: helm-test
helm-test: ## run Helm tests
	kubectl delete pods -l "helm.sh/hook=test" -n $(NAMESPACE) --ignore-not-found=true
	helm test $(CLUSTER_NAME) -n $(NAMESPACE) --timeout=300s --logs

.PHONY: generate-helm-docs
generate-helm-docs: ## generates helm docu in /helm folder 
	@docker run --rm -v "$(ROOT_DIR)/charts/video-to-podcast:/helm-docs" -w /helm-docs jnorwood/helm-docs:latest
