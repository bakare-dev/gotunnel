# GoTunnel Architecture

## High-Level Overview

GoTunnel is a **reverse TCP tunneling system** inspired by ngrok.

It allows a client behind NAT/firewalls to expose a local TCP service to the public internet via a server.

The tunnel is **protocol-agnostic**, operating at the TCP layer to support any application protocol (HTTP, gRPC, databases, SSH, etc.).

---

## Core Components

```
+------------+     +-------------+     +-------------+
|  Public    |     |   Tunnel    |     |   Client    |
|  TCP       | --> |   Server    | --> |   Agent     |
|  Clients   |     |             |     |             |
+------------+     +-------------+     +-------------+
                                             |
                                             v
                                       Local Service
                                       (HTTP, gRPC, etc.)
```

---

## Server Architecture

### Responsibilities

-   Accept tunnel connections from clients
-   Authenticate clients
-   Assign and expose public TCP ports
-   Forward public traffic into tunnel streams
-   Manage multiple concurrent client sessions

### Key Modules

| Module           | Responsibility                           |
| ---------------- | ---------------------------------------- |
| `PublicListener` | Accept public TCP connections            |
| `Session`        | Protocol state machine per client        |
| `StreamManager`  | Multiplex streams over tunnel connection |
| `PortAllocator`  | Assign available public ports            |
| `Router`         | Route public connections to sessions     |

---

## Client Architecture

### Responsibilities

-   Connect to tunnel server
-   Authenticate and bind
-   Declare exposed local service
-   Forward stream data to/from local application
-   Maintain session liveness

### Key Modules

| Module          | Responsibility                     |
| --------------- | ---------------------------------- |
| `Forwarder`     | TCP ↔ stream bridge                |
| `Session`       | Protocol state machine             |
| `StreamManager` | Stream lifecycle management        |
| `Connector`     | Manage connection to local service |

---

## Connection Lifecycle

1. **Server starts** listening on tunnel port (env-configured)
2. **Client connects** to tunnel port
3. **Client sends handshake** (declares local service address)
4. **Server acknowledges** handshake
5. **Client authenticates** with token
6. **Server confirms** authentication
7. **Client sends bind** request
8. **Server assigns public port** and responds with `BindOK`
9. **Tunnel becomes active**
10. **Public connections create streams**
11. **Streams forward data bidirectionally**

---

## Stream Flow

### Generic Flow

```
Public TCP Connection
       ↓
  MsgStreamOpen
       ↓
  MsgStreamData (bidirectional)
       ↓
  MsgStreamClose
```

### Example: HTTP API Request

```
curl http://server:10000/api/users
       ↓
Server receives TCP connection on port 10000
       ↓
Server creates stream_id=1
       ↓
MsgStreamOpen (stream_id=1) → Client
       ↓
Client opens connection to localhost:6001
       ↓
MsgStreamData: "GET /api/users HTTP/1.1\r\n..." → Client
       ↓
Client forwards to localhost:6001
       ↓
Local HTTP server processes request
       ↓
Local HTTP server responds: "HTTP/1.1 200 OK\r\n..."
       ↓
Client reads response from localhost:6001
       ↓
MsgStreamData: "HTTP/1.1 200 OK\r\n..." → Server
       ↓
Server forwards to public TCP connection
       ↓
curl receives normal HTTP response
```

**Key Point**: The tunnel is completely transparent to the application protocol. HTTP, gRPC, databases, and any TCP-based protocol work without modification.

---

## Concurrency Model

### Server Side

-   **Main goroutine**: Accept tunnel connections
-   **Per session goroutines**:
    -   Read loop (process incoming frames)
    -   Write loop (send outgoing frames)
    -   Heartbeat ticker
-   **Per public connection**:
    -   Stream handler (forward TCP ↔ tunnel)

### Client Side

-   **Main goroutine**: Maintain tunnel connection
-   **Per session goroutines**:
    -   Read loop (process incoming frames)
    -   Write loop (send outgoing frames)
    -   Heartbeat ticker
-   **Per stream**:
    -   Forwarder (forward tunnel ↔ local service)

### Synchronization

-   Channels for frame passing
-   Mutex-protected session state
-   Context-based cancellation for cleanup

---

## Multi-Client Support

The server supports **multiple concurrent clients**, each with:

-   Unique session ID
-   Independent public port assignment
-   Isolated stream namespace
-   Separate authentication

**Port allocation**:

```
Client A: assigned port 10000 → exposes localhost:6001
Client B: assigned port 10001 → exposes localhost:8080
Client C: assigned port 10002 → exposes localhost:3000
```

**Routing table**:

```
Port 10000 → Session A → Stream X → Client A → localhost:6001
Port 10001 → Session B → Stream Y → Client B → localhost:8080
Port 10002 → Session C → Stream Z → Client C → localhost:3000
```

---

## Failure Handling

| Failure                   | Action                           |
| ------------------------- | -------------------------------- |
| Invalid frame             | Close session immediately        |
| Missed heartbeat (3x)     | Expire session, close streams    |
| Stream error              | Close affected stream only       |
| Client disconnect         | Drop all public connections      |
| Public connection drop    | Send `MsgStreamClose` to client  |
| Local service unreachable | Close stream, return TCP RST     |
| Authentication failure    | Send `MsgAuthErr`, close session |

---

## Configuration Model

### Server

Configured via **environment variables** (suitable for Docker/Kubernetes):

```bash
TUNNEL_PORT=9000
AUTH_TOKEN=secret123
MAX_CLIENTS=100
HEARTBEAT_INTERVAL=30s
HEARTBEAT_TIMEOUT=90s
```

### Client

Configured via **CLI flags** (runtime flexibility):

```bash
gotunnel-client \
  --server tunnel.example.com:9000 \
  --local localhost:6001 \
  --token secret123
```

---

## Design Principles

-   **Single source of truth** for protocol state
-   **Explicit acknowledgments** (no implicit state changes)
-   **No shared mutable state** without locks
-   **Fail fast** on protocol violations
-   **Separation of transport and logic**
-   **Protocol-agnostic forwarding**: Tunnel operates at TCP layer, supports any application protocol (HTTP, gRPC, databases, etc.)
-   **Graceful degradation**: Stream failures don't affect session
-   **Resource isolation**: Each client session is independent

---

## Data Path

### Outbound (Public → Local)

```
Public TCP client
       ↓
Server public listener (port 10000)
       ↓
Server routing table lookup
       ↓
Server session writer
       ↓
TCP tunnel connection
       ↓
Client session reader
       ↓
Client stream manager
       ↓
Client forwarder
       ↓
Local service (localhost:6001)
```

### Inbound (Local → Public)

```
Local service (localhost:6001)
       ↓
Client forwarder
       ↓
Client session writer
       ↓
TCP tunnel connection
       ↓
Server session reader
       ↓
Server stream manager
       ↓
Server public connection
       ↓
Public TCP client
```

---

## State Management

### Session States

```
INIT → HANDSHAKEN → AUTHENTICATED → BOUND → FORWARDING
```

-   **INIT**: Connection established, awaiting handshake
-   **HANDSHAKEN**: Handshake complete, awaiting auth
-   **AUTHENTICATED**: Auth complete, awaiting bind
-   **BOUND**: Public port assigned, ready for streams
-   **FORWARDING**: Active streams in progress

### Stream States

```
IDLE → OPENING → ACTIVE → CLOSING → CLOSED
```

-   **IDLE**: Stream ID allocated but not opened
-   **OPENING**: `MsgStreamOpen` sent, awaiting local connection
-   **ACTIVE**: Bidirectional data flow
-   **CLOSING**: `MsgStreamClose` sent/received, draining buffers
-   **CLOSED**: Stream resources released

---

## Performance Considerations

### Buffering

-   Frame-level buffering (4KB default)
-   Stream-level buffering (configurable)
-   TCP socket buffering (OS-managed)

### Backpressure

-   TCP flow control applied transparently
-   Slow clients don't block other sessions
-   Slow streams don't block other streams in same session

### Resource Limits

-   Max concurrent clients: configurable (default 100)
-   Max concurrent streams per client: configurable (default 1000)
-   Max frame size: 1MB + header
-   Connection timeout: configurable (default 30s)

---

## Security Architecture

### Current Implementation

-   **Token-based authentication**: Shared secret
-   **State machine enforcement**: Prevents protocol violations
-   **Payload size limits**: Prevents DoS
-   **Session isolation**: One client can't access another's streams
-   **Heartbeat-based liveness**: Detects dead connections

### Future Enhancements

-   **TLS encryption**: Encrypt tunnel traffic
-   **JWT tokens**: Replace shared secrets
-   **Rate limiting**: Per-client bandwidth caps
-   **ACLs**: Port-based access control
-   **Audit logging**: Track all tunnel activities

---

## Monitoring and Observability

### Metrics (Planned)

-   Active sessions
-   Active streams per session
-   Bytes transferred per session/stream
-   Connection success/failure rate
-   Heartbeat miss rate
-   Protocol error count

### Logging

Current implementation logs:

-   Session lifecycle events
-   Authentication attempts
-   Port assignments
-   Stream creation/destruction
-   Protocol errors

---

## Deployment Patterns

### Single Server

```
┌─────────────────┐
│  Tunnel Server  │
│   (VPS/Cloud)   │
└─────────────────┘
         ↑
         │ (tunnel connections)
         │
    ┌────┴────┬────────┬────────┐
    │         │        │        │
 Client A  Client B Client C Client D
    │         │        │        │
 Local    Local    Local    Local
 Service  Service  Service  Service
```

### Load Balanced (Future)

```
        ┌─────────────┐
        │ Load Balancer│
        └──────┬───────┘
               │
     ┌─────────┼─────────┐
     │         │         │
┌────▼───┐ ┌──▼────┐ ┌──▼────┐
│Server 1│ │Server 2│ │Server 3│
└────────┘ └───────┘ └───────┘
```

---

## Future Enhancements

### Protocol Level

-   **TLS negotiation**: In-protocol encryption upgrade
-   **Compression**: Per-stream gzip/lz4 compression
-   **HTTP routing**: Host-based routing (subdomain support)
-   **Metrics frames**: Built-in observability

### Operational

-   **Web dashboard**: Real-time session monitoring
-   **REST API**: Programmatic session management
-   **Multi-region**: Geographic load balancing
-   **Webhooks**: Event notifications

### Features

-   **Custom domains**: Bring your own domain
-   **TCP + UDP**: Support UDP tunneling
-   **File transfer**: Built-in file sharing
-   **Replay buffer**: Capture traffic for debugging

---

## Comparison with Alternatives

| Feature             | GoTunnel | ngrok | localtunnel | Cloudflare Tunnel |
| ------------------- | -------- | ----- | ----------- | ----------------- |
| Custom protocol     | ✅       | ✅    | ❌          | ✅                |
| TCP tunneling       | ✅       | ✅    | ❌          | ✅                |
| HTTP tunneling      | ✅       | ✅    | ✅          | ✅                |
| Self-hosted         | ✅       | ❌    | ✅          | ❌                |
| Multi-client        | ✅       | ✅    | ✅          | ✅                |
| Stream multiplexing | ✅       | ✅    | ❌          | ✅                |
| Open source         | ✅       | ❌    | ✅          | ✅                |

---

## References

-   Inspired by ngrok architecture
-   Protocol design influenced by HTTP/2 and QUIC
-   Compatible with standard TCP tooling
