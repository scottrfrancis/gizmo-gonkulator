# OAuth 2.1 Authentication Specification

## Overview

This specification defines the OAuth 2.1 authentication mechanism for the MCP server, following the MCP Authorization Specification (June 2025 revision). The MCP server acts as an OAuth 2.1 Resource Server.

## Architecture

```
┌─────────────────┐     ┌──────────────────────┐     ┌─────────────────┐
│   MCP Client    │────▶│  Authorization       │────▶│   MCP Server    │
│   (OAuth Client)│     │  Server (IdP)        │     │  (Resource Svr) │
└─────────────────┘     └──────────────────────┘     └─────────────────┘
        │                        │                          │
        │  1. Auth Request       │                          │
        │───────────────────────▶│                          │
        │                        │                          │
        │  2. Auth Code + PKCE   │                          │
        │◀───────────────────────│                          │
        │                        │                          │
        │  3. Token Request      │                          │
        │───────────────────────▶│                          │
        │                        │                          │
        │  4. Access Token       │                          │
        │◀───────────────────────│                          │
        │                        │                          │
        │  5. API Request + Token│                          │
        │──────────────────────────────────────────────────▶│
        │                        │                          │
        │  6. Token Validation   │◀─────────────────────────│
        │                        │─────────────────────────▶│
        │                        │                          │
        │  7. API Response       │                          │
        │◀──────────────────────────────────────────────────│
```

## OAuth 2.1 Requirements

### PKCE (Proof Key for Code Exchange)

PKCE is REQUIRED for all clients (RFC 7636):

- `code_verifier`: Random 43-128 character string
- `code_challenge`: SHA-256 hash of verifier, base64url encoded
- `code_challenge_method`: Must be `S256`

### Resource Indicators (RFC 8707)

Clients MUST include resource indicators in token requests:

```
resource=https://mcp-calculator.example.com/mcp
```

This binds the access token to the specific MCP server.

## Server Configuration

### Protected Resource Metadata (RFC 9728)

The server MUST expose metadata at:

```
GET /.well-known/oauth-protected-resource
```

**Response:**
```json
{
  "resource": "https://mcp-calculator.example.com/mcp",
  "authorization_servers": [
    "https://auth.example.com"
  ],
  "bearer_methods_supported": ["header"],
  "scopes_supported": ["mcp:calculate", "mcp:read"],
  "resource_documentation": "https://docs.example.com/mcp"
}
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `MCP_OAUTH_ENABLED` | No | Enable OAuth (default: false) |
| `MCP_OAUTH_ISSUER` | Yes* | OAuth issuer URL |
| `MCP_OAUTH_AUDIENCE` | Yes* | Expected audience (this server) |
| `MCP_OAUTH_JWKS_URL` | No | JWKS endpoint (default: issuer/.well-known/jwks.json) |
| `MCP_OAUTH_SCOPES` | No | Required scopes (comma-separated) |
| `MCP_OAUTH_ALLOWED_ALGORITHMS` | No | Allowed signing algorithms (default: RS256,ES256) |

*Required when OAuth is enabled

## Token Validation

### Request Header

```
Authorization: Bearer <access_token>
```

### Validation Steps

1. **Extract Token**: Parse `Authorization` header
2. **Decode JWT**: Validate JWT structure
3. **Verify Signature**: Check against JWKS
4. **Validate Claims**:
   - `iss`: Must match configured issuer
   - `aud`: Must include this server's resource URI
   - `exp`: Must not be expired
   - `iat`: Must not be in the future
   - `scope`: Must include required scopes (if configured)

### Token Introspection (Optional)

For opaque tokens, use RFC 7662 introspection:

```http
POST /oauth/introspect
Content-Type: application/x-www-form-urlencoded

token=<access_token>&token_type_hint=access_token
```

## Scopes

### Defined Scopes

| Scope | Description |
|-------|-------------|
| `mcp:calculate` | Execute calculate tool |
| `mcp:read` | List tools and read server info |
| `mcp:*` | Full access (all operations) |

### Scope Enforcement

```go
func requireScopes(required []string, token Token) error {
    for _, scope := range required {
        if !token.HasScope(scope) {
            return ErrInsufficientScope
        }
    }
    return nil
}
```

## Error Responses

### 401 Unauthorized

Missing or invalid token:

```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Bearer realm="mcp-calculator", error="invalid_token"

{
  "jsonrpc": "2.0",
  "id": null,
  "error": {
    "code": -32001,
    "message": "Unauthorized: invalid or missing access token"
  }
}
```

### 403 Forbidden

Insufficient scopes:

```http
HTTP/1.1 403 Forbidden
WWW-Authenticate: Bearer realm="mcp-calculator", error="insufficient_scope", scope="mcp:calculate"

{
  "jsonrpc": "2.0",
  "id": null,
  "error": {
    "code": -32002,
    "message": "Forbidden: insufficient scope"
  }
}
```

## Implementation

### Token Validator Interface

```go
type TokenValidator interface {
    // Validate validates an access token and returns claims
    Validate(ctx context.Context, token string) (*Claims, error)
}

type Claims struct {
    Subject   string   // sub claim
    Issuer    string   // iss claim
    Audience  []string // aud claim
    Scopes    []string // scope claim (space-separated in JWT)
    ExpiresAt time.Time
    IssuedAt  time.Time
    Extra     map[string]interface{} // additional claims
}
```

### JWT Validator

```go
type JWTValidator struct {
    issuer    string
    audience  string
    jwksURL   string
    keySet    jwk.Set
    algorithms []string
}

func (v *JWTValidator) Validate(ctx context.Context, token string) (*Claims, error) {
    // 1. Parse and verify JWT signature
    // 2. Validate standard claims
    // 3. Extract and return claims
}
```

### Middleware

```go
func AuthMiddleware(validator TokenValidator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearerToken(r)
            if token == "" {
                unauthorized(w, "missing token")
                return
            }

            claims, err := validator.Validate(r.Context(), token)
            if err != nil {
                unauthorized(w, err.Error())
                return
            }

            ctx := context.WithValue(r.Context(), claimsKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Security Considerations

### Token Storage

- NEVER log access tokens
- NEVER store tokens in session state
- Validate tokens on every request

### Token Lifetime

- Access tokens SHOULD be short-lived (< 1 hour)
- Server SHOULD NOT cache token validation results

### Algorithm Restrictions

- MUST reject "none" algorithm
- MUST reject symmetric algorithms (HS256, etc.) for public clients
- RECOMMENDED: RS256, ES256

### Token Binding

- Tokens MUST be bound to the resource via `aud` claim
- Tokens MUST NOT be forwarded to upstream services

## Testing Without OAuth

For development/testing, OAuth can be disabled:

```bash
MCP_OAUTH_ENABLED=false ./mcp-calculator
```

When disabled:
- All requests are allowed
- No `Authorization` header required
- Scopes are not enforced

## Integration with Identity Providers

### Auth0

```bash
MCP_OAUTH_ISSUER=https://your-tenant.auth0.com/
MCP_OAUTH_AUDIENCE=https://mcp-calculator.example.com
```

### Keycloak

```bash
MCP_OAUTH_ISSUER=https://keycloak.example.com/realms/myrealm
MCP_OAUTH_AUDIENCE=mcp-calculator
```

### Okta

```bash
MCP_OAUTH_ISSUER=https://your-org.okta.com/oauth2/default
MCP_OAUTH_AUDIENCE=api://mcp-calculator
```
