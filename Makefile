.DEFAULT_GOAL := help

.PHONY: all
all: ## all
	@go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.9
	@$(shell go env GOPATH)/bin/controller-gen paths="./..." object crd:trivialVersions=true output:crd:artifacts:config=manifests/crd

.PHONY: test
test: ## test
	@go test ./... -race -bench . -benchmem -trimpath -cover

.PHONY: lint
lint: ## lint
	@go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.26.0
	@$(shell go env GOPATH)/bin/golangci-lint run --timeout=10m --enable=goimports --fix
	@go get github.com/instrumenta/kubeval@0.14.0
	@$(shell go env GOPATH)/bin/kubeval --strict --ignore-missing-schemas $(shell find manifests -type f -name kustomization.yaml -prune -o -type f -name '*.yaml' -print)

.PHONY: dev
dev: ## dev
	@skaffold dev

.PHONY: help
help: ## help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
