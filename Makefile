# Makefile for Skupper CNI Plugin

# Variables
PLUGIN_NAME := skupper-cni
BUILD_DIR := bin
INSTALL_DIR := /opt/cni/bin
CNI_CONFIG_DIR := /etc/cni/net.d

# Go build variables
GO := go
GOFLAGS := -v
LDFLAGS := -s -w

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
BUILD_FLAGS := -ldflags "$(LDFLAGS) -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)"

.PHONY: all
all: build

.PHONY: build
build: clean
	@echo "Building $(PLUGIN_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME) ./cmd/skupper-cni

.PHONY: build-linux
build-linux: clean
	@echo "Building $(PLUGIN_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME) ./cmd/skupper-cni

.PHONY: install
install: build
	@echo "Installing $(PLUGIN_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@install -m 0755 $(BUILD_DIR)/$(PLUGIN_NAME) $(INSTALL_DIR)/$(PLUGIN_NAME)
	@echo "Plugin installed successfully"

.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(PLUGIN_NAME) from $(INSTALL_DIR)..."
	@rm -f $(INSTALL_DIR)/$(PLUGIN_NAME)
	@echo "Plugin uninstalled successfully"

.PHONY: install-config
install-config:
	@echo "Installing example CNI configuration..."
	@mkdir -p $(CNI_CONFIG_DIR)
	@cp examples/10-skupper.conflist $(CNI_CONFIG_DIR)/
	@echo "Configuration installed successfully"

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, install it from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

.PHONY: mod-tidy
mod-tidy:
	@echo "Tidying go modules..."
	$(GO) mod tidy

.PHONY: mod-download
mod-download:
	@echo "Downloading go modules..."
	$(GO) mod download

.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t skupper-cni:$(VERSION) .

.PHONY: help
help:
	@echo "Skupper CNI Plugin Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build           - Build the plugin binary"
	@echo "  make build-linux     - Build the plugin binary for Linux"
	@echo "  make install         - Install the plugin to $(INSTALL_DIR)"
	@echo "  make uninstall       - Remove the plugin from $(INSTALL_DIR)"
	@echo "  make install-config  - Install example CNI configuration"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make test            - Run tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make fmt             - Format code"
	@echo "  make vet             - Run go vet"
	@echo "  make lint            - Run golangci-lint"
	@echo "  make mod-tidy        - Tidy go modules"
	@echo "  make mod-download    - Download go modules"
	@echo "  make docker-build    - Build Docker image"
	@echo "  make help            - Show this help message"