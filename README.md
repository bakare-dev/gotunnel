# GoTunnel

A lightweight, self-hosted TCP tunneling system written in Go. Expose local services to the public internet through a secure tunnel, similar to ngrok.

## Features

-   🔒 **Custom Binary Protocol** - Efficient, versioned wire protocol
-   🔐 **Token-Based Authentication** - Secure tunnel access control
-   🌊 **Stream Multiplexing** - Multiple concurrent connections over single tunnel
-   🎯 **Protocol Agnostic** - Works with HTTP, gRPC, databases, SSH, and any TCP service
-   🔄 **Auto Port Assignment** - Server dynamically assigns public ports
-   💓 **Heartbeat Management** - Automatic session health monitoring
-   🚀 **Multi-Client Support** - Run multiple tunnels simultaneously
-   📦 **Zero Dependencies** - Pure Go implementation, single binary deployment

## Quick Start

### Server

```bash
# Start the tunnel server (listens on port 9000 by default)
go run cmd/server/main.go

# Or with custom configuration via environment variables
TUNNEL_PORT=9000 AUTH_TOKEN=your-secret-token go run cmd/server/main.go
```

### Client

```bash
# Expose a local service (e.g., HTTP server on port 6001)
go run cmd/client/main.go --local localhost:6001

# Output:
# Tunnel established: tcp://localhost:10000 -> localhost:6001
```

### Access Your Service

```bash
# From anywhere, access your local service via the tunnel
curl http://your-server-ip:10000
```

## Use Cases

-   **Local Development**: Test webhooks and integrations with external services
-   **Remote Access**: Access services behind NAT/firewall
-   **Demos**: Share local work with clients without deployment
-   **Testing**: Test mobile apps against local backend
-   **IoT**: Connect devices in restricted networks

## How It Works

```
Public Internet          Tunnel Server           Your Machine
     │                        │                        │
     │  HTTP Request          │                        │
     ├───────────────────────>│                        │
     │                        │  Binary Protocol       │
     │                        ├───────────────────────>│
     │                        │                        │
     │                        │    (localhost:6001)    │
     │                        │<───────────────────────┤
     │  HTTP Response         │                        │
     │<───────────────────────┤                        │
```

1. Client establishes persistent connection to server
2. Server assigns a public port (e.g., 10000)
3. Public connections to that port are multiplexed through the tunnel
4. Client forwards traffic to your local service
5. Responses flow back through the same tunnel

## Architecture

-   **Binary Protocol**: Custom frame-based protocol with versioning
-   **State Machine**: Enforced handshake → auth → bind → forwarding flow
-   **Stream Multiplexing**: Each public connection becomes a unique stream
-   **Concurrent**: Handles multiple clients and streams simultaneously

See [PROTOCOL.md](docs/protocol.md) and [ARCHITECTURE.md](docs/architecture.md) for details.

## Configuration

### Server (Environment Variables)

```bash
TUNNEL_PORT=9000           # Port for tunnel connections (default: 9000)
AUTH_TOKEN=secret123       # Authentication token (optional)
```

### Client (CLI Flags)

```bash
--server    # Tunnel server address (default: localhost:9000)
--local     # Local service to expose (required, e.g., localhost:8080)
--token     # Authentication token (optional)
```

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/bakare-dev/gotunnel.git
cd gotunnel

# Build server
go build -o bin/gotunnel-server cmd/server/main.go

# Build client
go build -o bin/gotunnel-client cmd/client/main.go

# Run
./bin/gotunnel-server
./bin/gotunnel-client --local localhost:8080
```

### Using Docker (Coming Soon)

```bash
# Server
docker run -p 9000:9000 gotunnel/server

# Client
docker run gotunnel/client --server server:9000 --local host.docker.internal:8080
```

## Examples

### Expose Local Web Server

```bash
# Start your local web server
python3 -m http.server 6001

# Start tunnel client
go run cmd/client/main.go --local localhost:6001

# Access from anywhere
curl http://your-server-ip:10000
```

### Expose Local API

```bash
# Your local API running on port 3000
npm start

# Tunnel it
go run cmd/client/main.go --local localhost:3000

# Test the API
curl http://your-server-ip:10000/api/users
```

### Expose Database (Use with Caution)

```bash
# PostgreSQL running locally
go run cmd/client/main.go --local localhost:5432

# Connect from remote machine
psql -h your-server-ip -p 10000 -U postgres
```

## Protocol Support

GoTunnel operates at the TCP layer and is completely **protocol-agnostic**:

-   ✅ HTTP/HTTPS (REST APIs, web servers)
-   ✅ WebSockets
-   ✅ gRPC
-   ✅ Databases (PostgreSQL, MySQL, Redis, MongoDB)
-   ✅ SSH
-   ✅ Any TCP-based protocol

## Project Status

**Current Version**: v0.1.0 (Alpha)

✅ Core tunneling functionality working  
✅ Binary protocol implementation  
✅ Multi-client support  
✅ Stream multiplexing  
🔄 TLS encryption (planned)  
🔄 HTTP host-based routing (planned)  
🔄 Web dashboard (planned)

## Documentation

-   [Protocol Specification](docs/protocol.md) - Wire protocol details
-   [Architecture](docs/architecture.md) - System design and components
-   [Development Guide](docs/DEVELOPMENT.md) - Contributing guidelines (coming soon)

## Roadmap

-   [ ] TLS encryption for tunnel traffic
-   [ ] Custom domain support
-   [ ] HTTP/HTTPS routing with subdomains
-   [ ] Web dashboard for monitoring
-   [ ] Client auto-reconnection
-   [ ] Rate limiting and bandwidth controls
-   [ ] Docker images
-   [ ] UDP tunneling support

## Contributing

Contributions are welcome! This is a portfolio/learning project, but PRs for bug fixes and improvements are appreciated.

## License

MIT License - See [LICENSE](LICENSE) for details

## Inspiration

Built as a learning project inspired by:

-   [ngrok](https://ngrok.com/)
-   [localtunnel](https://localtunnel.github.io/www/)
-   [Cloudflare Tunnel](https://www.cloudflare.com/products/tunnel/)

## Author

Built by [Your Name] as a portfolio project demonstrating:

-   Network programming in Go
-   Binary protocol design
-   Concurrent system architecture
-   Real-world TCP tunneling implementation
