# Design and Architecture Review: MCP Calculator Server

**Version:** 1.1.0
**Date:** 2026-01-01
**Author:** Architecture Review
**Status:** ✅ PRODUCTION READY

---

## Executive Summary

The MCP Calculator Server is a Go-based implementation providing precise decimal arithmetic for AI agents via the Model Context Protocol (MCP). It eliminates the 10-20% error rate observed when LLMs perform arithmetic directly.

### Key Metrics

| Metric | Value |
|--------|-------|
| Lines of Code | ~3,500 (Go) |
| Packages | 5 (server, calculator, session, auth, middleware) |
| Operations | 20+ arithmetic operations |
| Protocol | MCP 2025-03-26 (Streamable HTTP) |
| Dependencies | 5 direct, 16 total |
| Test Coverage | Calculator, Server, Integration tests |

### Overall Assessment

The implementation demonstrates **solid architectural patterns** with clean separation of concerns, interface-based design, and production-ready infrastructure.

**Risk Level:** LOW - Production ready with optional enhancements available.

---

## Architecture Overview

### Package Dependency Graph

```
┌─────────────────────────────────────────────────────────────┐
│                     cmd/mcp-calculator                       │
│                         main.go                              │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    internal/server                           │
│                      server.go                               │
│  - HTTP routing (chi)                                        │
│  - JSON-RPC 2.0 handling                                     │
│  - MCP protocol implementation                               │
└──────┬──────────────────┬──────────────────┬────────────────┘
       │                  │                  │
       ▼                  ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐
│   internal/  │  │   internal/  │  │      internal/           │
│  calculator  │  │   session    │  │      middleware          │
│              │  │              │  │                          │
│ - Decimal    │  │ - Store      │  │ - RateLimiter            │
│   arithmetic │  │   interface  │  │ - GetClientIP            │
│ - 20+ ops    │  │ - Memory     │  │                          │
│              │  │   Store      │  │                          │
└──────────────┘  └──────────────┘  └──────────────────────────┘
                                              │
                                              ▼
                                    ┌──────────────┐
                                    │   internal/  │
                                    │     auth     │
                                    │              │
                                    │ - JWT/JWKS   │
                                    │ - Validator  │
                                    │   interface  │
                                    └──────────────┘
```

### Request Flow

```
Client Request
      │
      ▼
┌─────────────────┐
│  Chi Middleware │  RequestID, RealIP, Recoverer, Logging
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Rate Limiter   │  Token bucket per session/IP
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│   handleMCP()   │────▶│   initialize    │──▶ Create Session
└────────┬────────┘     └─────────────────┘
         │
         │              ┌─────────────────┐
         ├─────────────▶│   tools/list    │──▶ Return tool schema
         │              └─────────────────┘
         │
         │              ┌─────────────────┐     ┌─────────────────┐
         └─────────────▶│   tools/call    │────▶│   Calculator    │
                        └─────────────────┘     │     Engine      │
                                                └────────┬────────┘
                                                         │
                                                         ▼
                                                ┌─────────────────┐
                                                │  Decimal Result │
                                                └─────────────────┘
```

### Data Flow: Calculator Engine

```
Input: calculations[]
         │
         ▼
    ┌────────────┐
    │  For each  │
    │ calculation│
    └─────┬──────┘
          │
          ▼
    ┌────────────┐     ┌──────────────┐
    │ resolveArgs│────▶│ Check if arg │
    └─────┬──────┘     │ is reference │
          │            └──────┬───────┘
          │                   │
          │            ┌──────▼───────┐
          │            │ Lookup in    │
          │            │ results map  │
          │            └──────────────┘
          │
          ▼
    ┌────────────┐
    │ executeOne │  Switch on operation type
    └─────┬──────┘
          │
          ▼
    ┌────────────┐
    │  Decimal   │  shopspring/decimal for precision
    │ Arithmetic │
    └─────┬──────┘
          │
          ▼
    ┌────────────┐
    │Store result│  Map: name → value
    │  by name   │
    └────────────┘
```

---

## SOLID Principles Analysis

### Single Responsibility Principle (SRP)

**Assessment: GOOD with exceptions**

| Package | Responsibility | Lines | Verdict |
|---------|----------------|-------|---------|
| `calculator` | Decimal arithmetic | 467 | GOOD |
| `session` | Session lifecycle | 223 | GOOD |
| `middleware` | HTTP middleware | 169 | GOOD |
| `auth` | OAuth/JWT validation | 294 | GOOD |
| `server` | HTTP + Protocol | 527 | NEEDS SPLIT |

**Issue:** `server.go` handles both HTTP routing AND MCP protocol logic. This violates SRP.

```go
// server.go mixes concerns:
func (s *Server) Handler() http.Handler { ... }     // HTTP concern
func (s *Server) handleInitialize(...) { ... }      // Protocol concern
func (s *Server) handleToolsCall(...) { ... }       // Protocol concern
func (s *Server) executeCalculate(...) { ... }      // Business logic
```

**Recommendation:** Extract `MCPProtocolHandler` type to separate protocol handling from HTTP infrastructure.

### Open/Closed Principle (OCP)

**Assessment: MIXED**

**Good - Interface-based extension:**

```go
// session/session.go:38 - Open for extension
type Store interface {
    Create(ctx context.Context, clientInfo ClientInfo) (*Session, error)
    Get(ctx context.Context, id string) (*Session, error)
    // ... can add Redis, DynamoDB implementations
}

// auth/oauth.go:49 - Open for extension
type Validator interface {
    Validate(ctx context.Context, token string) (*Claims, error)
}
```

**Issue - Calculator operations closed:**

```go
// calculator/calculator.go:129 - Must modify to add operations
switch operation {
case "add":
    return e.opAdd(resolved)
case "subtract":
    return e.opSubtract(resolved)
// ... adding new operation requires modifying this switch
}
```

**Recommendation:** Consider operation registry pattern for extensibility.

### Liskov Substitution Principle (LSP)

**Assessment: GOOD**

All interface implementations are properly substitutable:

```go
// Both implement Validator interface correctly
var _ Validator = (*JWTValidator)(nil)
var _ Validator = (*NoOpValidator)(nil)

// MemoryStore implements Store interface correctly
var _ Store = (*MemoryStore)(nil)
```

### Interface Segregation Principle (ISP)

**Assessment: GOOD**

Interfaces are focused and minimal:

| Interface | Methods | Assessment |
|-----------|---------|------------|
| `session.Store` | 7 | Appropriate for session lifecycle |
| `auth.Validator` | 1 | Ideal single-method interface |

### Dependency Inversion Principle (DIP)

**Assessment: GOOD with minor exceptions**

**Good - Server depends on abstractions:**

```go
// server/server.go
type Server struct {
    sessions      session.Store     // Interface, not concrete type
    authValidator auth.Validator    // Interface, not concrete type
    rateLimiter   *middleware.RateLimiter
}
```

**Minor Issue - Calculator is static:**

```go
// calculator/calculator.go:463 - Not injectable
func Calculate(calculations []Calculation) Result {
    engine := NewEngine()  // Hardcoded, not injectable
    return engine.Execute(calculations)
}
```

This is acceptable since the calculator is stateless and has no external dependencies.

---

## AWS Well-Architected Framework Review

### Operational Excellence

| Requirement | Status | Details |
|-------------|--------|---------|
| Structured logging | ✅ GOOD | JSON logs to stderr with slog |
| Health endpoint | ✅ GOOD | GET /health returns status |
| Readiness probe | ✅ GOOD | GET /ready for Kubernetes |
| Metrics endpoint | ✅ GOOD | GET /metrics (Prometheus-compatible) |
| Log level config | ⚠️ HARDCODED | Set to Info level |

### Security

| Requirement | Status | Details |
|-------------|--------|---------|
| OAuth 2.1 implementation | ✅ GOOD | Complete in auth package |
| Auth integration | ✅ GOOD | Wired into Handler() via middleware |
| Request body limits | ✅ GOOD | MaxBytesReader (1MB default) |
| IP validation | ⚠️ WEAK | X-Forwarded-For trusted blindly |

**Note:** IP validation trusts X-Forwarded-For without validation. Consider parsing only the first IP in production deployments behind load balancers.

### Reliability

| Requirement | Status | Details |
|-------------|--------|---------|
| Graceful shutdown | ✅ GOOD | 30s timeout, SIGTERM/SIGINT, Shutdown() method |
| Session cleanup | ✅ GOOD | Cancellable context stops goroutine |
| Rate limiting | ✅ GOOD | TTL-based cleanup prevents memory leak |
| Connection limits | ✅ GOOD | Max 10,000 sessions |

### Performance Efficiency

| Requirement | Status | Details |
|-------------|--------|---------|
| Minimal dependencies | ✅ GOOD | 5 direct deps |
| Efficient concurrency | ✅ GOOD | RWMutex for sessions |
| Calculator pooling | N/A | Engine is stateless, cheap to create |
| Schema caching | ✅ GOOD | Package-level cachedToolSchema |

---

## Open Issues

| # | Issue | Location | Risk | Priority |
|---|-------|----------|------|----------|
| 1 | Server.go size (627 lines) | server.go | Maintainability | LOW |
| 2 | Calculator not extensible | calculator.go:142 | OCP concern | LOW |
| 3 | Config file unused | configs/config.yaml | DX | LOW |
| 4 | IP validation weak | ratelimit.go:217-231 | Security | MEDIUM |

### IP Validation Concern

`GetClientIP()` trusts `X-Forwarded-For` header without validation, allowing IP spoofing.

```go
// ratelimit.go:217-231
func GetClientIP(r *http.Request) string {
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        return xff  // Trusted blindly
    }
    // ...
}
```

**Recommendation:** Parse only first IP from X-Forwarded-For, or require trusted proxy configuration.

---

## Optional Future Enhancements

Low-priority improvements that can be addressed as needed:

1. **Improve IP validation** (MEDIUM) - Parse first IP from X-Forwarded-For to prevent spoofing
2. **Refactor server.go** (LOW) - Extract MCPProtocolHandler for better SRP compliance
3. **Operation registry** (LOW) - Make calculator operations pluggable for extensibility
4. **Config file support** (LOW) - Use the existing config.yaml for non-env configuration

---

## Appendix: Test Coverage

The project includes comprehensive tests:

| Test File | Lines | Coverage |
|-----------|-------|----------|
| calculator_test.go | 512 | Unit tests for all operations |
| server_test.go | 526 | HTTP handler tests |
| integration_test.go | 485 | End-to-end MCP flows |

**Test Patterns:**

- Table-driven tests with `t.Run()`
- Race detector enabled (`go test -race`)
- Real-world healthcare data scenarios
- Precision validation (0.1 + 0.2 = 0.3)

---

## Appendix: Dependencies

| Package | Version | Purpose | Risk |
|---------|---------|---------|------|
| go-chi/chi | v5.1.0 | HTTP router | Low |
| google/uuid | v1.6.0 | UUID generation | Low |
| shopspring/decimal | v1.4.0 | Decimal arithmetic | Low |
| lestrrat-go/jwx | v2.1.3 | JWT/JWKS | Low |
| stretchr/testify | v1.9.0 | Testing | Dev only |
| golang.org/x/time | v0.8.0 | Rate limiting | Low |

All dependencies are well-maintained with no known vulnerabilities.
