# Troubleshooting Guide: GoTunnel Protocol Issues

## Issue #1: "protocol: short frame header" / "handshake rejected by server"

### Symptoms

```bash
# Server output:
2026/01/09 22:23:55 session error: protocol: short frame header

# Client output:
2026/01/09 22:23:55 handshake rejected by server
exit status 1
```

### Root Cause

The issue was caused by **state validation logic inside the `ReadFrame()` method** that prevented proper bidirectional communication during the handshake phase.

#### The Problem Flow

1. **Client** sends `MsgHandshake` frame
2. **Server** receives handshake via `ReadFrame()`
    - State changes from `StateInit` → `StateHandshaken`
    - Returns the frame successfully
3. **Server** sends `MsgHandshakeAck` response
4. **Client** tries to read the ack via `ReadFrame()`
    - **Client's state is now `StateHandshaken`**
    - **`ReadFrame()` expects `MsgAuth` when in `StateHandshaken`**
    - **Receives `MsgHandshakeAck` instead → Returns `ErrAuthRequired`**
5. Connection fails

#### Original Buggy Code

```go
func (s *Session) ReadFrame() (*Frame, error) {
    frame, err := DecodeFrame(s.r)
    if err != nil {
        return nil, err
    }

    switch s.state {
    case StateInit:
        if frame.Type != MsgHandshake {
            return nil, ErrHandshakeRequired
        }
        // Process handshake...
        s.state = StateHandshaken
        return frame, nil

    case StateHandshaken:
        if frame.Type != MsgAuth {  //  This blocks MsgHandshakeAck!
            return nil, ErrAuthRequired
        }
        // Process auth...
        return frame, nil
    }
}
```

**Why This Failed:**

-   `ReadFrame()` enforced a rigid state machine that assumed the same sequence for both client and server
-   After processing a handshake, the state changed immediately, preventing the client from reading the server's response
-   The state validation was asymmetric: both sides couldn't be in `StateHandshaken` simultaneously and communicate

### Solution

**Separate frame reading from state validation.**

The fix involved:

1. **Remove state enforcement from `ReadFrame()`** - Make it a simple frame decoder
2. **Create explicit state processing methods** - `ProcessHandshake()` and `ProcessAuth()`
3. **Let the application layer handle state transitions** - Server explicitly processes frames in the correct order

#### Fixed Code Structure

```go
// Simple frame decoder - no state validation
func (s *Session) ReadFrame() (*Frame, error) {
    frame, err := DecodeFrame(s.r)
    if err != nil {
        return nil, err
    }
    s.touch()
    return frame, nil  // Just return the frame
}

// Explicit state processing methods
func (s *Session) ProcessHandshake(frame *Frame) error {
    if s.state != StateInit {
        return ErrHandshakeRequired
    }
    // Process handshake and update state
    s.state = StateHandshaken
    return nil
}

func (s *Session) ProcessAuth(frame *Frame) error {
    if s.state != StateHandshaken {
        return ErrAuthRequired
    }
    // Process auth and update state
    s.state = StateAuthenticated
    return nil
}
```

#### Server Handler Pattern

```go
// Read handshake
frame, err := sess.ReadFrame()
if frame.Type != protocol.MsgHandshake {
    return // Reject
}
sess.ProcessHandshake(frame)  // Update state

// Send ack
sess.WriteFrame(protocol.NewFrame(protocol.MsgHandshakeAck, nil))

// Read auth
frame, err = sess.ReadFrame()
if frame.Type != protocol.MsgAuth {
    return // Reject
}
sess.ProcessAuth(frame)  // Update state

// Send auth OK
sess.WriteFrame(protocol.NewFrame(protocol.MsgAuthOK, nil))
```

### Key Lessons Learned

1. **Separation of Concerns**: Frame decoding should be separate from protocol state validation
2. **Bidirectional Communication**: State machines must account for both sending and receiving
3. **Explicit State Transitions**: Application-level handlers should control state changes, not transport-level methods
4. **Testing Bidirectional Protocols**: Always test both client→server and server→client message flows

### Additional Fix: Protocol Version

During debugging, we also ensured frames always include the correct protocol version:

```go
// Helper function to create frames with correct version
func NewFrame(typ MessageType, payload []byte) *Frame {
    return &Frame{
        Version:  ProtocolVersion1,  // Always set
        Type:     typ,
        StreamID: 0,
        Payload:  payload,
    }
}

// WriteFrame ensures version is set
func (s *Session) WriteFrame(f *Frame) error {
    f.Version = ProtocolVersion1  // Defensive: ensure version is always set
    return f.Encode(s.w)
}
```

### Verification

After the fix, the handshake succeeds:

```bash
# Server output:
2026/01/09 22:40:15 client connected from 127.0.0.1:xxxxx
2026/01/09 22:40:15 waiting for handshake...
2026/01/09 22:40:15 handshake processed, sending ack
2026/01/09 22:40:15 waiting for auth...
2026/01/09 22:40:15 auth processed, sending ok
2026/01/09 22:40:15 tunnel client authenticated

# Client output:
2026/01/09 22:40:15 connected, creating session
2026/01/09 22:40:15 sending handshake (payload size: X bytes)
2026/01/09 22:40:15 handshake ack received
2026/01/09 22:40:15 sending auth...
2026/01/09 22:40:15 tunnel authenticated, forwarding traffic
```

### Prevention Tips

-   **Unit test state machines** with both valid and invalid sequences
-   **Test bidirectional flows** explicitly in integration tests
-   **Keep transport layer simple** - complex logic belongs in application layer
-   **Use explicit methods** for state transitions instead of implicit side effects
-   **Add comprehensive logging** during development to trace message flows

---

## Related Files Modified

-   `protocol/session.go` - Removed state validation from `ReadFrame()`, added `ProcessHandshake()` and `ProcessAuth()`
-   `protocol/frame.go` - Added `NewFrame()` helper to ensure version is set
-   `cmd/server/main.go` - Explicit handshake/auth processing with state methods
-   `cmd/client/main.go` - Added debug logging, uses `NewFrame()` helper

## Version

Fixed in commit: [commit-hash]
Date: January 9, 2026
