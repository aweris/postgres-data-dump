# Directories
ROOT_DIR        ?= $(realpath $(dir $(lastword $(MAKEFILE_LIST))))/../../ # In case of ROOT_DIR not defined main make file

# Commands
GOTEST          := go test

.PHONY: test
test: ## Runs go test
test: ; $(info $(M) runnig tests)
	$(Q) $(GOTEST) -race -cover -v $(ROOT_DIR)/...