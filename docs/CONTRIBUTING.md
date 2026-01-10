# Contributing to GoTunnel

First off, thank you for considering contributing to GoTunnel! It's people like you that make GoTunnel such a great tool.

## Table of Contents

-   [Code of Conduct](#code-of-conduct)
-   [Getting Started](#getting-started)
-   [Development Setup](#development-setup)
-   [How Can I Contribute?](#how-can-i-contribute)
-   [Development Workflow](#development-workflow)
-   [Coding Guidelines](#coding-guidelines)
-   [Testing](#testing)
-   [Documentation](#documentation)
-   [Submitting Changes](#submitting-changes)
-   [Community](#community)

---

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to [bakarepraise@example.com](mailto:bakarepraise@example.com).

### Our Standards

**Examples of behavior that contributes to a positive environment:**

-   Using welcoming and inclusive language
-   Being respectful of differing viewpoints and experiences
-   Gracefully accepting constructive criticism
-   Focusing on what is best for the community
-   Showing empathy towards other community members

**Examples of unacceptable behavior:**

-   The use of sexualized language or imagery
-   Trolling, insulting/derogatory comments, and personal or political attacks
-   Public or private harassment
-   Publishing others' private information without explicit permission
-   Other conduct which could reasonably be considered inappropriate

---

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

-   **Go**: Version 1.21 or higher
-   **Git**: For version control
-   **Make**: (Optional) For using Makefile commands

### Quick Start

1. **Fork the repository** on GitHub
2. **Clone your fork**:

```bash
   git clone https://github.com/YOUR_USERNAME/gotunnel.git
   cd gotunnel
```

3. **Add upstream remote**:

```bash
   git remote add upstream https://github.com/bakare-dev/gotunnel.git
```

4. **Install dependencies**:

```bash
   go mod download
```

5. **Run tests**:

```bash
   go test ./...
```

6. **Build**:

```bash
   go build -o bin/gotunnel-server cmd/server/main.go
   go build -o bin/gotunnel-client cmd/client/main.go
```

---

## Development Setup

### Project Structure

```
gotunnel/
├── cmd/
│   ├── server/          # Server entry point
│   │   └── main.go
│   └── client/          # Client entry point
│       └── main.go
├── internal/
│   ├── protocol/        # Protocol implementation
│   │   ├── frame.go
│   │   ├── session.go
│   │   ├── stream.go
│   │   └── types.go
│   ├── server/          # Server logic
│   │   ├── router.go
│   │   └── public.go
│   ├── client/          # Client logic
│   │   ├── forwarder.go
│   │   └── reconnect.go
│   ├── tunnel/          # Tunnel utilities
│   │   └── http_parser.go
│   └── metrics/         # Metrics tracking
│       ├── metrics.go
│       └── display.go
├── scripts/             # Utility scripts
│   └── gen-cert.sh
├── docs/                # Documentation
│   ├── PROTOCOL.md
│   └── ARCHITECTURE.md
├── certs/              # TLS certificates (gitignored)
├── go.mod
├── go.sum
├── README.md
├── CONTRIBUTING.md
└── LICENSE
```

### Development Environment

#### VSCode Settings (Recommended)

Create `.vscode/settings.json`:

```json
{
	"go.useLanguageServer": true,
	"go.lintTool": "golangci-lint",
	"go.lintOnSave": "workspace",
	"go.formatTool": "goimports",
	"editor.formatOnSave": true,
	"go.testFlags": ["-v"],
	"go.coverOnSave": true
}
```

#### GoLand/IntelliJ IDEA

1. Open the project directory
2. Enable Go Modules integration
3. Set GOROOT to Go 1.21+
4. Enable code inspections

---

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates.

When you create a bug report, include as many details as possible:

**Template**:

```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce:

1. Start server with '...'
2. Connect client with '...'
3. Make request '...'
4. See error

**Expected behavior**
What you expected to happen.

**Actual behavior**
What actually happened.

**Environment:**

-   OS: [e.g., Ubuntu 22.04]
-   Go version: [e.g., 1.21.5]
-   GoTunnel version: [e.g., v1.0.0]
-   Server/Client: [which component]

**Logs**
```

Paste relevant logs here

```

**Additional context**
Any other relevant information.
```

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

-   **Clear title** and description
-   **Use case**: Why is this enhancement needed?
-   **Proposed solution**: How should it work?
-   **Alternatives**: What other solutions did you consider?
-   **Examples**: Code examples or mockups if applicable

### Your First Code Contribution

Unsure where to begin? Look for issues labeled:

-   `good first issue` - Good for newcomers
-   `help wanted` - Need community help
-   `bug` - Something isn't working
-   `enhancement` - New feature or request
-   `documentation` - Improvements or additions to docs

### Pull Requests

1. **Small, focused PRs** are easier to review
2. **One feature/fix per PR**
3. **Update tests** for your changes
4. **Update documentation** if needed
5. **Follow coding guidelines**

---

## Development Workflow

### 1. Create a Branch

```bash
# Update your fork
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/bug-description
```

### 2. Make Your Changes

```bash
# Make changes
vim internal/protocol/session.go

# Test as you go
go test ./internal/protocol/

# Run full test suite
go test ./...

# Check formatting
go fmt ./...

# Run linter
golangci-lint run
```

### 3. Commit Your Changes

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
# Format: <type>(<scope>): <subject>

git commit -m "feat(client): add automatic reconnection"
git commit -m "fix(protocol): prevent race condition in WriteFrame"
git commit -m "docs(readme): update installation instructions"
git commit -m "test(server): add tests for router"
```

**Types**:

-   `feat`: New feature
-   `fix`: Bug fix
-   `docs`: Documentation changes
-   `test`: Adding or updating tests
-   `refactor`: Code refactoring
-   `perf`: Performance improvements
-   `chore`: Maintenance tasks

### 4. Push and Create PR

```bash
# Push to your fork
git push origin feature/your-feature-name

# Go to GitHub and create Pull Request
# Fill out the PR template
```

---

## Coding Guidelines

### Go Style Guide

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

#### General Principles

1. **Keep it simple**: Prefer simple, readable code over clever code
2. **Handle errors**: Never ignore errors
3. **Use context**: Pass context.Context for cancellation
4. **Comment exported items**: All exported functions, types, and constants
5. **Avoid global state**: Prefer dependency injection

#### Naming Conventions

```go
// Good
func ProcessRequest(ctx context.Context, req *Request) error
type StreamManager struct { ... }
const MaxPayloadSize = 16 * 1024 * 1024

// Avoid
func process_request(req *Request) error  // Snake case
type streammanager struct { ... }         // Lowercase exported type
const max_payload_size = 16777216        // Snake case, unclear size
```

#### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to connect: %w", err)
}

// Avoid
if err != nil {
    panic(err)  // Don't panic in libraries
}

// Avoid
_ = conn.Close()  // Don't ignore errors silently
```

#### Concurrency

```go
// Use mutexes for shared state
type Session struct {
    mu      sync.Mutex
    streams map[uint32]*Stream
}

// Use channels for communication
dataChan := make(chan []byte, 16)

// Always use context for cancellation
func ProcessStream(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case data := <-dataChan:
            // Process data
        }
    }
}
```

#### Logging

```go
// Use structured logging
log.Printf("│ INFO  │ Client connected: %s", addr)
log.Printf("│ ERROR │ Failed to parse frame: %v", err)
log.Printf("│ DEBUG │ [Stream %d] Received %d bytes", streamID, n)

// Avoid
fmt.Println("Connected!")  // No context
log.Print(err)             // No context
```

### Code Organization

#### Function Length

-   Keep functions **under 50 lines**
-   Extract complex logic into separate functions
-   One function = one responsibility

#### File Organization

```go
// 1. Package declaration
package protocol

// 2. Imports (grouped: stdlib, external, internal)
import (
    "fmt"
    "io"
    "sync"

    "github.com/external/package"

    "github.com/bakare-dev/gotunnel/internal/metrics"
)

// 3. Constants
const (
    MaxPayloadSize = 16 * 1024 * 1024
)

// 4. Types
type Session struct { ... }

// 5. Constructors
func NewSession() *Session { ... }

// 6. Methods (grouped by receiver)
func (s *Session) ReadFrame() { ... }
func (s *Session) WriteFrame() { ... }

// 7. Package-level functions
func ValidateToken() { ... }
```

---

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/protocol/

# Run with race detector
go test -race ./...

# Verbose output
go test -v ./...

# Run specific test
go test -run TestSessionHandshake ./internal/protocol/
```

### Writing Tests

#### Unit Tests

```go
func TestSessionHandshake(t *testing.T) {
    // Setup
    r, w := io.Pipe()
    sess := NewSession(r, w)

    // Execute
    err := sess.ProcessHandshake(frame)

    // Assert
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if sess.State() != StateHandshaken {
        t.Errorf("Expected state %v, got %v", StateHandshaken, sess.State())
    }
}
```

#### Table-Driven Tests

```go
func TestFrameEncoding(t *testing.T) {
    tests := []struct {
        name    string
        frame   Frame
        wantErr bool
    }{
        {
            name: "valid frame",
            frame: Frame{Type: MsgHandshake, Payload: []byte("test")},
            wantErr: false,
        },
        {
            name: "oversized payload",
            frame: Frame{Type: MsgStreamData, Payload: make([]byte, MaxPayloadSize+1)},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.frame.Encode(w)
            if (err != nil) != tt.wantErr {
                t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

#### Integration Tests

```go
func TestEndToEndTunnel(t *testing.T) {
    // Start test HTTP server
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer srv.Close()

    // Start tunnel server
    // Start tunnel client
    // Make request through tunnel
    // Verify response
}
```

### Test Coverage

Aim for **>80% test coverage** for critical paths:

-   Protocol encoding/decoding
-   Session state machine
-   Stream lifecycle
-   Error handling

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Documentation

### Code Documentation

```go
// ProcessHandshake validates and processes the initial handshake frame.
// It transitions the session from INIT to HANDSHAKEN state.
//
// Returns ErrHandshakeRequired if called in wrong state.
func (s *Session) ProcessHandshake(frame *Frame) error {
    // Implementation
}
```

### README Updates

When adding features, update:

-   Feature list
-   Usage examples
-   Configuration options
-   Roadmap (if applicable)

### Architecture Docs

For significant changes, update:

-   `docs/ARCHITECTURE.md` - System design
-   `docs/PROTOCOL.md` - Protocol changes

---

## Submitting Changes

### Pull Request Process

1. **Update documentation**
2. **Add/update tests**
3. **Run full test suite**
4. **Update CHANGELOG.md** (if applicable)
5. **Fill out PR template**

### PR Template

```markdown
## Description

Brief description of changes.

## Type of Change

-   [ ] Bug fix (non-breaking change)
-   [ ] New feature (non-breaking change)
-   [ ] Breaking change
-   [ ] Documentation update

## Testing

-   [ ] Added unit tests
-   [ ] Added integration tests
-   [ ] Manual testing performed

## Checklist

-   [ ] Code follows style guidelines
-   [ ] Self-reviewed code
-   [ ] Commented hard-to-understand areas
-   [ ] Updated documentation
-   [ ] No new warnings
-   [ ] Added tests that prove fix/feature works
-   [ ] New and existing tests pass

## Related Issues

Fixes #123
```

### Review Process

1. **Automated checks** must pass (tests, linting)
2. **At least one approval** from maintainer
3. **All comments addressed**
4. **Squash commits** if needed
5. **Maintainer merges**

---

## Community

### Communication Channels

-   **GitHub Issues**: Bug reports, feature requests
-   **GitHub Discussions**: Questions, ideas, general discussion
-   **Email**: bakarepraise@example.com (for private concerns)

### Getting Help

-   Check existing issues and discussions
-   Read documentation thoroughly
-   Provide minimal reproducible examples
-   Be patient and respectful

### Recognition

Contributors will be:

-   Listed in CONTRIBUTORS.md
-   Mentioned in release notes
-   Credited in commit messages

---

## Development Tips

### Debugging

```bash
# Run with debug logging
go run cmd/client/main.go --local localhost:3000 -v

# Use delve debugger
dlv debug cmd/client/main.go -- --local localhost:3000

# Profile performance
go run cmd/server/main.go -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### Common Pitfalls

1. **Race conditions**: Always run `go test -race`
2. **Goroutine leaks**: Ensure all goroutines exit
3. **Channel deadlocks**: Always handle channel closure
4. **Mutex deadlocks**: Don't hold locks during I/O
5. **Memory leaks**: Close connections and release resources

### Useful Commands

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Update dependencies
go get -u ./...
go mod tidy

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o bin/gotunnel-linux-amd64 cmd/client/main.go
GOOS=darwin GOARCH=amd64 go build -o bin/gotunnel-darwin-amd64 cmd/client/main.go
GOOS=windows GOARCH=amd64 go build -o bin/gotunnel-windows-amd64.exe cmd/client/main.go
```

---

## License

By contributing to GoTunnel, you agree that your contributions will be licensed under its MIT License.

---

## Questions?

Feel free to reach out:

-   Open a [GitHub Discussion](https://github.com/bakare-dev/gotunnel/discussions)
-   Email: bakarepraise@example.com

Thank you for contributing to GoTunnel! 🚀
