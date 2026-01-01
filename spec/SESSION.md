# Session Management and Rate Limiting Specification

## Overview

This specification defines session management and rate limiting for the MCP server. These features are critical for production deployments to ensure resource control and fair access.

## Session Management

### Session Lifecycle

```
┌─────────┐     initialize      ┌─────────┐
│  None   │ ──────────────────▶ │ Active  │
└─────────┘                     └────┬────┘
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
               request           timeout          DELETE
                    │                │                │
                    ▼                ▼                ▼
               ┌─────────┐     ┌─────────┐     ┌─────────┐
               │ Active  │     │ Expired │     │ Closed  │
               │(renewed)│     │         │     │         │
               └─────────┘     └─────────┘     └─────────┘
```

### Session Data Structure

```go
type Session struct {
    ID           string            // UUID v4
    CreatedAt    time.Time         // Session creation time
    LastActivity time.Time         // Last request time
    ClientInfo   ClientInfo        // Client identification
    Metadata     map[string]any    // Custom metadata
}

type ClientInfo struct {
    Name    string // Client name from initialize
    Version string // Client version
    IP      string // Client IP address
}
```

### Session Store Interface

```go
type SessionStore interface {
    // Create creates a new session and returns its ID
    Create(ctx context.Context, clientInfo ClientInfo) (*Session, error)

    // Get retrieves a session by ID, returns ErrSessionNotFound if not found
    Get(ctx context.Context, id string) (*Session, error)

    // Touch updates the LastActivity timestamp
    Touch(ctx context.Context, id string) error

    // Delete removes a session
    Delete(ctx context.Context, id string) error

    // Cleanup removes expired sessions
    Cleanup(ctx context.Context) (int, error)

    // Count returns the number of active sessions
    Count(ctx context.Context) (int, error)
}
```

### In-Memory Session Store

For single-node deployments:

```go
type MemorySessionStore struct {
    sessions map[string]*Session
    mu       sync.RWMutex
    timeout  time.Duration
    maxSize  int
}
```

**Configuration:**

| Config | Default | Description |
|--------|---------|-------------|
| `session_timeout` | 600s | Inactivity timeout |
| `max_sessions` | 10000 | Maximum concurrent sessions |
| `cleanup_interval` | 60s | Expired session cleanup interval |

### Session Errors

| Error | HTTP Code | Description |
|-------|-----------|-------------|
| `ErrSessionNotFound` | 404 | Session ID not found |
| `ErrSessionExpired` | 401 | Session has expired |
| `ErrSessionLimitReached` | 503 | Maximum sessions reached |

## Rate Limiting

### Rate Limit Algorithm

Use the **Token Bucket** algorithm:

- Each client gets a bucket with N tokens
- Tokens replenish at rate R per second
- Each request consumes 1 token
- If bucket is empty, request is rejected (429)

### Rate Limit Tiers

| Tier | Requests/min | Burst | Description |
|------|--------------|-------|-------------|
| Default | 60 | 10 | Unauthenticated or basic |
| Authenticated | 300 | 50 | Valid OAuth token |
| Premium | 1000 | 100 | Premium scope |

### Rate Limit Headers

**Response Headers:**

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1704067200
Retry-After: 30
```

### Rate Limiter Interface

```go
type RateLimiter interface {
    // Allow checks if a request is allowed
    // Returns remaining tokens and reset time
    Allow(ctx context.Context, key string) (allowed bool, remaining int, reset time.Time, err error)

    // Tier returns the rate limit tier for a key
    Tier(ctx context.Context, key string) RateTier
}

type RateTier struct {
    Name         string
    RequestsPerMin int
    BurstSize    int
}
```

### Rate Limit Key

Rate limits are applied per:

1. **Session ID** (primary): Each session has its own bucket
2. **IP Address** (fallback): For unauthenticated requests

```go
func getRateLimitKey(r *http.Request) string {
    if sessionID := r.Header.Get("Mcp-Session-Id"); sessionID != "" {
        return "session:" + sessionID
    }
    return "ip:" + getClientIP(r)
}
```

### Rate Limit Response

**429 Too Many Requests:**

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1704067200
Retry-After: 30

{
  "jsonrpc": "2.0",
  "id": null,
  "error": {
    "code": -32000,
    "message": "Rate limit exceeded. Retry after 30 seconds."
  }
}
```

## Request Queuing

For fair access under high load:

### Queue Configuration

| Config | Default | Description |
|--------|---------|-------------|
| `max_queue_size` | 1000 | Maximum queued requests |
| `queue_timeout` | 30s | Maximum time in queue |
| `max_concurrent` | 100 | Maximum concurrent processing |

### Queue Behavior

1. If concurrent requests < max, process immediately
2. If concurrent >= max but queue < max_queue_size, enqueue
3. If queue is full, reject with 503

## Connection Limits

### Per-IP Limits

| Config | Default | Description |
|--------|---------|-------------|
| `max_connections_per_ip` | 10 | Maximum connections per IP |
| `max_sessions_per_ip` | 50 | Maximum sessions per IP |

### Implementation

```go
type ConnectionLimiter struct {
    connections map[string]int // IP -> connection count
    mu          sync.Mutex
    maxPerIP    int
}

func (l *ConnectionLimiter) Acquire(ip string) error {
    l.mu.Lock()
    defer l.mu.Unlock()

    if l.connections[ip] >= l.maxPerIP {
        return ErrConnectionLimitReached
    }
    l.connections[ip]++
    return nil
}
```

## Metrics

### Session Metrics

```
mcp_sessions_active{} 150
mcp_sessions_created_total{} 1000
mcp_sessions_expired_total{} 850
mcp_sessions_closed_total{} 50
```

### Rate Limit Metrics

```
mcp_ratelimit_requests_total{tier="default"} 5000
mcp_ratelimit_rejected_total{tier="default"} 100
mcp_ratelimit_remaining{session="abc"} 45
```

### Queue Metrics

```
mcp_queue_size{} 25
mcp_queue_wait_seconds{quantile="0.5"} 0.1
mcp_queue_wait_seconds{quantile="0.99"} 2.5
mcp_queue_rejected_total{} 10
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_SESSION_TIMEOUT` | 600 | Session timeout (seconds) |
| `MCP_MAX_SESSIONS` | 10000 | Maximum sessions |
| `MCP_RATE_LIMIT_RPM` | 60 | Requests per minute |
| `MCP_RATE_LIMIT_BURST` | 10 | Burst size |
| `MCP_MAX_CONNECTIONS_PER_IP` | 10 | Max connections per IP |
| `MCP_MAX_QUEUE_SIZE` | 1000 | Request queue size |
| `MCP_QUEUE_TIMEOUT` | 30 | Queue timeout (seconds) |

### Config File (YAML)

```yaml
session:
  timeout: 600s
  max_sessions: 10000
  cleanup_interval: 60s

rate_limit:
  enabled: true
  default:
    requests_per_minute: 60
    burst_size: 10
  authenticated:
    requests_per_minute: 300
    burst_size: 50

queue:
  max_size: 1000
  timeout: 30s
  max_concurrent: 100

connections:
  max_per_ip: 10
  max_sessions_per_ip: 50
```

## Graceful Degradation

Under high load:

1. **Warning** (75% capacity): Log warning, increase monitoring
2. **Throttle** (90% capacity): Reduce rate limits by 50%
3. **Shed Load** (100% capacity): Reject new sessions with 503

```go
type LoadShedder struct {
    currentLoad func() float64
    thresholds  []LoadThreshold
}

type LoadThreshold struct {
    Percentage float64
    Action     LoadAction
}

type LoadAction int

const (
    ActionNone LoadAction = iota
    ActionWarn
    ActionThrottle
    ActionReject
)
```
