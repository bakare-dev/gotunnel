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
-   Optional: **TLS 1.2+** for encryption

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
| Payload Len | 4 bytes | Payload size in bytes                    |
| Payload     | N bytes | Message payload                          |

**Total header size**: 10 bytes

---

## Protocol Versions

| Version | Meaning                  | Status    |
| ------- | ------------------------ | --------- |
| `0x01`  | Initial protocol version | ✅ Stable |

Unsupported versions result in `ErrUnsupportedProto` and immediate connection closure.

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
INIT → HANDSHAKEN → AUTHENTICATED → FORWARDING
```

### State Transitions

| State           | Allowed Incoming Messages                                                       | Next State      |
| --------------- | ------------------------------------------------------------------------------- | --------------- |
| `INIT`          | `MsgHandshake`                                                                  | `HANDSHAKEN`    |
| `HANDSHAKEN`    | `MsgAuth`                                                                       | `AUTHENTICATED` |
| `AUTHENTICATED` | `MsgBindOK`, `MsgStreamOpen`, `MsgStreamData`, `MsgStreamClose`, `MsgHeartbeat` | `FORWARDING`    |
| `FORWARDING`    | `MsgStreamData`, `MsgStreamClose`, `MsgHeartbeat`                               | `FORWARDING`    |

Any invalid transition results in a protocol error and session termination.

**Note**: Current implementation auto-sends `MsgBindOK` after authentication, so client transitions directly to FORWARDING state.

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
-   **Expose Addr**: Local service address (e.g. `localhost:3000`)

**Server → Client**: `MsgHandshakeAck`

No payload. Confirms protocol version compatibility.

**Example**:

```
Client sends:
  Version: 0x01
  Type: 0x01 (MsgHandshake)
  Stream ID: 0x00000000
  Payload Len: 0x00000015 (21 bytes)
  Payload:
    Role: 0x01 (Client)
    Capabilities: 0x0000000000000000
    Expose Addr: "localhost:3000" (15 bytes)

Server responds:
  Version: 0x01
  Type: 0x02 (MsgHandshakeAck)
  Stream ID: 0x00000000
  Payload Len: 0x00000000
  Payload: (empty)
```

---

### 2. Authentication

**Client → Server**: `MsgAuth`

Payload:

```
+----------------+
| Token (string) |
+----------------+
```

The token is sent as raw UTF-8 bytes.

**Server → Client**: `MsgAuthOK` or `MsgAuthErr`

-   `MsgAuthOK`: No payload, authentication successful
-   `MsgAuthErr`: Payload contains error message string (UTF-8)

**Example (Success)**:

```
Client sends:
  Version: 0x01
  Type: 0x03 (MsgAuth)
  Stream ID: 0x00000000
  Payload Len: 0x00000009 (9 bytes)
  Payload: "dev-token"

Server responds:
  Version: 0x01
  Type: 0x04 (MsgAuthOK)
  Stream ID: 0x00000000
  Payload Len: 0x00000000
  Payload: (empty)
```

**Example (Failure)**:

```
Server responds:
  Version: 0x01
  Type: 0x05 (MsgAuthErr)
  Stream ID: 0x00000000
  Payload Len: 0x0000000D (13 bytes)
  Payload: "Invalid token"
```

---

### 3. Bind

**Note**: In current implementation, bind happens automatically after successful authentication.

**Server → Client**: `MsgBindOK`

Payload:

```
+----------------+
| Public Port    |
| (uint16)       |
+----------------+
```

The server-assigned public port number in big-endian format.

**Example**:

```
Server sends:
  Version: 0x01
  Type: 0x07 (MsgBindOK)
  Stream ID: 0x00000000
  Payload Len: 0x00000002 (2 bytes)
  Payload: 0x2710 (port 10000 in big-endian)
```

---

### 4. Heartbeat

**Either → Either**: `MsgHeartbeat`

No payload. Sent periodically to maintain session liveness.

-   **Interval**: 10 seconds (configurable via `HeartbeatInterval`)
-   **Timeout**: 30 seconds (configurable via `HeartbeatTimeout`)
-   **Behavior**: If no frames received within timeout, session expires

**Example**:

```
Either side sends:
  Version: 0x01
  Type: 0x08 (MsgHeartbeat)
  Stream ID: 0x00000000
  Payload Len: 0x00000000
  Payload: (empty)
```

---

### 5. Stream Lifecycle

#### Stream Open

**Server → Client**: `MsgStreamOpen`

Sent when a new public TCP connection arrives at the server.

Payload: **None** (Stream ID in header identifies the stream)

**Example**:

```
Server sends:
  Version: 0x01
  Type: 0x10 (MsgStreamOpen)
  Stream ID: 0x00000001 (stream 1)
  Payload Len: 0x00000000
  Payload: (empty)
```

**Client behavior**:

1. Receives `MsgStreamOpen` with stream ID
2. Opens TCP connection to local service
3. Begins forwarding data bidirectionally

---

#### Stream Data

**Either → Either**: `MsgStreamData`

Bidirectional. Carries raw application protocol data.

Payload:

```
+-----------------+
| Raw bytes       |
| (application    |
|  protocol data) |
+-----------------+
```

**Maximum payload size**: 16MB (configurable via `MaxPayloadSize`)

**Example (HTTP Request)**:

```
Server sends:
  Version: 0x01
  Type: 0x11 (MsgStreamData)
  Stream ID: 0x00000001
  Payload Len: 0x0000004E (78 bytes)
  Payload: "GET /api/users HTTP/1.1\r\nHost: localhost\r\n..."
```

**Example (HTTP Response)**:

```
Client sends:
  Version: 0x01
  Type: 0x11 (MsgStreamData)
  Stream ID: 0x00000001
  Payload Len: 0x000001F8 (504 bytes)
  Payload: "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n..."
```

---

#### Stream Close

**Either → Either**: `MsgStreamClose`

Signals that one side has finished sending data.

Payload: **None** (Stream ID in header identifies the stream)

**Behavior**:

-   Sender closes its write side
-   Receiver should drain any remaining data and close stream
-   Both sides release stream resources

**Example**:

```
Either side sends:
  Version: 0x01
  Type: 0x12 (MsgStreamClose)
  Stream ID: 0x00000001
  Payload Len: 0x00000000
  Payload: (empty)
```

---

## Stream Multiplexing

-   Each public TCP connection maps to **one stream**
-   Streams are identified by a unique **Stream ID** (uint32)
-   Stream ID `0` is reserved for control frames
-   Stream IDs are assigned sequentially by the server (1, 2, 3, ...)
-   Multiple streams may be active concurrently
-   Stream lifecycle is independent (one stream failure doesn't affect others)
-   Maximum concurrent streams: 2^32-1 (protocol limit, typically limited by implementation)

### Stream ID Allocation

```
Server assigns stream IDs:
  First public connection  → Stream ID 1
  Second public connection → Stream ID 2
  Third public connection  → Stream ID 3
  ...
```

### Concurrent Streams Example

```
Time    Event
────────────────────────────────────────
T0      Client connects, authenticated
T1      Public conn A → Server opens Stream 1
T2      Public conn B → Server opens Stream 2
T3      Stream 1 transfers data
T4      Stream 2 transfers data
T5      Public conn C → Server opens Stream 3
T6      Stream 1 closes
T7      Stream 2, 3 still active
T8      Public conn D → Server opens Stream 4
```

---

## Usage Example

Once authenticated and bound, the tunnel is **transparent to application protocols**.

### Example: Exposing a Local HTTP API

**Scenario**: You have a local HTTP API server running on `localhost:3000`

**Step 1**: Start the client

```bash
gotunnel-client --server tunnel.example.com:9000 --local localhost:3000 --token secret123
```

**Step 2**: Client output

```
Tunnel established: tcp://tunnel.example.com:10000 -> localhost:3000
```

**Step 3**: Make a request from anywhere

```bash
curl http://tunnel.example.com:10000/api/users
```

### Data Flow (Wire Protocol)

```
1. Public client connects to tunnel.example.com:10000

2. Server → Client: MsgStreamOpen
   Version: 0x01
   Type: 0x10
   Stream ID: 0x00000001
   Payload Len: 0x00000000

3. Client opens TCP to localhost:3000

4. Public client sends HTTP request

5. Server → Client: MsgStreamData
   Version: 0x01
   Type: 0x11
   Stream ID: 0x00000001
   Payload Len: 0x0000004E
   Payload: "GET /api/users HTTP/1.1\r\n..."

6. Client forwards to localhost:3000

7. Local server responds

8. Client → Server: MsgStreamData
   Version: 0x01
   Type: 0x11
   Stream ID: 0x00000001
   Payload Len: 0x000001F8
   Payload: "HTTP/1.1 200 OK\r\n..."

9. Server forwards to public client

10. Public client receives response

11. Either side: MsgStreamClose
    Version: 0x01
    Type: 0x12
    Stream ID: 0x00000001
    Payload Len: 0x00000000
```

### Supported Protocols

The tunnel operates at the **TCP layer** and is agnostic to application protocols:

-   ✅ HTTP/HTTPS (REST APIs, web servers)
-   ✅ gRPC
-   ✅ WebSockets
-   ✅ Databases (PostgreSQL, MySQL, Redis, MongoDB)
-   ✅ SSH
-   ✅ SMTP, IMAP, POP3
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

| Code   | Meaning                      | Action               |
| ------ | ---------------------------- | -------------------- |
| `1000` | Unsupported protocol version | Close connection     |
| `1001` | Invalid state transition     | Close connection     |
| `1002` | Authentication failed        | Close connection     |
| `1003` | Payload size exceeded        | Close connection     |
| `1004` | Stream not found             | Log and ignore frame |
| `1005` | Heartbeat timeout            | Close connection     |

**Example**:

```
Server sends:
  Version: 0x01
  Type: 0x09 (MsgError)
  Stream ID: 0x00000000
  Payload Len: 0x00000015 (21 bytes)
  Payload:
    Error code: 0x03EB (1003)
    Error message: "Payload too large"
```

### Error Recovery

-   **Frame-level errors**: Log and close connection
-   **Stream-level errors**: Close affected stream only, session continues
-   **Session-level errors**: Close entire session
-   **Network errors**: Client auto-reconnects with exponential backoff

---

## Security Considerations

### Current Implementation

-   **Max payload size enforced**: 16MB per frame (prevents DoS)
-   **Handshake & auth required** before any stream operations
-   **Heartbeat ensures liveness**: Sessions expire after 30s without activity
-   **Invalid frames terminate session immediately**
-   **Token-based authentication**: Simple shared secret
-   **Write synchronization**: Mutex prevents frame corruption
-   **Session isolation**: Clients cannot access each other's streams

### Optional TLS Encryption

-   **TLS 1.2+** for end-to-end encryption
-   **Certificate-based validation** using CA certificates
-   **Perfect Forward Secrecy** (PFS) supported
-   **No encryption overhead** on application data (happens at transport layer)

### Security Recommendations

1. ✅ **Always use TLS in production**
2. ✅ **Use strong, unique authentication tokens** (min 32 characters)
3. ✅ **Rotate tokens regularly** (every 90 days)
4. ✅ **Run server behind firewall** with limited port exposure
5. ✅ **Monitor logs** for suspicious activity
6. ✅ **Keep software updated** to latest version
7. ✅ **Use HTTPS** for web services (TLS + application-level encryption)

---

## Constraints and Limits

| Constraint             | Value       | Configurable | Notes                      |
| ---------------------- | ----------- | ------------ | -------------------------- |
| Max payload size       | 16 MB       | Yes          | Per frame                  |
| Max concurrent streams | 2^32-1      | No           | Protocol limit             |
| Heartbeat interval     | 10s         | Yes          | Time between heartbeats    |
| Heartbeat timeout      | 30s (3x)    | Yes          | Session expires after this |
| Max frame size         | 16 MB + 10B | No           | Payload + header           |
| Stream ID range        | 1 - 2^32-1  | No           | 0 reserved for control     |
| Connection timeout     | 10s         | Yes          | Initial connect timeout    |
| Reconnection max wait  | 30s         | Yes          | Max backoff duration       |

---

## Extensibility

The protocol is designed for future enhancements:

### Planned Extensions (v2.0+)

-   **Compression**: Per-stream gzip/lz4 compression (Capability bit 1)
-   **HTTP routing**: Host-based routing in handshake (Capability bit 2)
-   **Subdomains**: Automatic subdomain assignment (Capability bit 2)
-   **Metrics**: Built-in observability frames (Capability bit 3)
-   **Node discovery**: P2P node discovery messages
-   **Credit system**: Credit tracking for P2P mode

### Capability Negotiation

The `Capabilities` field in `MsgHandshake` is a bitmask:

```
Bit 0: TLS support (currently unused, negotiated at transport layer)
Bit 1: Compression support (planned)
Bit 2: HTTP routing (planned)
Bit 3: Metrics support (planned)
Bit 4: P2P node mode (planned)
Bits 5-63: Reserved for future use
```

**Negotiation Process**:

1. Client sends capabilities it supports
2. Server responds with capabilities it supports
3. Connection uses intersection of both (common capabilities)
4. Unsupported features are disabled

**Example**:

```
Client capabilities: 0x0000000000000003 (TLS + Compression)
Server capabilities: 0x0000000000000001 (TLS only)
Result: 0x0000000000000001 (TLS enabled, Compression disabled)
```

---

## Wire Format Examples

### Complete Handshake Exchange

```
Client → Server (Handshake):
  Hex: 01 01 00 00 00 00 00 00 00 15 01 00 00 00 00 00 00 00 00 0F 6C 6F 63 61 6C 68 6F 73 74 3A 33 30 30 30

  Breakdown:
    Version: 0x01
    Type: 0x01 (MsgHandshake)
    Stream ID: 0x00000000
    Payload Len: 0x00000015 (21 bytes)
    Payload:
      Role: 0x01 (Client)
      Capabilities: 0x0000000000000000
      Expose Addr length: 0x000F (15)
      Expose Addr: "localhost:3000"

Server → Client (HandshakeAck):
  Hex: 01 02 00 00 00 00 00 00 00 00

  Breakdown:
    Version: 0x01
    Type: 0x02 (MsgHandshakeAck)
    Stream ID: 0x00000000
    Payload Len: 0x00000000
```

### Stream Data Frame (HTTP Request)

```
Server → Client:
  Hex: 01 11 00 00 00 01 00 00 00 4E [... 78 bytes of HTTP request ...]

  Breakdown:
    Version: 0x01
    Type: 0x11 (MsgStreamData)
    Stream ID: 0x00000001 (stream 1)
    Payload Len: 0x0000004E (78 bytes)
    Payload: "GET /api/users HTTP/1.1\r\nHost: localhost\r\n..."
```

### Stream Close Frame

```
Either side:
  Hex: 01 12 00 00 00 01 00 00 00 00

  Breakdown:
    Version: 0x01
    Type: 0x12 (MsgStreamClose)
    Stream ID: 0x00000001 (stream 1)
    Payload Len: 0x00000000
```

---

## Implementation Notes

### Frame Serialization

```go
// Writing a frame
frame := Frame{
    Version:  0x01,
    Type:     MsgStreamData,
    StreamID: streamID,
    Payload:  data,
}

// Header (10 bytes)
binary.Write(w, binary.BigEndian, frame.Version)  // 1 byte
binary.Write(w, binary.BigEndian, frame.Type)     // 1 byte
binary.Write(w, binary.BigEndian, frame.StreamID) // 4 bytes
binary.Write(w, binary.BigEndian, uint32(len(frame.Payload))) // 4 bytes

// Payload (N bytes)
w.Write(frame.Payload)
```

### Frame Deserialization

```go
// Reading a frame
var version uint8
var msgType uint8
var streamID uint32
var payloadLen uint32

binary.Read(r, binary.BigEndian, &version)
binary.Read(r, binary.BigEndian, &msgType)
binary.Read(r, binary.BigEndian, &streamID)
binary.Read(r, binary.BigEndian, &payloadLen)

// Validate
if payloadLen > MaxPayloadSize {
    return ErrPayloadTooLarge
}

// Read payload
payload := make([]byte, payloadLen)
io.ReadFull(r, payload)
```

### Write Synchronization

**Critical**: All frame writes MUST be synchronized with a mutex to prevent interleaving.

```go
type Session struct {
    writeMu sync.Mutex
    // ...
}

func (s *Session) WriteFrame(f *Frame) error {
    s.writeMu.Lock()
    defer s.writeMu.Unlock()

    return f.Encode(s.w)
}
```

---

## Protocol Evolution

### Version History

| Version | Date       | Changes                            |
| ------- | ---------- | ---------------------------------- |
| 0x01    | 2026-01-10 | Initial protocol specification     |
| 0x02    | TBD        | Planned: Compression, HTTP routing |

### Backward Compatibility

-   **Version field** allows detection of incompatible clients
-   **Capability negotiation** allows graceful feature degradation
-   **Unknown message types** are logged and ignored (forward compatibility)
-   **Protocol changes** increment version number

---

## Testing and Debugging

### Protocol Inspection

Use standard tools to inspect traffic:

```bash
# Capture tunnel traffic (without TLS)
tcpdump -i any -s 0 -X 'tcp port 9000'

# With Wireshark
wireshark -i any -f 'tcp port 9000'
```

### Frame Validation

Implementations should validate:

1. ✅ Version matches expected
2. ✅ Frame type is known
3. ✅ Payload length doesn't exceed max
4. ✅ State machine allows this message type
5. ✅ Stream ID exists (for stream messages)

---

## References

-   Inspired by ngrok protocol design
-   Follows binary protocol best practices
-   Frame format influenced by HTTP/2 and QUIC
-   Compatible with standard TCP tooling (tcpdump, Wireshark)
-   TLS integration follows Go crypto/tls best practices
