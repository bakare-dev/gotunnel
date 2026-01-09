package client

func (f *Forwarder) writeToLocal(streamID uint32, data []byte) {
	f.mu.Lock()
	conn, ok := f.conns[streamID]
	f.mu.Unlock()

	if !ok {
		return
	}

	_, _ = conn.Write(data)
}
