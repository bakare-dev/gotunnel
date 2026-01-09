# GoTunnel Protocol Specification

## Overview

GoTunnel uses a **binary, framed, stateful protocol** over TCP to establish a secure tunnel between a **client** and a **server**, allowing the server to forward public TCP connections to a client-side local service.

The protocol is:

-   Versioned
-   Stateful
-   Stream-multiplexed
-   Extensible
-   **Protocol-agnostic** (supports HTTP, gRPC, databases, SSH, etc.)

---

## Transport

-   Transport: **TCP**
-   Encoding: **Binary**
-   Byte order: **Big Endian**

---

## Frame Format

All messages are transmitted as frames.
```
+---------+--------+-----------+------------+-------------+
| Version | Type   | Stream ID | Payload Len| Payload     |
| 1 byte  | 1 byte | 4 bytes   | 4 bytes    | N bytes     |
+---------+--------+-----------+------------+-------------+
```

### Fields

| Field       | Size    | Description                              |
| ----------- | ------- | ---------------------------------------- |
| Version     | 1 byte  | Protocol version                         |
| Type        | 1 byte  | Message type                             |
| Stream ID   | 4 bytes | Stream identifier (0 for control frames) |
| Payload Len | 4 bytes | Payload size                             |
| Payload     | N bytes | Message payload                          |

---

## Protocol Versions

| Version | Meaning                  |
| ------- | ------------------------ |
| `0x01`  | Initial protocol version |

Unsupported versions result in `ErrUnsupportedProto`.

---

## Message Types

### Control Messages

| Type              | Value  | Description                     |
| ----------------- | ------ | ------------------------------- |
| `MsgHandshake`    | `0x01` | Client handshake request        |
| `MsgHandshakeAck` | `0x02` | Server handshake acknowledgment |
| `MsgAuth`         | `0x03` | Client authentication           |
| `MsgAuthOK`       | `0x04` | Authentication success          |
| `MsgAuthErr`      | `0x05` | Authentication failure          |
| `MsgBind`         | `0x06` | Client bind request             |
| `MsgBindOK`       | `0x07` | Server bind acknowledgment      |
| `MsgHeartbeat`    | `0x08` | Keepalive ping                  |
| `MsgError`        | `0x09` | Protocol error                  |

---

### Stream Messages

| Type             | Value  | Description       |
| ---------------- | ------ | ----------------- |
| `MsgStreamOpen`  | `0x10` | Open a new stream |
| `MsgStreamData`  | `0x11` | Stream data       |
| `MsgStreamClose` | `0x12` | Close a stream    |

---

## Session State Machine
```
INIT → HANDSHAKEN → AUTHENTICATED → BOUND → FORWARDING
```

### State Transitions

| State           | Allowed Incoming Messages    | Next State      |
| --------------- | ---------------------------- | --------------- |
| `INIT`          | `MsgHandshake`               | `HANDSHAKEN`    |
| `HANDSHAKEN`    | `MsgAuth`                    | `AUTHENTICATED` |
| `AUTHENTICATED` | `MsgBind`                    | `BOUND`         |
| `BOUND`         | `MsgStreamOpen`, `MsgStreamData`, `MsgStreamClose`, `MsgHeartbeat` | `FORWARDING` |
| `FORWARDING`    | `MsgStreamData`, `MsgStreamClose`, `MsgHeartbeat` | `FORWARDING` |

Any invalid transition results in a protocol error and session termination.

---

## Protocol Flow

### 1. Handshake

**Client → Server**: `MsgHandshake`

Payload structure:
```
+--------+---------------+-------------+
| Role   | Capabilities  | Expose Addr |
| 1 byte | 8 bytes       | var length  |
+--------+---------------+-------------+
```

-   **Role**: Client (0x01) or Server (0x02)
-   **Capabilities**: Feature bitmask (reserved for future use)
-   **Expose Addr**: Local service address (e.g. `localhost:6001`)

**Server → Client**: `MsgHandshakeAck`

No payload. Confirms protocol version compatibility.

---

### 2. Authentication

**Client → Server**: `MsgAuth`

Payload:
```
+----------------+
| Token (string) |
+----------------+
```

**Server → Client**: `MsgAuthOK` or `MsgAuthErr`

-   `MsgAuthOK`: No payload, authentication successful
-   `MsgAuthErr`: Payload contains error message string

---

### 3. Bind

**Client → Server**: `MsgBind`

Payload:
```
+-----------------+
| Expose Addr     |
| (string)        |
+-----------------+
```

The address the client wants to expose (e.g., `localhost:6001`).

**Server → Client**: `MsgBindOK`

Payload:
```
+----------------+
| Public Port    |
| (uint16)       |
+----------------+
```

The server-assigned public port number.

---

### 4. Heartbeat

**Either → Either**: `MsgHeartbeat`

No payload. Sent periodically to maintain session liveness.

Expected interval: configurable (default 30s)

---

### 5. Stream Lifecycle

#### Stream Open

**Server → Client**: `MsgStreamOpen`

Sent when a new public TCP connection arrives.

Payload:
```
+------------+
| Stream ID  |
| (uint32)   |
+------------+
```

#### Stream Data

**Either → Either**: `MsgStreamData`

Bidirectional. Carries application protocol data.

Payload:
```
+-----------------+
| Raw bytes       |
| (application    |
|  protocol data) |
+-----------------+
```

#### Stream Close

**Either → Either**: `MsgStreamClose`

Payload:
```
+------------+
| Stream ID  |
| (uint32)   |
+------------+
```

---

## Stream Multiplexing

-   Each public TCP connection maps to **one stream**
-   Streams are identified by a unique **Stream ID** (uint32)
-   Stream ID `0` is reserved for control frames
-   Multiple streams may be active concurrently
-   Stream lifecycle is independent
-   Maximum concurrent streams: implementation-defined (recommended: 1000)

---

## Usage Example

Once authenticated and bound, the tunnel is **transparent to application protocols**.

### Example: Exposing a Local HTTP API

**Scenario**: You have a local HTTP API server running on `localhost:6001`

**Step 1**: Start the client
```bash
gotunnel-client --server tunnel.example.com:9000 --local localhost:6001 --token secret123
```

**Step 2**: Client output
```
Tunnel established: tcp://tunnel.example.com:10000 -> localhost:6001
```

**Step 3**: Make a request from anywhere
```bash
curl http://tunnel.example.com:10000/api/users
```

### Data Flow
```
1. Public client connects to tunnel.example.com:10000
   ↓
2. Server wraps connection in MsgStreamOpen (stream_id=1)
   ↓
3. Client receives MsgStreamOpen
   ↓
4. Client opens TCP connection to localhost:6001
   ↓
5. Public client sends: "GET /api/users HTTP/1.1\r\n..."
   ↓
6. Server wraps in MsgStreamData (stream_id=1) → Client
   ↓
7. Client unwraps and writes to localhost:6001
   ↓
8. Local HTTP server processes request
   ↓
9. Local HTTP server responds: "HTTP/1.1 200 OK\r\n..."
   ↓
10. Client wraps in MsgStreamData (stream_id=1) → Server
    ↓
11. Server unwraps and writes to public TCP connection
    ↓
12. Public client receives normal HTTP response
```

### Supported Protocols

The tunnel operates at the **TCP layer** and is agnostic to application protocols:

-   ✅ HTTP/HTTPS (REST APIs, web servers)
-   ✅ gRPC
-   ✅ WebSockets
-   ✅ Databases (PostgreSQL, MySQL, Redis)
-   ✅ SSH
-   ✅ Any TCP-based protocol

---

## Error Handling

### Protocol Errors

**Either → Either**: `MsgError`

Payload:
```
+------------------+
| Error code       |
| (uint16)         |
+------------------+
| Error message    |
| (string)         |
+------------------+
```

Common error codes:

| Code   | Meaning                        |
| ------ | ------------------------------ |
| `1000` | Unsupported protocol version   |
| `1001` | Invalid state transition       |
| `1002` | Authentication failed          |
| `1003` | Payload size exceeded          |
| `1004` | Stream not found               |
| `1005` | Heartbeat timeout              |

---

## Security Considerations

-   **Max payload size enforced**: 1MB per frame (configurable)
-   **Handshake & auth required** before any stream operations
-   **Heartbeat ensures liveness**: Sessions expire after missed heartbeats
-   **Invalid frames terminate session immediately**
-   **Token-based authentication**: Simple shared secret (upgrade to JWT/TLS recommended)
-   **No data encryption by default**: Use TLS wrapper for production

---

## Constraints and Limits

| Constraint              | Value         | Configurable |
| ----------------------- | ------------- | ------------ |
| Max payload size        | 1 MB          | Yes          |
| Max concurrent streams  | 1000          | Yes          |
| Heartbeat interval      | 30s           | Yes          |
| Heartbeat timeout       | 90s (3x)      | Yes          |
| Max frame size          | 1 MB + header | No           |
| Stream ID range         | 1 - 2^32-1    | No           |

---

## Extensibility

The protocol is designed for future enhancements:

### Planned Extensions

-   **TLS negotiation**: In-protocol TLS upgrade
-   **Compression**: Per-stream compression flags
-   **HTTP routing**: Host-based routing in handshake
-   **Subdomains**: Automatic subdomain assignment
-   **Rate limiting**: Per-stream bandwidth control
-   **Metrics**: Built-in observability frames

### Capability Negotiation

The `Capabilities` field in `MsgHandshake` is a bitmask:
```
Bit 0: TLS support
Bit 1: Compression support
Bit 2: HTTP routing
Bit 3: Metrics support
Bits 4-63: Reserved
```

Both client and server exchange capabilities. Unsupported features are disabled.

---

## Wire Format Examples

### Handshake Frame
```
Version: 0x01
Type: 0x01 (MsgHandshake)
Stream ID: 0x00000000
Payload Len: 0x00000015 (21 bytes)
Payload:
  Role: 0x01 (Client)
  Capabilities: 0x0000000000000000
  Expose Addr: "localhost:6001" (14 bytes)
```

### Stream Data Frame
```
Version: 0x01
Type: 0x11 (MsgStreamData)
Stream ID: 0x00000001
Payload Len: 0x00000100 (256 bytes)
Payload:
  [256 bytes of HTTP request data]
```

---

## Version History

| Version | Date       | Changes                          |
| ------- | ---------- | -------------------------------- |
| 0x01    | 2026-01-10 | Initial protocol specification   |

---

## References

-   Inspired by ngrok protocol design
-   Follows binary protocol best practices
-   Compatible with standard TCP tooling (tcpdump, Wireshark)