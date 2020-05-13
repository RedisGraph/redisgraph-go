# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

.PHONY: all test coverage
all: test coverage examples

get:
	$(GOGET) -t -v ./...

examples: get
	@echo " "
	@echo "Building the examples..."
	$(GOBUILD) ./examples/redisgraph_tls_client/.

test: get
	$(GOTEST) -race -covermode=atomic ./...

coverage: get test
	$(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic ./...

