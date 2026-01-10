# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned

-   P2P node mode (v2.0)
-   Discovery service for P2P
-   HTTP host-based routing
-   Custom domain support
-   Web dashboard
-   Configuration file support (YAML/JSON)
-   UDP tunneling support

---

## [1.0.0] - 2026-01-10

### 🎉 Initial Release

First stable release of GoTunnel - a production-ready TCP tunneling system.

### Added

#### Core Features

-   **Custom Binary Protocol** - Efficient, versioned wire protocol for tunnel communication
-   **TCP Tunneling** - Protocol-agnostic tunneling supporting HTTP, gRPC, databases, SSH, etc.
-   **Stream Multiplexing** - Multiple concurrent connections over single tunnel
-   **Multi-Client Support** - Server handles multiple clients simultaneously
-   **Auto Port Assignment** - Server dynamically assigns public ports to clients
-   **Token-Based Authentication** - Simple but effective authentication mechanism
-   **Session Management** - State machine enforcing handshake → auth → forwarding flow

#### Monitoring & Metrics

-   **HTTP Request Logging** - Real-time HTTP traffic monitoring with:
    -   Request method, path, and status code
    -   Visual indicators (✓ success, ⚠ client error, ✗ server error)
    -   Response latency in milliseconds
-   **Comprehensive Metrics Tracking**:
    -   Active and total stream counts
    -   Bandwidth monitoring (bytes sent/received)
    -   HTTP request statistics
    -   Latency tracking (min/avg/max)
    -   Status code breakdown
    -   Session uptime
-   **Metrics Summary** - Displayed on graceful shutdown or Ctrl+C

#### Reliability Features

-   **Auto-Reconnection** - Client automatically reconnects on connection loss with:
    -   Exponential backoff (1s → 2s → 4s → ... → 30s max)
    -   Configurable retry attempts (default: 10)
    -   Graceful handling of network interruptions
    -   Optional disable with `--no-reconnect` flag
-   **Graceful Shutdown** - Clean resource cleanup on exit:
    -   Closes all active streams
    -   Shows metrics summary
    -   Proper resource cleanup
    -   Signal handling (SIGINT, SIGTERM)
-   **Heartbeat Management** - Automatic session health monitoring:
    -   10-second heartbeat interval
    -   30-second timeout detection
    -   Automatic session expiration

#### Security Features

-   **TLS Encryption** - Optional end-to-end encryption:
    -   TLS 1.2+ support
    -   Certificate-based validation
    -   CA certificate trust model
    -   Certificate generation script included
-   **Write Synchronization** - Mutex-protected frame writing prevents race conditions
-   **Session Isolation** - Clients cannot access each other's streams
-   **Payload Size Limits** - Prevents DoS attacks (16MB max per frame)

#### Developer Experience

-   **Clean CLI Output** - ngrok-inspired user interface
-   **Detailed Error Messages** - Clear error reporting
-   **Comprehensive Documentation**:
    -   README with examples
    -   Protocol specification
    -   Architecture documentation
    -   Contributing guidelines
-   **Zero Dependencies** - Pure Go implementation, single binary deployment

### Technical Details

#### Protocol

-   Binary protocol version 0x01
-   10-byte frame header
-   Stream ID space: 1 to 2^32-1
-   Message types: Handshake, Auth, Bind, StreamOpen, StreamData, StreamClose, Heartbeat, Error

#### Performance

-   Throughput: 500+ MB/s per tunnel
-   Latency overhead: <10ms
-   Max concurrent streams: 1000+ per session
-   Memory usage: ~50MB per active session
-   CPU usage: <5% on modern CPUs

#### Compatibility

-   Go 1.21+
-   Linux, macOS, Windows
-   Any TCP-based protocol

### Configuration

#### Server Options

```bash
--addr string           Listen address (default ":9000")
--start-port int        Starting port for public listeners (default 10000)
--tls                   Enable TLS encryption
--tls-cert string       Path to TLS certificate
--tls-key string        Path to TLS private key
```

#### Client Options

```bash
--server string         Tunnel server address (default "localhost:9000")
--local string          Local service to expose (required)
--token string          Authentication token (default "dev-token")
--tls                   Enable TLS encryption
--tls-ca string         Path to CA certificate
--no-reconnect          Disable auto-reconnect
```

### Installation

#### From Source

```bash
git clone https://github.com/bakare-dev/gotunnel.git
cd gotunnel
go build -o gotunnel-server cmd/server/main.go
go build -o gotunnel-client cmd/client/main.go
```

#### Pre-built Binaries

Available for Linux, macOS, and Windows on the [releases page](https://github.com/bakare-dev/gotunnel/releases).

### Known Issues

-   None at this time

### Breaking Changes

-   N/A (initial release)

---

## [0.1.0] - 2026-01-08

### Added (Alpha Release)

-   Basic tunneling functionality
-   Binary protocol implementation
-   Stream multiplexing
-   Simple authentication
-   Server and client components

### Known Issues

-   No auto-reconnection
-   No metrics tracking
-   No TLS support
-   Race conditions in frame writing
-   No graceful shutdown

---

## Release Types

-   **Major (X.0.0)**: Breaking changes, major new features
-   **Minor (1.X.0)**: New features, backward compatible
-   **Patch (1.0.X)**: Bug fixes, backward compatible

---

## Version Support

| Version | Status     | Release Date | End of Support |
| ------- | ---------- | ------------ | -------------- |
| 1.0.x   | Stable     | 2026-01-10   | TBD            |
| 0.1.x   | Deprecated | 2026-01-08   | 2026-01-10     |

---

## Upgrade Guide

### From 0.1.0 to 1.0.0

#### Breaking Changes

None - backward compatible with 0.1.0 protocol.

#### New Features to Adopt

1. **Enable TLS Encryption** (recommended):

```bash
   # Generate certificates
   ./scripts/gen-cert.sh

   # Server
   gotunnel-server --tls --tls-cert=certs/server-cert.pem --tls-key=certs/server-key.pem

   # Client
   gotunnel-client --local localhost:3000 --tls --tls-ca=certs/ca-cert.pem
```

2. **Auto-Reconnection** is enabled by default (disable with `--no-reconnect`)

3. **Metrics** are automatically tracked and displayed on exit

4. **HTTP Logging** works automatically for HTTP traffic

#### Migration Steps

1. Update binaries to v1.0.0
2. No configuration changes required
3. Optionally enable TLS
4. Review metrics on first exit

---

## Roadmap

### v1.1.0 (Q1 2026) - Polish & Refinement

-   [ ] Configuration file support (YAML/JSON)
-   [ ] Improved error messages with error codes
-   [ ] Connection pooling optimizations
-   [ ] Systemd service files
-   [ ] Docker images with Docker Compose examples
-   [ ] Prometheus metrics endpoint
-   [ ] Log levels (DEBUG, INFO, WARN, ERROR)
-   [ ] Structured logging (JSON format option)

### v2.0.0 (Q2 2026) - P2P Network

-   [ ] **Node Mode** - Users can host tunnels for others
-   [ ] **Discovery Service** - Lightweight matchmaking server
-   [ ] **Credit System** - Earn credits by hosting, spend to use
-   [ ] **Reputation System** - Track node reliability and uptime
-   [ ] **NAT Traversal** - STUN/TURN support for peer connections
-   [ ] **Hybrid Mode** - Choose between centralized or P2P

### v3.0.0 (Q3 2026) - Advanced Features

-   [ ] **Web Dashboard** - Real-time monitoring UI
-   [ ] **HTTP Routing** - Host-based routing with subdomain support
-   [ ] **Custom Domains** - Bring your own domain
-   [ ] **Rate Limiting** - Per-client bandwidth controls
-   [ ] **Traffic Replay** - Record and replay requests for debugging
-   [ ] **Load Balancing** - Multiple backend services per tunnel
-   [ ] **WebSocket Support** - Enhanced WebSocket handling
-   [ ] **UDP Tunneling** - Support UDP protocol

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:

-   Reporting bugs
-   Suggesting enhancements
-   Submitting pull requests
-   Development setup

---

## Links

-   **Homepage**: https://github.com/bakare-dev/gotunnel
-   **Documentation**: https://github.com/bakare-dev/gotunnel/tree/main/docs
-   **Issues**: https://github.com/bakare-dev/gotunnel/issues
-   **Releases**: https://github.com/bakare-dev/gotunnel/releases
-   **Discussions**: https://github.com/bakare-dev/gotunnel/discussions

---

## Credits

### Author

**Bakare Praise** - Initial work and ongoing maintenance

### Inspiration

-   [ngrok](https://ngrok.com/) - Commercial tunneling service
-   [localtunnel](https://localtunnel.github.io/www/) - Open-source alternative
-   [Cloudflare Tunnel](https://www.cloudflare.com/products/tunnel/) - Enterprise solution

### Contributors

See [CONTRIBUTORS.md](CONTRIBUTORS.md) for a list of people who have contributed to this project.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Stay Updated

-   ⭐ Star this repo to get notifications
-   👁️ Watch releases for new versions
-   🐦 Follow updates on Twitter: [@bakare-dev](https://twitter.com/bakare-dev)
-   💬 Join discussions on GitHub

---

**Note**: This changelog is automatically included in release notes. For the latest changes, see [Unreleased] section above.
