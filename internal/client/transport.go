package client

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/joshiste/sma_chg_log/internal/log"
)

// loggingTransport wraps an http.RoundTripper and logs requests/responses at trace level
type loggingTransport struct {
	transport http.RoundTripper
}

// newLoggingTransport creates a new logging transport wrapper
func newLoggingTransport(transport http.RoundTripper) *loggingTransport {
	return &loggingTransport{
		transport: transport,
	}
}

// RoundTrip implements http.RoundTripper
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !slog.Default().Enabled(context.Background(), log.LevelTrace) {
		return t.transport.RoundTrip(req)
	}

	// Read and log request body
	var reqBody []byte
	if req.Body != nil {
		var err error
		reqBody, err = io.ReadAll(req.Body)
		if err != nil {
			reqBody = []byte("[error reading request body]")
		}
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}

	t.logRequest(req, reqBody)

	// Execute request
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Read response body for logging, then restore it
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		respBody = []byte("[error reading response body]")
	}
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	t.logResponse(resp, respBody)

	return resp, nil
}

func (t *loggingTransport) logRequest(req *http.Request, body []byte) {
	headers := make(map[string]string)
	for key, values := range req.Header {
		if key == "Authorization" {
			headers[key] = "[REDACTED]"
		} else {
			headers[key] = values[0]
		}
	}

	slog.Log(req.Context(), log.LevelTrace, "request",
		"method", req.Method,
		"path", req.URL.Path,
		"url", req.URL.String(),
		"headers", headers,
		"payload", string(body),
	)
}

func (t *loggingTransport) logResponse(resp *http.Response, body []byte) {
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = values[0]
	}

	slog.Log(resp.Request.Context(), log.LevelTrace, "response",
		"status", resp.StatusCode,
		"path", resp.Request.URL.Path,
		"headers", headers,
		"payload", string(body),
	)
}
