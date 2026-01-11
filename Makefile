.PHONY: build clean test install

VERSION ?= 1.0.1
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin

build:
	@echo "Building gotunnel..."
	@mkdir -p $(GOBIN)
	@go build -o $(GOBIN)/gotunnel cmd/gotunnel/*.go
	@echo "✓ Binary created: $(GOBIN)/gotunnel"

build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p dist
	
	# Linux AMD64
	@echo "Building Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build -o dist/gotunnel-linux-amd64 cmd/gotunnel/*.go
	
	# macOS Intel
	@echo "Building macOS Intel..."
	@GOOS=darwin GOARCH=amd64 go build -o dist/gotunnel-darwin-amd64 cmd/gotunnel/*.go
	
	# macOS Apple Silicon
	@echo "Building macOS Apple Silicon..."
	@GOOS=darwin GOARCH=arm64 go build -o dist/gotunnel-darwin-arm64 cmd/gotunnel/*.go
	
	# Windows
	@echo "Building Windows..."
	@GOOS=windows GOARCH=amd64 go build -o dist/gotunnel-windows-amd64.exe cmd/gotunnel/*.go
	
	@echo ""
	@echo "✓ All binaries created in dist/"
	@ls -lh dist/

test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(GOBIN) dist/

install:
	@echo "Installing gotunnel to GOPATH..."
	@go install ./cmd/gotunnel
	@echo "✓ Installed to $(shell go env GOPATH)/bin/gotunnel"