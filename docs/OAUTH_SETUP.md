# OAuth 2.1 Setup Guide

This guide explains how to configure OAuth 2.1 authentication for the MCP Calculator Server.

## Overview

The MCP server acts as an **OAuth 2.1 Resource Server**, validating JWT access tokens issued by your identity provider (IdP). This follows the MCP Authorization Specification (June 2025).

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  MCP Client │────▶│  Identity       │────▶│  MCP Server     │
│             │     │  Provider (IdP) │     │  (Resource Svr) │
└─────────────┘     └─────────────────┘     └─────────────────┘
      │                     │                       │
      │ 1. Login/Auth       │                       │
      │────────────────────▶│                       │
      │                     │                       │
      │ 2. Access Token     │                       │
      │◀────────────────────│                       │
      │                     │                       │
      │ 3. API Request + Token                      │
      │────────────────────────────────────────────▶│
      │                     │                       │
      │                     │ 4. Validate Token     │
      │                     │◀──────────────────────│
      │                     │──────────────────────▶│
      │                     │                       │
      │ 5. Response                                 │
      │◀────────────────────────────────────────────│
```

## Quick Start

### 1. Enable OAuth

Set the required environment variables:

```bash
export MCP_OAUTH_ENABLED=true
export MCP_OAUTH_ISSUER=https://your-idp.example.com/
export MCP_OAUTH_AUDIENCE=https://mcp-calculator.example.com
```

### 2. Start the Server

```bash
./mcp-calculator
# or
docker run -p 8080:8080 \
  -e MCP_OAUTH_ENABLED=true \
  -e MCP_OAUTH_ISSUER=https://your-idp.example.com/ \
  -e MCP_OAUTH_AUDIENCE=https://mcp-calculator.example.com \
  mcp-calculator
```

### 3. Make Authenticated Requests

```bash
# Get token from your IdP first, then:
curl -X POST http://localhost:8080/mcp \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize",...}'
```

## Configuration Reference

| Variable | Required | Description |
|----------|----------|-------------|
| `MCP_OAUTH_ENABLED` | Yes | Set to `true` to enable OAuth |
| `MCP_OAUTH_ISSUER` | Yes | OAuth issuer URL (e.g., `https://tenant.auth0.com/`) |
| `MCP_OAUTH_AUDIENCE` | Yes | Expected `aud` claim in tokens |
| `MCP_OAUTH_JWKS_URL` | No | JWKS endpoint (defaults to `{issuer}/.well-known/jwks.json`) |
| `MCP_OAUTH_SCOPES` | No | Required scopes, comma-separated |

## Identity Provider Setup

### Auth0

#### 1. Create an API

1. Go to **Auth0 Dashboard** → **Applications** → **APIs**
2. Click **Create API**
3. Configure:
   - **Name**: `MCP Calculator`
   - **Identifier**: `https://mcp-calculator.example.com` (this is your audience)
   - **Signing Algorithm**: RS256

#### 2. Configure Permissions (Scopes)

In the API settings, add permissions:
- `mcp:calculate` - Execute calculations
- `mcp:read` - List tools

#### 3. Create an Application

1. Go to **Applications** → **Create Application**
2. Choose **Machine to Machine** (for server-to-server) or **Single Page App** (for browser clients)
3. Authorize the application to use your API

#### 4. Configure MCP Server

```bash
export MCP_OAUTH_ENABLED=true
export MCP_OAUTH_ISSUER=https://YOUR_TENANT.auth0.com/
export MCP_OAUTH_AUDIENCE=https://mcp-calculator.example.com
```

#### 5. Get a Token (for testing)

```bash
# Machine-to-machine token
curl -X POST https://YOUR_TENANT.auth0.com/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "YOUR_CLIENT_ID",
    "client_secret": "YOUR_CLIENT_SECRET",
    "audience": "https://mcp-calculator.example.com",
    "grant_type": "client_credentials"
  }'
```

---

### Keycloak

#### 1. Create a Realm

1. Log into Keycloak Admin Console
2. Create a new realm (e.g., `mcp`)

#### 2. Create a Client

1. Go to **Clients** → **Create**
2. Configure:
   - **Client ID**: `mcp-calculator`
   - **Client Protocol**: `openid-connect`
   - **Access Type**: `confidential` (for server apps) or `public` (for SPAs)
   - **Valid Redirect URIs**: Your client's redirect URI

#### 3. Create Client Scopes

1. Go to **Client Scopes** → **Create**
2. Add scopes: `mcp:calculate`, `mcp:read`
3. Assign scopes to your client

#### 4. Configure MCP Server

```bash
export MCP_OAUTH_ENABLED=true
export MCP_OAUTH_ISSUER=https://keycloak.example.com/realms/mcp
export MCP_OAUTH_AUDIENCE=mcp-calculator
```

#### 5. Get a Token

```bash
# Client credentials flow
curl -X POST https://keycloak.example.com/realms/mcp/protocol/openid-connect/token \
  -d "client_id=mcp-calculator" \
  -d "client_secret=YOUR_SECRET" \
  -d "grant_type=client_credentials"
```

---

### Okta

#### 1. Create an Authorization Server

1. Go to **Security** → **API** → **Authorization Servers**
2. Click **Add Authorization Server**
3. Configure:
   - **Name**: `MCP Calculator`
   - **Audience**: `api://mcp-calculator`

#### 2. Add Scopes

1. In your authorization server, go to **Scopes**
2. Add: `mcp:calculate`, `mcp:read`

#### 3. Create an Application

1. Go to **Applications** → **Create App Integration**
2. Choose **API Services** (for M2M) or **OIDC - Single Page App**
3. Assign the application to your authorization server

#### 4. Configure MCP Server

```bash
export MCP_OAUTH_ENABLED=true
export MCP_OAUTH_ISSUER=https://YOUR_ORG.okta.com/oauth2/YOUR_AUTH_SERVER_ID
export MCP_OAUTH_AUDIENCE=api://mcp-calculator
```

#### 5. Get a Token

```bash
curl -X POST https://YOUR_ORG.okta.com/oauth2/YOUR_AUTH_SERVER_ID/v1/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "grant_type=client_credentials" \
  -d "scope=mcp:calculate mcp:read"
```

---

### Google Cloud Identity

#### 1. Create OAuth Credentials

1. Go to **Google Cloud Console** → **APIs & Services** → **Credentials**
2. Create **OAuth 2.0 Client ID**

#### 2. Configure MCP Server

```bash
export MCP_OAUTH_ENABLED=true
export MCP_OAUTH_ISSUER=https://accounts.google.com
export MCP_OAUTH_AUDIENCE=YOUR_CLIENT_ID.apps.googleusercontent.com
```

---

## Token Requirements

The MCP server validates the following JWT claims:

| Claim | Validation |
|-------|------------|
| `iss` | Must match `MCP_OAUTH_ISSUER` |
| `aud` | Must include `MCP_OAUTH_AUDIENCE` |
| `exp` | Must not be expired |
| `iat` | Must not be in the future |
| `scope` | Must include required scopes (if configured) |

### Example Valid Token Payload

```json
{
  "iss": "https://your-tenant.auth0.com/",
  "sub": "auth0|user123",
  "aud": "https://mcp-calculator.example.com",
  "iat": 1704067200,
  "exp": 1704070800,
  "scope": "mcp:calculate mcp:read"
}
```

## Scopes

### Available Scopes

| Scope | Description |
|-------|-------------|
| `mcp:calculate` | Execute the calculate tool |
| `mcp:read` | List tools, read server info |
| `mcp:*` | Full access (all operations) |

### Requiring Scopes

To require specific scopes, set:

```bash
export MCP_OAUTH_SCOPES=mcp:calculate,mcp:read
```

## Testing Without OAuth

For development, you can disable OAuth:

```bash
export MCP_OAUTH_ENABLED=false
./mcp-calculator
```

All requests will be allowed without authentication.

## Error Responses

### 401 Unauthorized

Missing or invalid token:

```json
{
  "jsonrpc": "2.0",
  "id": null,
  "error": {
    "code": -32001,
    "message": "Unauthorized: invalid token"
  }
}
```

**Common causes:**
- Missing `Authorization` header
- Malformed token
- Expired token
- Invalid signature (wrong JWKS)

### 403 Forbidden

Token valid but insufficient permissions:

```json
{
  "jsonrpc": "2.0",
  "id": null,
  "error": {
    "code": -32002,
    "message": "Forbidden: missing scope: mcp:calculate"
  }
}
```

## Troubleshooting

### Token Validation Fails

1. **Check issuer URL**: Ensure it exactly matches (including trailing slash)
   ```bash
   # Wrong
   export MCP_OAUTH_ISSUER=https://tenant.auth0.com
   # Correct
   export MCP_OAUTH_ISSUER=https://tenant.auth0.com/
   ```

2. **Check audience**: Verify the `aud` claim matches your config
   ```bash
   # Decode your token to check
   echo $TOKEN | cut -d'.' -f2 | base64 -d | jq .
   ```

3. **Check JWKS endpoint**: Verify it's accessible
   ```bash
   curl https://your-issuer/.well-known/jwks.json
   ```

### JWKS Fetch Errors

If you see "failed to fetch JWKS":
- Verify network connectivity to the IdP
- Check if the IdP requires mTLS
- Try setting `MCP_OAUTH_JWKS_URL` explicitly

### Clock Skew

Token validation may fail if server time is off. Ensure NTP is configured:

```bash
# Check time
date
# Sync (Linux)
sudo ntpdate -s time.nist.gov
```

## Security Best Practices

1. **Use short-lived tokens** (< 1 hour expiry)
2. **Always use HTTPS** in production
3. **Validate audience** to prevent token confusion attacks
4. **Use PKCE** for public clients (SPAs, mobile apps)
5. **Rotate secrets** regularly
6. **Monitor for failed auth attempts**

## Docker Compose Example

```yaml
version: "3.9"
services:
  mcp-calculator:
    image: mcp-calculator:latest
    ports:
      - "8080:8080"
    environment:
      - MCP_OAUTH_ENABLED=true
      - MCP_OAUTH_ISSUER=https://your-tenant.auth0.com/
      - MCP_OAUTH_AUDIENCE=https://mcp-calculator.example.com
      - MCP_OAUTH_SCOPES=mcp:calculate
```

## Further Reading

- [MCP Authorization Specification](https://modelcontextprotocol.io/specification/draft/basic/authorization)
- [OAuth 2.1 Specification](https://oauth.net/2.1/)
- [RFC 7519 - JSON Web Token](https://datatracker.ietf.org/doc/html/rfc7519)
- [RFC 8707 - Resource Indicators](https://datatracker.ietf.org/doc/html/rfc8707)
