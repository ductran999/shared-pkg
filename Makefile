PKG_SCRIPTS=scripts

default: help

help: ## Show help for each of the Makefile commands
	@awk 'BEGIN \
		{FS = ":.*##"; printf "Usage: make ${cyan}<command>\n${white}Commands:\n"} \
		/^[a-zA-Z_-]+:.*?##/ \
		{ printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' \
		$(MAKEFILE_LIST)

.PHONY: coverage 
coverage: ## code coverage
	${PKG_SCRIPTS}/coverage.sh

.PHONY: tidy
tidy: ## Tidy up the go.mod
	go mod tidy

.PHONY: deps
deps: ## Install dependencies
	go install github.com/vektra/mockery/v2@v2.46.3

.PHONY: lint
lint: ## Run linters
	golangci-lint run --timeout 10m --config .golangci.yml

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: check
check: tidy vet lint ## Run tidy, vet and lint (stop on first failure)
