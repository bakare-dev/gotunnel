package client

import (
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/bakare-dev/gotunnel/internal/protocol"
	"github.com/bakare-dev/gotunnel/internal/tunnel"
)

func (f *Forwarder) pipeLocalToTunnel(streamID uint32, conn net.Conn) {
	defer func() {
		f.mu.Lock()
		_, exists := f.conns[streamID]
		if exists {
			delete(f.conns, streamID)
		}
		delete(f.httpLogs, streamID)
		f.mu.Unlock()

		if exists {
			conn.Close()
		}
	}()

	buf := make([]byte, 4096)
	isFirstPacket := true

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			} else if err == io.EOF {
				// Silent EOF
			} else {
				log.Printf("│ ERROR │ [Stream %d] Local error: %v", streamID, err)
			}
			break
		}

		if n > 0 {
			if isFirstPacket {
				isFirstPacket = false
				f.mu.Lock()
				if httpLog, ok := f.httpLogs[streamID]; ok && httpLog.Request != nil {
					httpLog.Response = tunnel.ParseHTTPResponse(buf[:n])
					httpLog.Duration = time.Since(httpLog.StartTime)

					if logStr := httpLog.String(); logStr != "" {
						log.Printf("│ HTTP  │ %s", logStr)
					}
				}
				f.mu.Unlock()
			}

			err = f.sess.WriteFrame(&protocol.Frame{
				Type:     protocol.MsgStreamData,
				StreamID: streamID,
				Payload:  buf[:n],
			})
			if err != nil {
				log.Printf("│ ERROR │ [Stream %d] Tunnel write failed: %v", streamID, err)
				break
			}
		}
	}

	_ = f.sess.WriteFrame(&protocol.Frame{
		Type:     protocol.MsgStreamClose,
		StreamID: streamID,
	})
}
