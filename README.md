![MCP Calculator Banner](banner.jpg)

# mcp-calculator

<img src="Logo.jpg" alt="MCP Calculator Logo" width="120" align="right" />

**Precise Arithmetic for AI Agents**

A Model Context Protocol (MCP) server that eliminates AI math errors by providing deterministic arithmetic operations using Decimal precision.

## The Problem

LLMs confidently produce wrong arithmetic. This isn't ignorance—language models are text prediction engines, not calculators.

**Real-world examples:**

| Context | AI Output | Correct Value | Error |
|---------|-----------|---------------|-------|
| Financial report | "72% decline" | 26.3% decline | Confused ratio with % change |
| Trend analysis | "trending higher" | trending lower | Incorrect comparison |

**Observed error rate:** 10-20% when AI does math vs. 0% with this tool.

## Quick Start

### Running the Server

```bash
# Using Go
go run ./cmd/mcp-calculator

# Using Docker
docker run -p 8080:8080 mcp-calculator

# Using Make
make run
```

### MCP Client Configuration

Add to your MCP client configuration (e.g., Claude Desktop):

```json
{
  "mcpServers": {
    "calculator": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

### Example Request

```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-03-26",
      "clientInfo": {"name": "curl", "version": "1.0"}
    }
  }'
```

## Supported Operations

### Basic Arithmetic
- `add` - Sum values: `[100, 200]` → 300
- `subtract` - Difference: `[500, 200]` → 300
- `multiply` - Product: `[10, 5]` → 50
- `divide` - Division: `[100, 4]` → 25

### Statistical
- `sum` - Total of array: `[100, 200, 300]` → 600
- `average` - Mean: `[100, 200, 300]` → 200
- `min` - Minimum: `[100, 50, 200]` → 50
- `max` - Maximum: `[100, 50, 200]` → 200
- `median` - Median value
- `stddev` - Standard deviation

### Financial
- `percentage` - Percent change: `[150, 100]` → 50.0 (50% increase)
- `roi` - Return on investment: `[gain, cost]`
- `compound_interest` - Future value: `[principal, rate, periods]`
- `present_value` - Present value: `[future_value, rate, periods]`

### Utility
- `round` - Round to decimals: `[3.14159, 2]` → 3.14
- `abs` - Absolute value
- `ceil` - Round up
- `floor` - Round down

### Comparison
- `compare` - Boolean comparison: `[5, 10, "<"]` → True

## Variable References

Calculations can reference previous results by name:

```json
{
  "calculations": [
    {"name": "oct_rate", "operation": "divide", "args": [2561276, 8]},
    {"name": "sep_rate", "operation": "divide", "args": [8782334, 21]},
    {"name": "change", "operation": "percentage", "args": ["oct_rate", "sep_rate"]}
  ]
}
```

## Why Decimal Precision?

```
// Float (IEEE 754) - imprecise
0.1 + 0.2 = 0.30000000000000004

// mcp-calculator (Decimal) - exact
0.1 + 0.2 = 0.3
```

For financial applications, floating-point errors are unacceptable.

## Error Handling

Errors are returned per-calculation, not as exceptions:

```json
{
  "success": true,
  "results": {
    "valid": 6,
    "error": {"error": "division by zero"}
  }
}
```

## Configuration

The server is configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_HOST` | `0.0.0.0` | Host to bind to |
| `MCP_PORT` | `8080` | Port to listen on |
| `MCP_SESSION_TIMEOUT` | `10m` | Session timeout duration |
| `MCP_MAX_SESSIONS` | `10000` | Maximum concurrent sessions |
| `MCP_RATE_LIMIT` | `60` | Requests per minute |
| `MCP_RATE_LIMIT_BURST` | `10` | Burst size for rate limiting |
| `MCP_API_KEY` | (none) | API key for authentication |

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/mcp` | POST | MCP JSON-RPC endpoint |
| `/mcp` | DELETE | Delete session |
| `/health` | GET | Health check |
| `/ready` | GET | Readiness probe |
| `/metrics` | GET | Prometheus metrics |

## Development

```bash
# Clone and build
git clone https://github.com/scottrfrancis/mcp-calculator
cd mcp-calculator

# Install dependencies
make deps

# Run tests
make test

# Run with race detection
go test -race ./...

# Build binary
make build

# Run locally
make run
```

## Project Structure

```
mcp-calculator/
├── cmd/mcp-calculator/     # Server entry point
├── internal/
│   ├── calculator/         # Decimal arithmetic engine
│   ├── server/             # MCP server implementation
│   ├── session/            # Session management
│   ├── auth/               # OAuth/API key authentication
│   └── middleware/         # Rate limiting, metrics
├── test/                   # Test suites
├── spec/                   # Protocol specifications
├── docs/                   # Documentation
└── reference/              # Python reference implementation
```

## Architecture

- **Protocol:** MCP 2025-03-26 (Streamable HTTP)
- **Precision:** Decimal arithmetic via `shopspring/decimal`
- **Concurrency:** Goroutine-safe with RWMutex
- **Authentication:** OAuth 2.1 or API key
- **Rate Limiting:** Token bucket per session/IP

See [DESIGN_AND_ARCHITECTURE.md](docs/DESIGN_AND_ARCHITECTURE.md) for details.

## Python Reference Implementation

A Python reference implementation is available in the `reference/` directory for testing and development purposes.

```bash
cd reference
pip install -e ".[dev]"
python -m mcp_calculator --test
```

## License

MIT
