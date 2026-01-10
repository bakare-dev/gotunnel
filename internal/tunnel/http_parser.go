package tunnel

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"time"
)

type HTTPRequest struct {
	Method string
	Path   string
	Host   string
}

type HTTPResponse struct {
	StatusCode int
	Status     string
}

type HTTPLog struct {
	Request   *HTTPRequest
	Response  *HTTPResponse
	StartTime time.Time
	Duration  time.Duration
}

func ParseHTTPRequest(data []byte) *HTTPRequest {
	reader := bufio.NewReader(bytes.NewReader(data))
	req, err := http.ReadRequest(reader)
	if err != nil {
		return nil
	}

	return &HTTPRequest{
		Method: req.Method,
		Path:   req.URL.Path,
		Host:   req.Host,
	}
}

func ParseHTTPResponse(data []byte) *HTTPResponse {
	reader := bufio.NewReader(bytes.NewReader(data))
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		return nil
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}
}

func (h *HTTPLog) String() string {
	if h.Request == nil {
		return ""
	}

	method := h.Request.Method
	path := h.Request.Path
	if path == "" {
		path = "/"
	}

	if h.Response == nil {
		return fmt.Sprintf("%-6s %-40s ...", method, path)
	}

	duration := h.Duration.Milliseconds()
	statusCode := h.Response.StatusCode

	// Color-code status
	var statusColor string
	if statusCode >= 200 && statusCode < 300 {
		statusColor = "✓" // Success
	} else if statusCode >= 400 && statusCode < 500 {
		statusColor = "⚠" // Client error
	} else if statusCode >= 500 {
		statusColor = "✗" // Server error
	} else {
		statusColor = "•"
	}

	return fmt.Sprintf("%s %-6s %-40s %3d %-15s %4dms",
		statusColor,
		method,
		path,
		statusCode,
		http.StatusText(statusCode),
		duration,
	)
}
