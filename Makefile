.PHONY: build clean test install docker


VERSION ?= 1.0.0
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin


build: build-server build-client

build-server:
	@echo "Building server..."
	@go build -o $(GOBIN)/gotunnel-server cmd/server/main.go

build-client:
	@echo "Building client..."
	@go build -o $(GOBIN)/gotunnel-client cmd/client/main.go

build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/gotunnel-server-linux-amd64 cmd/server/main.go
	GOOS=linux GOARCH=amd64 go build -o dist/gotunnel-client-linux-amd64 cmd/client/main.go
	GOOS=darwin GOARCH=amd64 go build -o dist/gotunnel-server-darwin-amd64 cmd/server/main.go
	GOOS=darwin GOARCH=amd64 go build -o dist/gotunnel-client-darwin-amd64 cmd/client/main.go
	GOOS=darwin GOARCH=arm64 go build -o dist/gotunnel-server-darwin-arm64 cmd/server/main.go
	GOOS=darwin GOARCH=arm64 go build -o dist/gotunnel-client-darwin-arm64 cmd/client/main.go
	GOOS=windows GOARCH=amd64 go build -o dist/gotunnel-server-windows-amd64.exe cmd/server/main.go
	GOOS=windows GOARCH=amd64 go build -o dist/gotunnel-client-windows-amd64.exe cmd/client/main.go
	@echo "Binaries created in dist/"

test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(GOBIN) dist/

install:
	@echo "Installing..."
	@go install ./cmd/server
	@go install ./cmd/client

docker-build:
	@echo "Building Docker images..."
	docker build -t praisebaka/gotunnel-server:$(VERSION) -f Dockerfile.server .
	docker build -t praisebaka/gotunnel-client:$(VERSION) -f Dockerfile.client .

docker-push:
	docker push praisebaka/gotunnel-server:$(VERSION)
	docker push praisebaka/gotunnel-client:$(VERSION)