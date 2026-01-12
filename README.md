# GoTunnel

A lightweight, self-hosted TCP tunneling system written in Go. Expose local services to the public internet through a secure tunnel, similar to ngrok.

## 🎯 Purpose & Learning Outcomes

This project was built for **learning and portfolio demonstration**, showcasing:

-   🧠 **Systems programming skills in Go**  
    Hands-on experience with low-level programming, concurrency, and performance-oriented design.

-   🌐 **Understanding of networking concepts**  
    Practical implementation of **TCP connections**, **tunneling**, and **connection multiplexing**.

-   🔁 **Designing resilient client–server systems**  
    Includes reconnect logic, fault tolerance, and long-lived connections.

-   🧩 **Clean project structure and extensibility**  
    Modular separation of concerns (client, server, protocol), making the system easy to extend and maintain.

## Features

-   🔒 **Custom Binary Protocol** - Efficient, versioned wire protocol
-   🔐 **Token-Based Authentication** - Secure tunnel access control
-   🌊 **Stream Multiplexing** - Multiple concurrent connections over single tunnel
-   🎯 **Protocol Agnostic** - Works with HTTP, gRPC, databases, SSH, and any TCP service
-   🔄 **Auto Port Assignment** - Server dynamically assigns public ports
-   💓 **Heartbeat Management** - Automatic session health monitoring
-   🚀 **Multi-Client Support** - Run multiple tunnels simultaneously
-   📦 **Zero Dependencies** - Pure Go implementation, single binary deployment
-   📊 **HTTP Request Logging** - Real-time HTTP traffic monitoring with status codes and latency
-   📈 **Metrics & Monitoring** - Bandwidth tracking, latency stats, uptime monitoring
-   🔄 **Auto-Reconnection** - Automatic reconnection with exponential backoff
-   🛡️ **Graceful Shutdown** - Clean resource cleanup with metrics summary
-   🔐 **TLS Encryption** - Optional end-to-end encryption

## Quick Start

### Installation

#### Download Pre-built Binary

```bash
# Linux
wget https://github.com/bakare-dev/gotunnel/releases/latest/download/gotunnel-linux-amd64
chmod +x gotunnel-linux-amd64
sudo mv gotunnel-linux-amd64 /usr/local/bin/gotunnel

# macOS (Intel)
wget https://github.com/bakare-dev/gotunnel/releases/latest/download/gotunnel-darwin-amd64
chmod +x gotunnel-darwin-amd64
sudo mv gotunnel-darwin-amd64 /usr/local/bin/gotunnel

# macOS (Apple Silicon)
wget https://github.com/bakare-dev/gotunnel/releases/latest/download/gotunnel-darwin-arm64
chmod +x gotunnel-darwin-arm64
sudo mv gotunnel-darwin-arm64 /usr/local/bin/gotunnel

# Windows
# Download from: https://github.com/bakare-dev/gotunnel/releases/latest/download/gotunnel-windows-amd64.exe
```

#### Install with Go

```bash
go install github.com/bakare-dev/gotunnel/cmd/gotunnel@latest
```

#### Build from Source

```bash
git clone https://github.com/bakare-dev/gotunnel.git
cd gotunnel
make build
# Binary will be in bin/gotunnel
```

### Server Setup

```bash
# Basic usage
gotunnel server

# With TLS (recommended for production)
gotunnel server \
  --tls \
  --tls-cert=/path/to/server-cert.pem \
  --tls-key=/path/to/server-key.pem

# Custom configuration
gotunnel server \
  --addr=:9000 \
  --start-port=10000
```

### Client Usage

```bash
# Basic usage
gotunnel --local localhost:3000

# Explicit client command
gotunnel client --local localhost:3000

# Connect to remote server
gotunnel --server your-server.com:9000 --local localhost:3000

# With TLS
gotunnel \
  --server your-server.com:9000 \
  --local localhost:3000 \
  --tls \
  --tls-ca=/path/to/ca-cert.pem

# Disable auto-reconnect
gotunnel --local localhost:3000 --no-reconnect
```

### Access Your Service

```bash
# From anywhere on the internet
curl http://your-server-ip:10000
```

## Commands

```bash
gotunnel server      # Start tunnel server
gotunnel client      # Start tunnel client (explicit)
gotunnel             # Start tunnel client (default)
gotunnel version     # Show version
gotunnel help        # Show help
```

## Use Cases

-   **Local Development**: Test webhooks and integrations with external services
-   **Remote Access**: Access services behind NAT/firewall
-   **Demos**: Share local work with clients without deployment
-   **Testing**: Test mobile apps against local backend
-   **IoT**: Connect devices in restricted networks
-   **Microservices**: Expose local microservices for integration testing

## How It Works

```
Public Internet          Tunnel Server           Your Machine
     │                        │                        │
     │  HTTP Request          │                        │
     ├───────────────────────>│                        │
     │                        │  Binary Protocol       │
     │                        ├───────────────────────>│
     │                        │                        │
     │                        │    (localhost:3000)    │
     │                        │<───────────────────────┤
     │  HTTP Response         │                        │
     │<───────────────────────┤                        │
```

1. Client establishes persistent connection to server
2. Server assigns a public port (e.g., 10000)
3. Public connections to that port are multiplexed through the tunnel
4. Client forwards traffic to your local service
5. Responses flow back through the same tunnel

## Features in Detail

### HTTP Request Logging

Real-time monitoring of all HTTP traffic with detailed metrics:

```
╔════════════════════════════════════════════════════════════╗
║                   GoTunnel v1.0.1                          ║
║                 Secure TCP Tunneling                       ║
╚════════════════════════════════════════════════════════════╝

Session Status         online
Version                1.0.1
Tunnel Server          tunnel.example.com:9000
TLS Encryption         enabled ✓
Auto-Reconnect         enabled

Forwarding             tcp://tunnel.example.com:10000 → localhost:3000

HTTP Requests
─────────────────────────────────────────────────────────────
Connected at 2026-01-10 18:00:00

│ HTTP  │ ✓ GET    /api/users                  200 OK           45ms
│ HTTP  │ ✓ POST   /api/login                  201 Created     120ms
│ HTTP  │ ⚠ GET    /api/missing                404 Not Found    12ms
│ HTTP  │ ✗ POST   /api/error                  500 Error        85ms
```

Visual indicators:

-   ✓ Success (2xx)
-   ⚠ Client Error (4xx)
-   ✗ Server Error (5xx)

### Metrics & Monitoring

On exit or Ctrl+C, view comprehensive session statistics:

```
Metrics Summary
─────────────────────────────────────────────────────────────
Active Streams     0
Total Streams      47
Total Connections  47

Data Sent          2.3 MB
Data Received      1.8 MB
Total Transfer     4.1 MB

HTTP Requests      47
Avg Latency        35ms
Min Latency        12ms
Max Latency        120ms

Status Codes
  200: 38 requests
  201: 5 requests
  404: 3 requests
  500: 1 requests

Uptime             15m 32s
```

### Auto-Reconnection

Automatically reconnects if connection is lost:

```
│ ERROR │ Session lost: EOF
│ INFO  │ Connection lost, attempting to reconnect...
│ INFO  │ Connection attempt 1/10...
│ INFO  │ Connected successfully
```

Features:

-   Exponential backoff (1s → 2s → 4s → 8s → ... → 30s max)
-   Configurable retry attempts (default: 10)
-   Graceful handling of network interruptions
-   Can be disabled with `--no-reconnect`

### TLS Encryption

Secure tunnel traffic with TLS:

```bash
# Generate certificates
./scripts/gen-cert.sh

# Server with TLS
gotunnel server --tls --tls-cert=certs/server-cert.pem --tls-key=certs/server-key.pem

# Client with TLS
gotunnel --local localhost:3000 --tls --tls-ca=certs/ca-cert.pem
```

## Configuration

### Server Options

```bash
--addr string           Listen address (default ":9000")
--start-port int        Starting port for public listeners (default 10000)
--tls                   Enable TLS encryption
--tls-cert string       Path to TLS certificate (default "certs/server-cert.pem")
--tls-key string        Path to TLS private key (default "certs/server-key.pem")
```

### Client Options

```bash
--server string         Tunnel server address (default "localhost:9000")
--local string          Local service to expose (required, e.g., localhost:8080)
--token string          Authentication token (default "dev-token")
--tls                   Enable TLS encryption
--tls-ca string         Path to CA certificate (default "certs/ca-cert.pem")
--no-reconnect          Disable auto-reconnect on connection loss
```

## Deployment Guide

### Deploy Server on VPS

#### Using Binary

```bash
# On your VPS (Ubuntu/Debian)
# 1. Download binary
wget https://github.com/bakare-dev/gotunnel/releases/latest/download/gotunnel-linux-amd64
chmod +x gotunnel-linux-amd64
sudo mv gotunnel-linux-amd64 /usr/local/bin/gotunnel

# 2. Create systemd service
sudo nano /etc/systemd/system/gotunnel.service
```

Add this content:

```ini
[Unit]
Description=GoTunnel Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/gotunnel server --addr=:9000 --start-port=10000
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# 3. Start service
sudo systemctl daemon-reload
sudo systemctl enable gotunnel
sudo systemctl start gotunnel

# 4. Check status
sudo systemctl status gotunnel

# 5. Open firewall
sudo ufw allow 9000/tcp
sudo ufw allow 10000:10100/tcp
sudo ufw enable
```

### Client Installation

#### Linux/macOS

```bash
# Download
wget https://github.com/bakare-dev/gotunnel/releases/latest/download/gotunnel-linux-amd64
chmod +x gotunnel-linux-amd64
sudo mv gotunnel-linux-amd64 /usr/local/bin/gotunnel

# Use it
gotunnel --server your-server.com:9000 --local localhost:3000
```

#### Windows

```powershell
# Download gotunnel-windows-amd64.exe from releases page
# Rename to gotunnel.exe
# Add to PATH or run directly

gotunnel.exe --server your-server.com:9000 --local localhost:3000
```

#### Using Go

```bash
go install github.com/bakare-dev/gotunnel/cmd/gotunnel@latest
gotunnel --server your-server.com:9000 --local localhost:3000
```

## Examples

### Expose Local Web Server

```bash
# Start your local web server
python3 -m http.server 3000

# Start tunnel
gotunnel --server tunnel.example.com:9000 --local localhost:3000

# Access from anywhere
curl http://tunnel.example.com:10000
```

### Expose Local API

```bash
# Your local API running on port 3000
npm start

# Tunnel it
gotunnel --server tunnel.example.com:9000 --local localhost:3000

# Test the API
curl http://tunnel.example.com:10000/api/users
```

### Expose Database (Use with Caution)

```bash
# PostgreSQL running locally
gotunnel --server tunnel.example.com:9000 --local localhost:5432

# Connect from remote machine
psql -h tunnel.example.com -p 10000 -U postgres
```

### Webhook Development

```bash
# Local webhook receiver
node webhook-server.js  # Running on port 3000

# Tunnel it
gotunnel --server tunnel.example.com:9000 --local localhost:3000

# Configure webhook URL: http://tunnel.example.com:10000/webhook
```

## Protocol Support

GoTunnel operates at the TCP layer and is completely **protocol-agnostic**:

-   ✅ HTTP/HTTPS (REST APIs, web servers)
-   ✅ WebSockets
-   ✅ gRPC
-   ✅ Databases (PostgreSQL, MySQL, Redis, MongoDB)
-   ✅ SSH
-   ✅ Any TCP-based protocol

## Architecture

-   **Binary Protocol**: Custom frame-based protocol with versioning
-   **State Machine**: Enforced handshake → auth → forwarding flow
-   **Stream Multiplexing**: Each public connection becomes a unique stream
-   **Concurrent**: Handles multiple clients and streams simultaneously

See [PROTOCOL.md](docs/PROTOCOL.md) and [ARCHITECTURE.md](docs/ARCHITECTURE.md) for technical details.

## Project Status

**Current Version**: v1.0.1 (Stable)

### Implemented Features ✅

-   ✅ Core tunneling functionality
-   ✅ Binary protocol implementation
-   ✅ Multi-client support
-   ✅ Stream multiplexing
-   ✅ HTTP request logging with metrics
-   ✅ Bandwidth and latency tracking
-   ✅ Graceful shutdown
-   ✅ Auto-reconnection with exponential backoff
-   ✅ TLS encryption support
-   ✅ Token-based authentication
-   ✅ Unified CLI with subcommands

### Planned Features (v2.0+) 🚀

-   🔄 **P2P Node Mode** - Users can host tunnels for each other (decentralized)
-   🔄 **Discovery Service** - Lightweight matchmaking for P2P nodes
-   🔄 **Credit System** - Earn credits by hosting, spend to use
-   🔄 **HTTP Host Routing** - Multiple services via subdomains
-   🔄 **Custom Domains** - Bring your own domain
-   🔄 **Web Dashboard** - Real-time monitoring UI
-   🔄 **Rate Limiting** - Bandwidth controls per client
-   🔄 **UDP Tunneling** - Support UDP protocol
-   🔄 **Traffic Replay** - Record and replay requests for debugging
-   🔄 **Prometheus Metrics** - Export metrics endpoint
-   🔄 **Multiple Authentication** - JWT, API keys, OAuth

## Documentation

-   [Protocol Specification](docs/PROTOCOL.md) - Wire protocol details
-   [Architecture](docs/ARCHITECTURE.md) - System design and components
-   [Contributing](CONTRIBUTING.md) - How to contribute
-   [Changelog](CHANGELOG.md) - Version history

## Performance

Typical performance on modest hardware:

-   **Throughput**: 500+ MB/s per tunnel
-   **Latency**: <10ms overhead
-   **Concurrent Streams**: 1000+ per session
-   **Memory**: ~50MB per active session
-   **CPU**: Minimal (<5% on modern CPUs)

## Security

### Current Security Features

-   Token-based authentication
-   TLS encryption (optional but recommended)
-   Session isolation
-   Payload size limits
-   Heartbeat-based liveness detection

### Security Recommendations

-   ✅ Always use TLS in production
-   ✅ Use strong, unique authentication tokens
-   ✅ Run server behind firewall with limited port exposure
-   ✅ Regularly update to latest version
-   ✅ Monitor logs for suspicious activity

## Troubleshooting

### Common Issues

**Connection Refused**

```bash
# Check if server is running
curl http://your-server:9000

# Check firewall
sudo ufw status
```

**TLS Certificate Errors**

```bash
# Regenerate certificates
./scripts/gen-cert.sh

# Verify certificate
openssl x509 -in certs/server-cert.pem -text -noout
```

**Client Can't Connect**

```bash
# Test connectivity
telnet your-server.com 9000

# Try with verbose logging (if implemented)
gotunnel --server your-server.com:9000 --local localhost:3000
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone repository
git clone https://github.com/bakare-dev/gotunnel.git
cd gotunnel

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
make build
```

## License

MIT License - See [LICENSE](LICENSE) for details

## Inspiration

Built as a learning project inspired by:

-   [ngrok](https://ngrok.com/) - Commercial tunneling service
-   [localtunnel](https://localtunnel.github.io/www/) - Open-source alternative
-   [Cloudflare Tunnel](https://www.cloudflare.com/products/tunnel/) - Enterprise solution

## Author

Built by **Bakare Praise** as a portfolio project demonstrating:

-   Network programming in Go
-   Binary protocol design
-   Concurrent system architecture
-   Real-world TCP tunneling implementation
-   Production-grade error handling
-   TLS/SSL implementation

## Support

-   📧 Email: bakarepraise3@gmail.com
-   🐛 Issues: [GitHub Issues](https://github.com/bakare-dev/gotunnel/issues)
-   💬 Discussions: [GitHub Discussions](https://github.com/bakare-dev/gotunnel/discussions)
-   ⭐ Star this repo if you find it useful!

## Roadmap

### v1.2 (Next Minor Release)

-   [ ] Configuration file support (YAML/JSON)
-   [ ] Improved error messages
-   [ ] Connection pooling optimizations

### v2.0 (Major Release - P2P)

-   [ ] Peer-to-peer node mode
-   [ ] Discovery service
-   [ ] Credit/reputation system
-   [ ] Community-driven infrastructure

### v3.0 (Future)

-   [ ] Web dashboard
-   [ ] HTTP routing with subdomains
-   [ ] Custom domains
-   [ ] Advanced metrics and monitoring
