# Makefile for building plugins

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=hello_plugin.so
BINARY_UNIX=$(BINARY_NAME)_unix

# Build flags
BUILD_FLAGS=-buildmode=plugin

.PHONY: all build clean test deps help

all: build

build: 
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME) hello_plugin.go

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

test:
	$(GOTEST) -v ./...

deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_UNIX) hello_plugin.go

# Development targets
dev: deps build

install: build
	cp $(BINARY_NAME) ../../plugins/

help:
	@echo "Available targets:"
	@echo "  all         - Run build"
	@echo "  build       - Build the plugin"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  deps        - Download dependencies"
	@echo "  build-linux - Cross compile for Linux"
	@echo "  dev         - Download dependencies and build"
	@echo "  install     - Build and copy to plugins directory"
	@echo "  help        - Show this help message"