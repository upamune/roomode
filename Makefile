.PHONY: build clean test lint install help

# Build settings
BINARY_NAME=roomode
MAIN_PACKAGE=./cmd/roomode
GO=go
GOFLAGS=-trimpath
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "dev")
GIT_DIRTY=$(shell git status --porcelain 2>/dev/null)
ifneq ($(GIT_DIRTY),)
  VERSION:=$(VERSION)-dirty
endif
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X github.com/upamune/roomode/internal/cmd.Version=$(VERSION) -X github.com/upamune/roomode/internal/cmd.Commit=$(COMMIT) -X github.com/upamune/roomode/internal/cmd.BuildDate=$(BUILD_DATE)"

# Default target
.DEFAULT_GOAL := help

# Help message
help:
	@echo "Available commands:"
	@echo "  make build     - Build $(BINARY_NAME)"
	@echo "  make install   - Install $(BINARY_NAME)"
	@echo "  make test      - Run tests"
	@echo "  make lint      - Run lint"
	@echo "  make fmt       - Run go fmt"
	@echo "  make clean     - Remove build artifacts"
	@echo "  make help      - Show this help message"

# Install
install:
	$(GO) install $(GOFLAGS) $(LDFLAGS) $(MAIN_PACKAGE)

# Build
build:
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

# Test
test:
	$(GO) test -v ./...

# fmt
fmt:
	@goimports -local github.com/upamune/roomode -w .

# Lint
lint:
	$(GO) vet ./...
	revive -config ./revive.toml ./...

# Clean
clean:
	rm -f $(BINARY_NAME)
	$(GO) clean
