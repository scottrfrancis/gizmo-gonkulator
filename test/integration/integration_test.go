// Package integration_test contains integration tests for the full MCP server stack.
package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/scottrfrancis/mcp-calculator/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullMCPFlow tests the complete MCP interaction flow.
func TestFullMCPFlow(t *testing.T) {
	srv := server.New()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	// Step 1: Initialize
	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo": map[string]any{
				"name":    "integration-test",
				"version": "1.0.0",
			},
		},
	}
	body, _ := json.Marshal(initReq)

	req, _ := http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	sessionID := resp.Header.Get("Mcp-Session-Id")
	require.NotEmpty(t, sessionID, "should receive session ID")

	var initResp map[string]any
	err = json.NewDecoder(resp.Body).Decode(&initResp)
	require.NoError(t, err)
	assert.Equal(t, "2.0", initResp["jsonrpc"])

	// Step 2: Send initialized notification
	notifReq := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	body, _ = json.Marshal(notifReq)

	req, _ = http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Mcp-Session-Id", sessionID)

	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Step 3: List tools
	listReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	}
	body, _ = json.Marshal(listReq)

	req, _ = http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Mcp-Session-Id", sessionID)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var listResp map[string]any
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(t, err)

	result := listResp["result"].(map[string]any)
	tools := result["tools"].([]any)
	assert.Len(t, tools, 1)
	assert.Equal(t, "calculate", tools[0].(map[string]any)["name"])

	// Step 4: Call calculate tool
	callReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "calculate",
			"arguments": map[string]any{
				"calculations": []map[string]any{
					{"name": "total", "operation": "sum", "args": []any{100, 200, 300}},
					{"name": "avg", "operation": "average", "args": []any{100, 200, 300}},
					{"name": "growth", "operation": "percentage", "args": []any{"total", 500}},
				},
			},
		},
	}
	body, _ = json.Marshal(callReq)

	req, _ = http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Mcp-Session-Id", sessionID)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var callResp map[string]any
	err = json.NewDecoder(resp.Body).Decode(&callResp)
	require.NoError(t, err)

	callResult := callResp["result"].(map[string]any)
	content := callResult["content"].([]any)
	textContent := content[0].(map[string]any)

	var calcResult map[string]any
	err = json.Unmarshal([]byte(textContent["text"].(string)), &calcResult)
	require.NoError(t, err)

	assert.True(t, calcResult["success"].(bool))
	results := calcResult["results"].(map[string]any)
	assert.Equal(t, float64(600), results["total"])
	assert.Equal(t, float64(200), results["avg"])
	assert.Equal(t, float64(20), results["growth"]) // (600-500)/500 * 100 = 20%

	// Step 5: Delete session
	req, _ = http.NewRequest("DELETE", ts.URL+"/mcp", nil)
	req.Header.Set("Mcp-Session-Id", sessionID)

	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestMultipleSessions tests multiple concurrent sessions.
func TestMultipleSessions(t *testing.T) {
	srv := server.New()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	// Create multiple sessions
	sessions := make([]string, 5)
	for i := 0; i < 5; i++ {
		initReq := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "initialize",
			"params": map[string]any{
				"protocolVersion": "2025-03-26",
				"clientInfo": map[string]any{
					"name":    "client-" + string(rune('A'+i)),
					"version": "1.0.0",
				},
			},
		}
		body, _ := json.Marshal(initReq)

		req, _ := http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")

		resp, err := client.Do(req)
		require.NoError(t, err)
		sessions[i] = resp.Header.Get("Mcp-Session-Id")
		resp.Body.Close()
		require.NotEmpty(t, sessions[i])
	}

	// Verify all sessions are unique
	sessionSet := make(map[string]bool)
	for _, s := range sessions {
		assert.False(t, sessionSet[s], "sessions should be unique")
		sessionSet[s] = true
	}

	// Use each session
	for i, sessionID := range sessions {
		callReq := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "tools/call",
			"params": map[string]any{
				"name": "calculate",
				"arguments": map[string]any{
					"calculations": []map[string]any{
						{"name": "result", "operation": "sum", "args": []any{i, i}},
					},
				},
			},
		}
		body, _ := json.Marshal(callReq)

		req, _ := http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("Mcp-Session-Id", sessionID)

		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	// Cleanup sessions
	for _, sessionID := range sessions {
		req, _ := http.NewRequest("DELETE", ts.URL+"/mcp", nil)
		req.Header.Set("Mcp-Session-Id", sessionID)
		resp, _ := client.Do(req)
		resp.Body.Close()
	}
}

// TestHealthAndReady tests health and readiness endpoints.
func TestHealthAndReady(t *testing.T) {
	srv := server.New()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	// Test health
	resp, err := client.Get(ts.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var health map[string]any
	json.NewDecoder(resp.Body).Decode(&health)
	assert.Equal(t, "healthy", health["status"])

	// Test ready
	resp, err = client.Get(ts.URL + "/ready")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var ready map[string]any
	json.NewDecoder(resp.Body).Decode(&ready)
	assert.True(t, ready["ready"].(bool))
}

// TestRealWorldCalculation tests real-world calculation scenarios.
func TestRealWorldCalculation(t *testing.T) {
	srv := server.New()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	// Initialize
	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	}
	body, _ := json.Marshal(initReq)
	req, _ := http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, _ := client.Do(req)
	sessionID := resp.Header.Get("Mcp-Session-Id")
	resp.Body.Close()

	// Healthcare scenario: per-day rate comparison
	callReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "calculate",
			"arguments": map[string]any{
				"calculations": []map[string]any{
					{"name": "oct_per_day", "operation": "divide", "args": []any{2561276, 8}},
					{"name": "sep_per_day", "operation": "divide", "args": []any{8782334, 21}},
					{"name": "pct_change", "operation": "percentage", "args": []any{"oct_per_day", "sep_per_day"}},
					{"name": "is_declining", "operation": "compare", "args": []any{"pct_change", 0, "<"}},
				},
			},
		},
	}
	body, _ = json.Marshal(callReq)
	req, _ = http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Mcp-Session-Id", sessionID)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var callResp map[string]any
	json.NewDecoder(resp.Body).Decode(&callResp)

	callResult := callResp["result"].(map[string]any)
	content := callResult["content"].([]any)
	textContent := content[0].(map[string]any)

	var calcResult map[string]any
	json.Unmarshal([]byte(textContent["text"].(string)), &calcResult)

	results := calcResult["results"].(map[string]any)

	// October rate should be ~$320,159.50/day
	octRate := results["oct_per_day"].(float64)
	assert.InDelta(t, 320159.5, octRate, 0.5)

	// September rate should be ~$418,206.38/day
	sepRate := results["sep_per_day"].(float64)
	assert.InDelta(t, 418206.38, sepRate, 0.5)

	// Percentage change should be negative (declining)
	pctChange := results["pct_change"].(float64)
	assert.True(t, pctChange < 0, "should show decline")
	assert.InDelta(t, -23.44, pctChange, 0.5) // Approximately -23.44%

	// Is declining should be true
	assert.True(t, results["is_declining"].(bool))
}

// TestPrecisionInIntegration tests precision through the full stack.
func TestPrecisionInIntegration(t *testing.T) {
	srv := server.New()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	// Initialize
	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	}
	body, _ := json.Marshal(initReq)
	req, _ := http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, _ := client.Do(req)
	sessionID := resp.Header.Get("Mcp-Session-Id")
	resp.Body.Close()

	// Test classic floating point error: 0.1 + 0.2
	callReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "calculate",
			"arguments": map[string]any{
				"calculations": []map[string]any{
					{"name": "sum", "operation": "add", "args": []any{0.1, 0.2}},
				},
			},
		},
	}
	body, _ = json.Marshal(callReq)
	req, _ = http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Mcp-Session-Id", sessionID)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var callResp map[string]any
	json.NewDecoder(resp.Body).Decode(&callResp)

	callResult := callResp["result"].(map[string]any)
	content := callResult["content"].([]any)
	textContent := content[0].(map[string]any)

	var calcResult map[string]any
	json.Unmarshal([]byte(textContent["text"].(string)), &calcResult)

	results := calcResult["results"].(map[string]any)
	// Should be exactly 0.3, not 0.30000000000000004
	assert.Equal(t, 0.3, results["sum"])
}

// TestErrorHandlingIntegration tests error handling through the full stack.
func TestErrorHandlingIntegration(t *testing.T) {
	srv := server.New()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	// Initialize
	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	}
	body, _ := json.Marshal(initReq)
	req, _ := http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, _ := client.Do(req)
	sessionID := resp.Header.Get("Mcp-Session-Id")
	resp.Body.Close()

	// Test error cases in batch
	callReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "calculate",
			"arguments": map[string]any{
				"calculations": []map[string]any{
					{"name": "good", "operation": "sum", "args": []any{1, 2, 3}},
					{"name": "div_zero", "operation": "divide", "args": []any{100, 0}},
					{"name": "also_good", "operation": "multiply", "args": []any{5, 5}},
				},
			},
		},
	}
	body, _ = json.Marshal(callReq)
	req, _ = http.NewRequest("POST", ts.URL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Mcp-Session-Id", sessionID)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var callResp map[string]any
	json.NewDecoder(resp.Body).Decode(&callResp)

	callResult := callResp["result"].(map[string]any)
	content := callResult["content"].([]any)
	textContent := content[0].(map[string]any)

	var calcResult map[string]any
	json.Unmarshal([]byte(textContent["text"].(string)), &calcResult)

	// Batch should succeed overall
	assert.True(t, calcResult["success"].(bool))

	results := calcResult["results"].(map[string]any)
	// Good results should be present
	assert.Equal(t, float64(6), results["good"])
	assert.Equal(t, float64(25), results["also_good"])

	// Error should be captured for div_zero
	divZeroResult := results["div_zero"].(map[string]any)
	assert.Contains(t, divZeroResult["error"], "division by zero")
}
