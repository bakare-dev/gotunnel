package server

import (
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/bakare-dev/gotunnel/internal/protocol"
	"github.com/bakare-dev/gotunnel/internal/tunnel"
)

func (p *PublicListener) handleConn(conn net.Conn) {
	defer conn.Close()

	port := ExtractLocalPort(conn)

	sess, ok := p.router.Get(port)
	if !ok {
		log.Println("no session for port", port)
		return
	}

	if sess.IsClosed() {
		return
	}

	stream := sess.Streams().Open()
	sess.Metrics.StreamOpened()

	httpLog := &tunnel.HTTPLog{
		StartTime: time.Now(),
	}
	var firstRequest []byte

	if err := sess.WriteFrame(&protocol.Frame{
		Type:     protocol.MsgStreamOpen,
		StreamID: stream.ID,
	}); err != nil {
		log.Printf("│ ERROR │ [Stream %d] Failed to send StreamOpen: %v", stream.ID, err)
		sess.Metrics.StreamClosed()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		isFirstPacket := true

		for {
			if sess.IsClosed() {
				return
			}

			n, err := conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("│ DEBUG │ [Stream %d] Public read error: %v", stream.ID, err)
				}
				break
			}

			if isFirstPacket {
				isFirstPacket = false
				firstRequest = make([]byte, n)
				copy(firstRequest, buf[:n])
				httpLog.Request = tunnel.ParseHTTPRequest(firstRequest)
			}

			if err := sess.WriteFrame(&protocol.Frame{
				Type:     protocol.MsgStreamData,
				StreamID: stream.ID,
				Payload:  buf[:n],
			}); err != nil {
				if err != protocol.ErrSessionExpired {
					log.Printf("│ ERROR │ [Stream %d] Failed to forward to tunnel: %v", stream.ID, err)
				}
				break
			}
		}

		_ = sess.WriteFrame(&protocol.Frame{
			Type:     protocol.MsgStreamClose,
			StreamID: stream.ID,
		})
	}()

	go func() {
		defer wg.Done()
		isFirstPacket := true

		for {
			select {
			case data, ok := <-stream.In:
				if !ok {
					return
				}

				if isFirstPacket && httpLog.Request != nil {
					isFirstPacket = false
					httpLog.Response = tunnel.ParseHTTPResponse(data)
					httpLog.Duration = time.Since(httpLog.StartTime)

					if httpLog.Response != nil {
						sess.Metrics.RecordHTTPRequest(httpLog.Response.StatusCode, httpLog.Duration)
					}

					if logStr := httpLog.String(); logStr != "" {
						log.Printf("│ HTTP  │ %s", logStr)
					}
				}

				if _, err := conn.Write(data); err != nil {
					log.Printf("│ ERROR │ [Stream %d] Failed to write to public: %v", stream.ID, err)
					return
				}

			case <-stream.Done():
				return
			}
		}
	}()

	wg.Wait()
	sess.Streams().Close(stream.ID)
	sess.Metrics.StreamClosed()
}
