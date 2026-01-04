// Package server implements the MCP server with Streamable HTTP transport.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/scottrfrancis/mcp-calculator/internal/auth"
	"github.com/scottrfrancis/mcp-calculator/internal/calculator"
	"github.com/scottrfrancis/mcp-calculator/internal/middleware"
	"github.com/scottrfrancis/mcp-calculator/internal/session"
)

const (
	// ServerName is the MCP server name.
	ServerName = "mcp-calculator"
	// ServerVersion is the MCP server version.
	ServerVersion = "1.0.0"
	// ProtocolVersion is the supported MCP protocol version.
	ProtocolVersion = "2025-03-26"
)

// cachedToolSchema is the pre-built tool schema for tools/list responses.
// Built once at package init to avoid rebuilding on every request.
var cachedToolSchema = []map[string]any{
	{
		"name": "calculate",
		"description": "Perform precise arithmetic calculations using Decimal precision. " +
			"Use this tool for ALL math operations - never calculate in your response. " +
			"Supports batch operations: add, subtract, multiply, divide, sum, average, " +
			"percentage, round, min, max, median, stddev, compare, abs, ceil, floor, " +
			"roi, compound_interest, present_value. " +
			"Results can reference previous calculations by name.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"calculations": map[string]any{
					"type":        "array",
					"description": "List of calculations to perform",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name": map[string]any{
								"type":        "string",
								"description": "Label for the result (can be referenced by later calculations)",
							},
							"operation": map[string]any{
								"type": "string",
								"enum": []string{
									"add", "subtract", "multiply", "divide",
									"sum", "average", "percentage", "round",
									"min", "max", "median", "stddev",
									"compare", "abs", "ceil", "floor",
									"roi", "compound_interest", "present_value",
								},
								"description": "Operation to perform",
							},
							"args": map[string]any{
								"type": "array",
								"items": map[string]any{
									"oneOf": []map[string]any{
										{"type": "number"},
										{"type": "string", "description": "Reference to previous result by name"},
									},
								},
								"description": "Numeric arguments or references to previous results",
							},
						},
						"required": []string{"operation", "args"},
					},
				},
			},
			"required": []string{"calculations"},
		},
	},
}

// JSON-RPC error codes.
const (
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternalError  = -32603
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
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Server is the MCP server.
type Server struct {
	sessions      session.Store
	rateLimiter   *middleware.RateLimiter
	authValidator auth.Validator
	authScopes    []string
	maxBodySize   int64
	metrics       *middleware.Metrics
	logger        *slog.Logger
	startTime     time.Time
	cancelCleanup context.CancelFunc
}

// Config holds server configuration.
type Config struct {
	SessionTimeout    time.Duration
	MaxSessions       int
	RateLimit         int
	RateLimitBurst    int
	EnableRateLimiter bool
	// Auth configuration
	AuthEnabled    bool
	AuthIssuer     string
	AuthAudience   string
	AuthJWKSURL    string
	AuthScopes     []string
	AuthAlgorithms []string
	APIKey         string // Static API key (simpler than OAuth)
	// Request limits
	MaxBodySize int64
	// Metrics
	EnableMetrics bool
}

// DefaultConfig returns default server configuration.
func DefaultConfig() Config {
	return Config{
		SessionTimeout:    10 * time.Minute,
		MaxSessions:       10000,
		RateLimit:         60,
		RateLimitBurst:    10,
		EnableRateLimiter: true,
		AuthEnabled:       false,
		AuthAlgorithms:    []string{"RS256", "ES256"},
		MaxBodySize:       1 << 20, // 1MB default
		EnableMetrics:     true,
	}
}

// New creates a new MCP server with default configuration.
func New() *Server {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig creates a new MCP server with custom configuration.
func NewWithConfig(cfg Config) *Server {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	sessionStore := session.NewMemoryStore(session.MemoryStoreOptions{
		Timeout: cfg.SessionTimeout,
		MaxSize: cfg.MaxSessions,
	})

	// Start session cleanup with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	sessionStore.StartCleanup(ctx, time.Minute)

	var rateLimiter *middleware.RateLimiter
	if cfg.EnableRateLimiter {
		rateLimiter = middleware.NewRateLimiter(middleware.RateLimiterConfig{
			RequestsPerMinute: cfg.RateLimit,
			BurstSize:         cfg.RateLimitBurst,
		})
		// Start rate limiter cleanup with the same context
		rateLimiter.StartCleanup(ctx, time.Minute)
	}

	// Create auth validator
	var authValidator auth.Validator
	if cfg.APIKey != "" {
		// API key auth takes precedence (simpler than OAuth)
		authValidator = auth.NewAPIKeyValidator(cfg.APIKey)
		logger.Info("API key authentication enabled")
	} else if cfg.AuthEnabled {
		authValidator = auth.NewJWTValidator(auth.JWTValidatorConfig{
			Issuer:     cfg.AuthIssuer,
			Audience:   cfg.AuthAudience,
			JWKSURL:    cfg.AuthJWKSURL,
			Algorithms: cfg.AuthAlgorithms,
		})
		logger.Info("OAuth authentication enabled", "issuer", cfg.AuthIssuer)
	}

	// Create metrics collector
	var metrics *middleware.Metrics
	if cfg.EnableMetrics {
		metrics = middleware.NewMetrics()
	}

	return &Server{
		sessions:      sessionStore,
		rateLimiter:   rateLimiter,
		authValidator: authValidator,
		authScopes:    cfg.AuthScopes,
		maxBodySize:   cfg.MaxBodySize,
		metrics:       metrics,
		logger:        logger,
		startTime:     time.Now(),
		cancelCleanup: cancel,
	}
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(s.loggingMiddleware)

	// Health endpoints (no auth required)
	r.Get("/health", s.handleHealth)
	r.Get("/ready", s.handleReady)

	// Metrics endpoint (no auth required)
	if s.metrics != nil {
		r.Get("/metrics", s.metrics.Handler())
	}

	// MCP endpoint with optional auth and rate limiting
	mcpHandler := http.HandlerFunc(s.handleMCP)

	// Build middleware chain for MCP endpoint
	var handler http.Handler = mcpHandler

	// Apply rate limiting if enabled
	if s.rateLimiter != nil {
		handler = s.rateLimiter.Middleware(middleware.GetRateLimitKey)(handler)
	}

	// Apply auth if enabled
	if s.authValidator != nil {
		handler = auth.Middleware(s.authValidator, s.authScopes)(handler)
	}

	r.Handle("/mcp", handler)

	return r
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() {
	if s.cancelCleanup != nil {
		s.cancelCleanup()
	}
}

// loggingMiddleware logs requests and records metrics.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		s.logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration_ms", duration.Milliseconds(),
			"session_id", r.Header.Get("Mcp-Session-Id"),
		)

		// Record metrics
		if s.metrics != nil {
			s.metrics.RecordRequest(r.URL.Path, duration)

			// Update session count
			if count, err := s.sessions.Count(r.Context()); err == nil {
				s.metrics.SetActiveSessions(int64(count))
			}
		}
	})
}

// handleHealth handles health check requests.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "healthy",
		"version": ServerVersion,
		"uptime":  time.Since(s.startTime).Seconds(),
	})
}

// handleReady handles readiness check requests.
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ready": true,
	})
}

// handleMCP handles all MCP requests.
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleMCPPost(w, r)
	case http.MethodDelete:
		s.handleMCPDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMCPDelete handles session deletion.
func (s *Server) handleMCPDelete(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := s.sessions.Delete(ctx, sessionID); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleMCPPost handles MCP POST requests.
func (s *Server) handleMCPPost(w http.ResponseWriter, r *http.Request) {
	// Apply body size limit to prevent DoS
	if s.maxBodySize > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, s.maxBodySize)
	}

	// Parse request
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Check if it's a body too large error
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			s.sendError(w, nil, ErrCodeInvalidRequest, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		s.sendError(w, nil, ErrCodeParseError, "Parse error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		s.sendError(w, req.ID, ErrCodeInvalidRequest, "Invalid JSON-RPC version", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	sessionID := r.Header.Get("Mcp-Session-Id")

	// Handle initialize (no session required)
	if req.Method == "initialize" {
		s.handleInitialize(w, r, req)
		return
	}

	// All other methods require a valid session
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusNotFound)
		return
	}

	sess, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) || errors.Is(err, session.ErrSessionExpired) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		s.sendError(w, req.ID, ErrCodeInternalError, "Session error", http.StatusInternalServerError)
		return
	}

	// Touch session to extend timeout
	s.sessions.Touch(ctx, sessionID)

	// Handle notification (no response)
	if req.ID == nil {
		s.handleNotification(w, r, req, sess)
		return
	}

	// Handle request methods
	switch req.Method {
	case "tools/list":
		s.handleToolsList(w, req)
	case "tools/call":
		s.handleToolsCall(w, req)
	default:
		s.sendError(w, req.ID, ErrCodeMethodNotFound, "Method not found: "+req.Method, http.StatusOK)
	}
}

// handleInitialize handles the initialize method.
func (s *Server) handleInitialize(w http.ResponseWriter, r *http.Request, req JSONRPCRequest) {
	ctx := r.Context()

	// Parse params
	params, ok := req.Params.(map[string]any)
	if !ok {
		params = make(map[string]any)
	}

	clientInfo := session.ClientInfo{IP: middleware.GetClientIP(r)}
	if ci, ok := params["clientInfo"].(map[string]any); ok {
		if name, ok := ci["name"].(string); ok {
			clientInfo.Name = name
		}
		if version, ok := ci["version"].(string); ok {
			clientInfo.Version = version
		}
	}

	// Create session
	sess, err := s.sessions.Create(ctx, clientInfo)
	if err != nil {
		if errors.Is(err, session.ErrSessionLimitReached) {
			s.sendError(w, req.ID, ErrCodeInternalError, "Server at capacity", http.StatusServiceUnavailable)
			return
		}
		s.sendError(w, req.ID, ErrCodeInternalError, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set session ID header
	w.Header().Set("Mcp-Session-Id", sess.ID)
	w.Header().Set("Content-Type", "application/json")

	// Send response
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"protocolVersion": ProtocolVersion,
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    ServerName,
				"version": ServerVersion,
			},
		},
	}

	json.NewEncoder(w).Encode(resp)
}

// handleNotification handles notification methods.
func (s *Server) handleNotification(w http.ResponseWriter, r *http.Request, req JSONRPCRequest, sess *session.Session) {
	ctx := r.Context()

	switch req.Method {
	case "notifications/initialized":
		s.sessions.MarkInitialized(ctx, sess.ID)
		w.WriteHeader(http.StatusNoContent)
	default:
		// Unknown notifications are ignored per spec
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleToolsList handles the tools/list method.
func (s *Server) handleToolsList(w http.ResponseWriter, req JSONRPCRequest) {
	w.Header().Set("Content-Type", "application/json")

	// Use cached tool schema to avoid rebuilding on every request
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"tools": cachedToolSchema,
		},
	}

	json.NewEncoder(w).Encode(resp)
}

// handleToolsCall handles the tools/call method.
func (s *Server) handleToolsCall(w http.ResponseWriter, req JSONRPCRequest) {
	w.Header().Set("Content-Type", "application/json")

	params, ok := req.Params.(map[string]any)
	if !ok {
		s.sendError(w, req.ID, ErrCodeInvalidParams, "Invalid params", http.StatusOK)
		return
	}

	toolName, _ := params["name"].(string)
	arguments, _ := params["arguments"].(map[string]any)

	var result map[string]any
	var isError bool

	switch toolName {
	case "calculate":
		result, isError = s.executeCalculate(arguments)
	default:
		result = map[string]any{"error": fmt.Sprintf("Unknown tool: %s", toolName)}
		isError = true
	}

	// Format result as MCP CallToolResult
	resultJSON, _ := json.Marshal(result)

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": string(resultJSON),
				},
			},
			"isError": isError,
		},
	}

	json.NewEncoder(w).Encode(resp)
}

// executeCalculate executes the calculate tool.
func (s *Server) executeCalculate(arguments map[string]any) (map[string]any, bool) {
	calcsRaw, ok := arguments["calculations"]
	if !ok {
		return map[string]any{"error": "calculations required"}, true
	}

	calcsArray, ok := calcsRaw.([]any)
	if !ok {
		return map[string]any{"error": "calculations must be an array"}, true
	}

	// Convert to calculator.Calculation slice
	calculations := make([]calculator.Calculation, 0, len(calcsArray))
	for _, c := range calcsArray {
		calcMap, ok := c.(map[string]any)
		if !ok {
			continue
		}

		calc := calculator.Calculation{
			Operation: calcMap["operation"].(string),
		}

		if name, ok := calcMap["name"].(string); ok {
			calc.Name = name
		}

		if args, ok := calcMap["args"].([]any); ok {
			calc.Args = args
		}

		calculations = append(calculations, calc)
	}

	result := calculator.Calculate(calculations)
	return map[string]any{
		"success": result.Success,
		"results": result.Results,
	}, false
}

// sendError sends a JSON-RPC error response.
func (s *Server) sendError(w http.ResponseWriter, id any, code int, message string, httpStatus int) {
	// Record error in metrics
	if s.metrics != nil {
		s.metrics.RecordError(code)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}

	json.NewEncoder(w).Encode(resp)
}

// Run starts the server on the given address.
func (s *Server) Run(addr string) error {
	s.logger.Info("starting MCP server",
		"address", addr,
		"version", ServerVersion,
		"protocol", ProtocolVersion,
	)
	return http.ListenAndServe(addr, s.Handler())
}
