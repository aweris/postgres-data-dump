# Directories
ROOT_DIR  ?= $(realpath $(dir $(lastword $(MAKEFILE_LIST))))/../../ # In case of ROOT_DIR not defined main make file

# All go files belong to project
GOFILES    = $(shell find $(ROOT_DIR) -type f -name '*.go' -not -path '$(ROOT_DIR)/vendor/*')

.PHONY: fix
fix: ## Fix found issues (if it's supported by the $(GOLANGCI_LINT))
fix: $(GOLANGCI_LINT) ; $(info $(M) runing golangci-lint run --fix)
	$(Q) $(GOLANGCI_LINT) run --fix --enable-all -c .golangci.yml

.PHONY: fmt
fmt: ## Runs gofmt
fmt: ; $(info $(M) runnig gofmt )
	$(Q) gofmt -d -s $(GOFILES)

.PHONY: lint
lint: ## Runs golangci-lint analysis
lint: $(GOLANGCI_LINT) fmt ; $(info $(M) runnig golangci-lint analysis)
	$(Q) $(GOLANGCI_LINT) run
