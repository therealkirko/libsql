.PHONY: build build-server build-cli test clean install

# Binary names
SERVER_BIN=bin/turbo-server
CLI_BIN=bin/turbo-cli

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get

all: build

build: build-server build-cli

build-server:
	@echo "Building server..."
	@mkdir -p bin
	$(GOBUILD) -o $(SERVER_BIN) ./cmd/server

build-cli:
	@echo "Building CLI..."
	@mkdir -p bin
	$(GOBUILD) -o $(CLI_BIN) ./cmd/cli

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf databases/
	@rm -f turbo-meta.db
	@rm -f turbo-meta.db-wal
	@rm -f turbo-meta.db-shm

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

run-server: build-server
	@echo "Starting server..."
	./$(SERVER_BIN)

install: build
	@echo "Installing binaries..."
	@cp $(SERVER_BIN) $(GOPATH)/bin/
	@cp $(CLI_BIN) $(GOPATH)/bin/

help:
	@echo "Turbo DB Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build         - Build both server and CLI"
	@echo "  make build-server  - Build server only"
	@echo "  make build-cli     - Build CLI only"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts and databases"
	@echo "  make deps          - Download and tidy dependencies"
	@echo "  make run-server    - Build and run server"
	@echo "  make install       - Install binaries to GOPATH/bin"
