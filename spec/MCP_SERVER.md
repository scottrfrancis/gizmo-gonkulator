# MCP Server Specification

## Overview

The MCP server implements the Model Context Protocol (MCP) specification version 2025-03-26, exposing the calculator engine as a tool for AI agents. It supports the Streamable HTTP transport for remote access and STDIO for local development.

## Protocol Compliance

### Version Support

- Primary: MCP 2025-03-26 (Streamable HTTP)
- Backward Compatible: MCP 2024-11-05 (STDIO only)

### JSON-RPC 2.0

All communication uses JSON-RPC 2.0 format:

```json
{
  "jsonrpc": "2.0",
  "id": "request-id",
  "method": "method_name",
  "params": {}
}
```

## Server Capabilities

```json
{
  "protocolVersion": "2025-03-26",
  "capabilities": {
    "tools": {}
  },
  "serverInfo": {
    "name": "mcp-calculator",
    "version": "1.0.0"
  }
}
```

## Supported Methods

### initialize

Client-server handshake. Returns server capabilities.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2025-03-26",
    "capabilities": {},
    "clientInfo": {
      "name": "client-name",
      "version": "1.0.0"
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2025-03-26",
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "mcp-calculator",
      "version": "1.0.0"
    }
  }
}
```

### notifications/initialized

Client acknowledgment after initialize. No response.

### tools/list

List available tools.

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "calculate",
        "description": "Perform precise arithmetic calculations...",
        "inputSchema": {
          "type": "object",
          "properties": {
            "calculations": {
              "type": "array",
              "items": { ... }
            }
          },
          "required": ["calculations"]
        }
      }
    ]
  }
}
```

### tools/call

Execute a tool.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "calculate",
    "arguments": {
      "calculations": [
        {"name": "sum", "operation": "sum", "args": [1, 2, 3]}
      ]
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"success\": true, \"results\": {\"sum\": 6}}"
      }
    ]
  }
}
```

## Tool Definition

### calculate

```json
{
  "name": "calculate",
  "description": "Perform precise arithmetic calculations using Decimal precision. Use this tool for ALL math operations - never calculate in your response. Supports batch operations: add, subtract, multiply, divide, sum, average, percentage, round, min, max, median, stddev, compare, abs, ceil, floor, roi, compound_interest, present_value. Results can reference previous calculations by name.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "calculations": {
        "type": "array",
        "description": "List of calculations to perform",
        "items": {
          "type": "object",
          "properties": {
            "name": {
              "type": "string",
              "description": "Label for the result (can be referenced by later calculations)"
            },
            "operation": {
              "type": "string",
              "enum": [
                "add", "subtract", "multiply", "divide",
                "sum", "average", "percentage", "round",
                "min", "max", "median", "stddev",
                "compare", "abs", "ceil", "floor",
                "roi", "compound_interest", "present_value"
              ],
              "description": "Operation to perform"
            },
            "args": {
              "type": "array",
              "items": {
                "oneOf": [
                  {"type": "number"},
                  {"type": "string", "description": "Reference to previous result by name"}
                ]
              },
              "description": "Numeric arguments or references to previous results"
            }
          },
          "required": ["operation", "args"]
        }
      }
    },
    "required": ["calculations"]
  }
}
```

## Error Responses

### JSON-RPC Errors

| Code | Message | Description |
|------|---------|-------------|
| -32700 | Parse error | Invalid JSON |
| -32600 | Invalid Request | Malformed JSON-RPC |
| -32601 | Method not found | Unknown method |
| -32602 | Invalid params | Invalid method parameters |
| -32603 | Internal error | Server error |

### MCP Errors

Tool execution errors are returned in the result content, not as JSON-RPC errors:

```json
{
  "content": [
    {
      "type": "text",
      "text": "{\"error\": \"Unknown tool: foo\"}"
    }
  ],
  "isError": true
}
```

## Health Endpoints

### GET /health

Returns server health status.

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": 3600
}
```

### GET /ready

Returns readiness status (for Kubernetes).

```json
{
  "ready": true
}
```

## Metrics (Optional)

### GET /metrics

Prometheus-compatible metrics:

```
mcp_requests_total{method="tools/call"} 1234
mcp_request_duration_seconds{method="tools/call"} 0.005
mcp_active_sessions 10
mcp_errors_total{code="-32603"} 5
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_PORT` | 8080 | HTTP server port |
| `MCP_HOST` | 0.0.0.0 | HTTP server host |
| `MCP_LOG_LEVEL` | info | Log level: debug, info, warn, error |
| `MCP_LOG_FORMAT` | json | Log format: json, text |
| `MCP_ENABLE_METRICS` | true | Enable /metrics endpoint |
| `MCP_ENABLE_AUTH` | false | Enable OAuth 2.1 authentication |

## Logging

Structured JSON logging to stderr:

```json
{
  "timestamp": "2025-01-01T12:00:00Z",
  "level": "info",
  "msg": "request completed",
  "method": "tools/call",
  "session_id": "abc123",
  "duration_ms": 5,
  "status": "success"
}
```

## Graceful Shutdown

- Listen for SIGTERM/SIGINT
- Stop accepting new connections
- Wait for active requests to complete (max 30s)
- Close all sessions
- Exit cleanly
