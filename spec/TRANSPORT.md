# Streamable HTTP Transport Specification

## Overview

This specification defines the Streamable HTTP transport for the MCP server, as per the MCP 2025-03-26 specification. This transport enables remote, network-based connections with support for multiple concurrent clients.

## Endpoint

The server MUST provide a single HTTP endpoint:

```
POST|GET|DELETE /mcp
```

This endpoint handles all MCP protocol messages.

## HTTP Methods

### POST /mcp

Client-to-server messages.

**Headers (Required):**
- `Content-Type: application/json`
- `Accept: application/json, text/event-stream`

**Headers (Optional):**
- `Mcp-Session-Id: <session-id>` (required after initialization)

**Request Body:**
JSON-RPC 2.0 message(s).

**Response:**
- `Content-Type: application/json` for single responses
- `Content-Type: text/event-stream` for streaming responses

### GET /mcp

Server-to-client notifications (SSE stream).

**Headers (Required):**
- `Accept: text/event-stream`
- `Mcp-Session-Id: <session-id>`

**Response:**
- `Content-Type: text/event-stream`
- Keep-alive connection for server-initiated messages

### DELETE /mcp

Terminate session.

**Headers (Required):**
- `Mcp-Session-Id: <session-id>`

**Response:**
- `204 No Content` on success

## Session Management

### Session Creation

1. Client sends `POST /mcp` with `initialize` request (no session ID)
2. Server creates new session
3. Server responds with `Mcp-Session-Id` header
4. Client MUST include `Mcp-Session-Id` in all subsequent requests

### Session Headers

**Response Header (on initialize):**
```
Mcp-Session-Id: uuid-v4-session-id
```

**Request Header (after initialize):**
```
Mcp-Session-Id: uuid-v4-session-id
```

### Session Timeout

- Sessions timeout after 10 minutes of inactivity (configurable)
- Client can extend session by any request
- Server MAY close idle sessions

### Session Termination

1. Client sends `DELETE /mcp` with session ID
2. Server closes session immediately
3. Server responds with 204

## Streaming Responses (SSE)

For long-running operations, the server MAY respond with SSE:

```http
HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

event: message
data: {"jsonrpc":"2.0","id":1,"result":{...}}

event: message
data: {"jsonrpc":"2.0","method":"notifications/progress","params":{...}}

```

### SSE Event Format

```
event: message
data: <json-rpc-message>

```

Note: Each message ends with two newlines.

## Error Responses

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 204 | Success (no content) |
| 400 | Bad Request - malformed JSON |
| 401 | Unauthorized - invalid or missing auth |
| 403 | Forbidden - insufficient permissions |
| 404 | Not Found - invalid session |
| 429 | Too Many Requests - rate limited |
| 500 | Internal Server Error |
| 503 | Service Unavailable - server overloaded |

### Error Response Body

```json
{
  "jsonrpc": "2.0",
  "id": null,
  "error": {
    "code": -32600,
    "message": "Invalid Request"
  }
}
```

## Connection Handling

### Multiple Clients

- Server MUST support multiple concurrent sessions
- Each session is independent
- Sessions are identified by `Mcp-Session-Id`

### Connection Limits

| Config | Default | Description |
|--------|---------|-------------|
| `max_connections` | 100 | Maximum concurrent connections |
| `max_sessions` | 1000 | Maximum concurrent sessions |
| `session_timeout` | 600s | Session inactivity timeout |
| `request_timeout` | 30s | Individual request timeout |
| `idle_timeout` | 300s | Connection idle timeout |

### Keep-Alive

- Server SHOULD support HTTP keep-alive
- SSE connections use keep-alive by default
- Client MAY close connection and reconnect with same session ID

## Security

### DNS Rebinding Protection

Server MUST validate the `Host` header:

```go
func validateHost(r *http.Request) bool {
    host := r.Host
    // Allow configured hosts only
    return isAllowedHost(host)
}
```

### CORS

For browser-based clients:

```
Access-Control-Allow-Origin: <configured-origin>
Access-Control-Allow-Methods: GET, POST, DELETE
Access-Control-Allow-Headers: Content-Type, Accept, Mcp-Session-Id, Authorization
Access-Control-Expose-Headers: Mcp-Session-Id
```

### TLS

- Production deployments MUST use HTTPS
- Server SHOULD support TLS 1.2+ only
- Server MAY require client certificates

## Request/Response Examples

### Initialize Flow

**Request:**
```http
POST /mcp HTTP/1.1
Host: localhost:8080
Content-Type: application/json
Accept: application/json, text/event-stream

{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2025-03-26",
    "clientInfo": {"name": "test-client", "version": "1.0"}
  }
}
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json
Mcp-Session-Id: a1b2c3d4-e5f6-7890-abcd-ef1234567890

{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2025-03-26",
    "serverInfo": {"name": "mcp-calculator", "version": "1.0.0"},
    "capabilities": {"tools": {}}
  }
}
```

### Tool Call Flow

**Request:**
```http
POST /mcp HTTP/1.1
Host: localhost:8080
Content-Type: application/json
Accept: application/json, text/event-stream
Mcp-Session-Id: a1b2c3d4-e5f6-7890-abcd-ef1234567890

{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "calculate",
    "arguments": {
      "calculations": [{"operation": "sum", "args": [1, 2, 3]}]
    }
  }
}
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"success\":true,\"results\":{\"result_0\":6}}"
    }]
  }
}
```

## Implementation Requirements

1. Use standard Go `net/http` or compatible framework
2. Implement session store (in-memory for single node)
3. Support both JSON and SSE response types
4. Validate all headers and session IDs
5. Implement connection pooling for downstream services
6. Log all requests with correlation IDs
