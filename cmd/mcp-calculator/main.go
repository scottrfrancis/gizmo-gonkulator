// Package main is the entry point for the MCP calculator server.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/scottrfrancis/mcp-calculator/internal/server"
)

func main() {
	// Parse flags
	var (
		port    = flag.Int("port", 0, "Port to listen on (default: MCP_PORT or 8080)")
		host    = flag.String("host", "", "Host to bind to (default: MCP_HOST or 0.0.0.0)")
		version = flag.Bool("version", false, "Print version and exit")
	)
	flag.Parse()

	if *version {
		fmt.Printf("mcp-calculator %s (protocol %s)\n", server.ServerVersion, server.ProtocolVersion)
		os.Exit(0)
	}

	// Configure from environment
	cfg := server.DefaultConfig()

	if v := os.Getenv("MCP_SESSION_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.SessionTimeout = d
		}
	}

	if v := os.Getenv("MCP_MAX_SESSIONS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxSessions = n
		}
	}

	if v := os.Getenv("MCP_RATE_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.RateLimit = n
		}
	}

	if v := os.Getenv("MCP_RATE_LIMIT_BURST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.RateLimitBurst = n
		}
	}

	if v := os.Getenv("MCP_ENABLE_RATE_LIMIT"); v == "false" {
		cfg.EnableRateLimiter = false
	}

	// Auth configuration
	// API key auth (simpler alternative to OAuth)
	if v := os.Getenv("MCP_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	// OAuth configuration (used if API key not set)
	if v := os.Getenv("MCP_OAUTH_ENABLED"); v == "true" {
		cfg.AuthEnabled = true
	}
	if v := os.Getenv("MCP_OAUTH_ISSUER"); v != "" {
		cfg.AuthIssuer = v
	}
	if v := os.Getenv("MCP_OAUTH_AUDIENCE"); v != "" {
		cfg.AuthAudience = v
	}
	if v := os.Getenv("MCP_OAUTH_JWKS_URL"); v != "" {
		cfg.AuthJWKSURL = v
	}
	if v := os.Getenv("MCP_OAUTH_SCOPES"); v != "" {
		cfg.AuthScopes = strings.Split(v, ",")
	}

	// Determine address
	listenHost := *host
	if listenHost == "" {
		listenHost = os.Getenv("MCP_HOST")
		if listenHost == "" {
			listenHost = "0.0.0.0"
		}
	}

	listenPort := *port
	if listenPort == 0 {
		if v := os.Getenv("MCP_PORT"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				listenPort = n
			}
		}
		if listenPort == 0 {
			listenPort = 8080
		}
	}

	addr := fmt.Sprintf("%s:%d", listenHost, listenPort)

	// Create server
	srv := server.NewWithConfig(cfg)

	// Create HTTP server with graceful shutdown
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      srv.Handler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Setup signal handling
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown MCP server (stops cleanup goroutines)
		srv.Shutdown()

		httpServer.SetKeepAlivesEnabled(false)
		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Error("could not gracefully shutdown", "error", err)
		}
		close(done)
	}()

	// Start server
	logger.Info("starting MCP calculator server",
		"address", addr,
		"version", server.ServerVersion,
		"protocol", server.ProtocolVersion,
	)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	<-done
	logger.Info("server stopped")
}
