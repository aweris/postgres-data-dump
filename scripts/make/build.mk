# Directories
ROOT_DIR        ?= $(realpath $(dir $(lastword $(MAKEFILE_LIST))))/../../ # In case of ROOT_DIR not defined main make file
BUILD_DIR       := $(ROOT_DIR)/build
CMD_DIR         := $(ROOT_DIR)/cmd

# Commands
GOBUILD         := go build
GOMOD           := go mod
GOCLEAN         := go clean

# Variables
BINARY_NAME     := pdd
VERSION         := $(strip $(shell [ -d .git ] && git describe --always --tags --dirty))
VCS_REF         := $(strip $(shell [ -d .git ] && git rev-parse HEAD))
BUILD_TIMESTAMP := $(shell date -u +"%Y-%m-%dT%H:%M:%S%Z")
BUILD_LDFLAGS   := '-s -w -X "main.version=$(VERSION)" -X "main.commit=$(VCS_REF)" -X "main.date=$(BUILD_TIMESTAMP)"'

.PHONY: clean
clean: ## Cleanup everything
clean: ; $(info $(M) cleaning )
	$(Q) $(GOCLEAN)
	$(Q) $(shell rm -rf $(GOBIN) $(BUILD_DIR))

.PHONY: vendor
vendor: ## Updates vendored copy of dependencies
vendor: ; $(info $(M) running go mod vendor)
	$(Q) $(GOMOD) tidy
	$(Q) $(GOMOD) vendor

.PHONY: build
build: ## Builds binary
build: vendor $(CMD_DIR)/main.go $(wildcard *.go) $(wildcard */*.go) $(BUILD_DIR) ; $(info $(M) building binary)
	$(Q) CGO_ENABLED=0 $(GOBUILD) -a -tags netgo -ldflags $(BUILD_LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

.PHONY: version
version: ## Shows application version
	$(Q) echo $(VERSION)

$(BUILD_DIR): ; $(info $(M) creating build directory)
	$(Q) $(shell mkdir -p $@)