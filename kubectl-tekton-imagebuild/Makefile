.PHONY: build test clean install

BINARY_NAME=kubectl-tekton-imagebuild
BUILD_DIR=.
INSTALL_DIR=/usr/local/bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

test:
	go test -v ./...

test-coverage:
	go test -cover ./...

clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)

install: build
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/

install-user: build
	mkdir -p ~/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) ~/bin/

fmt:
	go fmt ./...

vet:
	go vet ./...

deps:
	go mod download
	go mod tidy

all: deps fmt vet test build