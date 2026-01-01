// Package server_test contains tests for the MCP server.
package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scottrfrancis/mcp-calculator/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id"`
	Result  any            `json:"result,omitempty"`
	Error   *JSONRPCError  `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Helper to make JSON-RPC request
func makeRequest(t *testing.T, handler http.Handler, sessionID string, req JSONRPCRequest) JSONRPCResponse {
	body, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", sessionID)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)

	var resp JSONRPCResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	return resp
}

// TestServerInitialize tests the initialize handshake.
func TestServerInitialize(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	body, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check session ID header
	sessionID := rr.Header().Get("Mcp-Session-Id")
	assert.NotEmpty(t, sessionID, "should return session ID")

	// Check response
	var resp JSONRPCResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, float64(1), resp.ID)
	assert.Nil(t, resp.Error)

	result := resp.Result.(map[string]any)
	assert.Equal(t, "2025-03-26", result["protocolVersion"])
	assert.NotNil(t, result["serverInfo"])
	assert.NotNil(t, result["capabilities"])

	serverInfo := result["serverInfo"].(map[string]any)
	assert.Equal(t, "mcp-calculator", serverInfo["name"])
}

// TestServerToolsList tests the tools/list method.
func TestServerToolsList(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	// First initialize to get session
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	}
	body, _ := json.Marshal(initReq)
	httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)
	sessionID := rr.Header().Get("Mcp-Session-Id")

	// Now list tools
	resp := makeRequest(t, handler, sessionID, JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	})

	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]any)
	tools := result["tools"].([]any)
	assert.Len(t, tools, 1)

	tool := tools[0].(map[string]any)
	assert.Equal(t, "calculate", tool["name"])
	assert.NotEmpty(t, tool["description"])
	assert.NotNil(t, tool["inputSchema"])
}

// TestServerToolsCall tests the tools/call method.
func TestServerToolsCall(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	// Initialize
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	}
	body, _ := json.Marshal(initReq)
	httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)
	sessionID := rr.Header().Get("Mcp-Session-Id")

	// Call calculate tool
	resp := makeRequest(t, handler, sessionID, JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "calculate",
			"arguments": map[string]any{
				"calculations": []map[string]any{
					{"name": "sum", "operation": "sum", "args": []any{1, 2, 3}},
				},
			},
		},
	})

	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]any)
	content := result["content"].([]any)
	assert.Len(t, content, 1)

	textContent := content[0].(map[string]any)
	assert.Equal(t, "text", textContent["type"])

	// Parse the text content as JSON
	var calcResult map[string]any
	err := json.Unmarshal([]byte(textContent["text"].(string)), &calcResult)
	require.NoError(t, err)

	assert.True(t, calcResult["success"].(bool))
	results := calcResult["results"].(map[string]any)
	assert.Equal(t, float64(6), results["sum"])
}

// TestServerUnknownMethod tests unknown method handling.
func TestServerUnknownMethod(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	// Initialize first
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	}
	body, _ := json.Marshal(initReq)
	httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)
	sessionID := rr.Header().Get("Mcp-Session-Id")

	resp := makeRequest(t, handler, sessionID, JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "unknown/method",
	})

	assert.NotNil(t, resp.Error)
	assert.Equal(t, -32601, resp.Error.Code) // Method not found
}

// TestServerUnknownTool tests unknown tool handling.
func TestServerUnknownTool(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	// Initialize first
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	}
	body, _ := json.Marshal(initReq)
	httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)
	sessionID := rr.Header().Get("Mcp-Session-Id")

	resp := makeRequest(t, handler, sessionID, JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: map[string]any{
			"name":      "unknown_tool",
			"arguments": map[string]any{},
		},
	})

	// The tool error should be in the result, not as a JSON-RPC error
	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]any)
	isError := result["isError"].(bool)
	assert.True(t, isError)
}

// TestServerSessionRequired tests that session is required after initialize.
func TestServerSessionRequired(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	// Try to list tools without session
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)

	// Should return error (no session)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestServerInvalidSession tests invalid session handling.
func TestServerInvalidSession(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	httpReq.Header.Set("Mcp-Session-Id", "invalid-session-id")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestServerMalformedJSON tests malformed JSON handling.
func TestServerMalformedJSON(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader([]byte("not valid json")))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp JSONRPCResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, -32700, resp.Error.Code) // Parse error
}

// TestServerNotificationsInitialized tests the notifications/initialized handling.
func TestServerNotificationsInitialized(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	// Initialize first
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	}
	body, _ := json.Marshal(initReq)
	httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)
	sessionID := rr.Header().Get("Mcp-Session-Id")

	// Send notification (no ID = notification)
	notifReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	body, _ = json.Marshal(notifReq)
	httpReq = httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	httpReq.Header.Set("Mcp-Session-Id", sessionID)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)

	// Notifications should return 204 No Content
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

// TestServerSessionDelete tests session deletion via DELETE.
func TestServerSessionDelete(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	// Initialize to get session
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	}
	body, _ := json.Marshal(initReq)
	httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)
	sessionID := rr.Header().Get("Mcp-Session-Id")

	// Delete session
	httpReq = httptest.NewRequest("DELETE", "/mcp", nil)
	httpReq.Header.Set("Mcp-Session-Id", sessionID)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Try to use deleted session
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	body, _ = json.Marshal(req)
	httpReq = httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	httpReq.Header.Set("Mcp-Session-Id", sessionID)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestHealthEndpoint tests the health endpoint.
func TestHealthEndpoint(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	httpReq := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var health map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &health)
	require.NoError(t, err)

	assert.Equal(t, "healthy", health["status"])
	assert.NotEmpty(t, health["version"])
}

// TestReadyEndpoint tests the ready endpoint.
func TestReadyEndpoint(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	httpReq := httptest.NewRequest("GET", "/ready", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)

	assert.Equal(t, http.StatusOK, rr.Code)

	var ready map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &ready)
	require.NoError(t, err)

	assert.True(t, ready["ready"].(bool))
}

// TestConcurrentRequests tests concurrent request handling.
func TestConcurrentRequests(t *testing.T) {
	srv := server.New()
	handler := srv.Handler()

	// Initialize
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	}
	body, _ := json.Marshal(initReq)
	httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httpReq)
	sessionID := rr.Header().Get("Mcp-Session-Id")

	// Make concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			resp := makeRequest(t, handler, sessionID, JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      id,
				Method:  "tools/call",
				Params: map[string]any{
					"name": "calculate",
					"arguments": map[string]any{
						"calculations": []map[string]any{
							{"name": "result", "operation": "sum", "args": []any{id, id}},
						},
					},
				},
			})
			assert.Nil(t, resp.Error)
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
