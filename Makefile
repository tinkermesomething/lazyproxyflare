# Makefile for lazyproxyflare
# Cloudflare DNS + Caddy reverse proxy TUI manager

BINARY_NAME := lazyproxyflare
BUILD_DIR := .
INSTALL_DIR := /usr/local/bin
CONFIG_DIR := $(HOME)/.config/lazyproxyflare

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

# Version (override with: make build VERSION=1.0.0)
VERSION ?= 1.2.1

# Build flags
LDFLAGS := -s -w -X main.Version=$(VERSION)
BUILD_FLAGS := -ldflags "$(LDFLAGS)"

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

.PHONY: all build clean install uninstall test deps check help

# Default target
all: check build

# Check if Go is installed
check:
	@echo "Checking prerequisites..."
	@command -v $(GOCMD) >/dev/null 2>&1 || { \
		echo "$(RED)Error: Go is not installed$(NC)"; \
		echo "Install Go from https://go.dev/dl/ or via package manager:"; \
		echo "  macOS:  brew install go"; \
		echo "  Ubuntu: sudo apt install golang-go"; \
		echo "  Fedora: sudo dnf install golang"; \
		exit 1; \
	}
	@echo "$(GREEN)✓ Go found:$(NC) $$($(GOCMD) version)"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$(GREEN)✓ Dependencies ready$(NC)"

# Build the binary
build: deps
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/lazyproxyflare
	@echo "$(GREEN)✓ Built:$(NC) $(BUILD_DIR)/$(BINARY_NAME)"

# Build for development (with debug symbols)
build-dev: deps
	@echo "Building $(BINARY_NAME) (development)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/lazyproxyflare
	@echo "$(GREEN)✓ Built (dev):$(NC) $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Install to system
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@if [ ! -d "$(INSTALL_DIR)" ]; then \
		echo "$(RED)Error: $(INSTALL_DIR) does not exist$(NC)"; \
		exit 1; \
	fi
	@if [ ! -w "$(INSTALL_DIR)" ]; then \
		echo "$(YELLOW)Requires sudo to install to $(INSTALL_DIR)$(NC)"; \
		sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME); \
		sudo chmod 755 $(INSTALL_DIR)/$(BINARY_NAME); \
	else \
		cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME); \
		chmod 755 $(INSTALL_DIR)/$(BINARY_NAME); \
	fi
	@echo "$(GREEN)✓ Installed:$(NC) $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "Run '$(BINARY_NAME)' from anywhere to start"

# Uninstall from system
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@if [ -f "$(INSTALL_DIR)/$(BINARY_NAME)" ]; then \
		if [ ! -w "$(INSTALL_DIR)" ]; then \
			echo "$(YELLOW)Requires sudo to uninstall from $(INSTALL_DIR)$(NC)"; \
			sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME); \
		else \
			rm -f $(INSTALL_DIR)/$(BINARY_NAME); \
		fi; \
		echo "$(GREEN)✓ Removed:$(NC) $(INSTALL_DIR)/$(BINARY_NAME)"; \
	else \
		echo "$(YELLOW)Binary not found at $(INSTALL_DIR)/$(BINARY_NAME)$(NC)"; \
	fi

# Uninstall and remove config (full purge)
purge: uninstall
	@echo "Removing configuration..."
	@if [ -d "$(CONFIG_DIR)" ]; then \
		echo "$(YELLOW)Warning: This will delete all profiles and settings$(NC)"; \
		read -p "Are you sure? [y/N] " confirm; \
		if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
			rm -rf $(CONFIG_DIR); \
			echo "$(GREEN)✓ Removed:$(NC) $(CONFIG_DIR)"; \
		else \
			echo "Cancelled"; \
		fi; \
	else \
		echo "$(YELLOW)Config directory not found at $(CONFIG_DIR)$(NC)"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@echo "$(GREEN)✓ Clean$(NC)"

# Show help
help:
	@echo "$(BINARY_NAME) - Cloudflare DNS + Caddy TUI Manager"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all        Check prerequisites and build (default)"
	@echo "  check      Verify Go is installed"
	@echo "  deps       Download Go dependencies"
	@echo "  build      Build optimized binary"
	@echo "  build-dev  Build with debug symbols"
	@echo "  test       Run tests"
	@echo "  install    Install to $(INSTALL_DIR)"
	@echo "  uninstall  Remove from $(INSTALL_DIR)"
	@echo "  purge      Uninstall and remove all config (~/.config/lazyproxyflare)"
	@echo "  clean      Remove build artifacts"
	@echo "  help       Show this help"
	@echo ""
	@echo "Quick start:"
	@echo "  make && sudo make install"
