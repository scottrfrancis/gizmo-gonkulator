# MCP Calculator Server Testing Guide

## Server Details

| Setting | Value |
|---------|-------|
| **Server URL** | `https://catalyst-mcp-ec2.scootersoft.info/calc` |
| **API Key** | `mcp-calc-secret-key-2026` |
| **Auth Header** | `X-API-Key: mcp-calc-secret-key-2026` |
| **Protocol** | MCP 2025-03-26 (Streamable HTTP) |
| **EC2 Instance** | `i-07f420b768f425ed9` (us-east-2) |

> **Local Development:** Use `catalyst-mcp.scootersoft.info` for local testing with mini.local.

---

## 1. Testing with curl

### 1.1 Health Check (requires API key via nginx)

```bash
curl -s https://catalyst-mcp-ec2.scootersoft.info/calc/health \
  -H "X-API-Key: mcp-calc-secret-key-2026" | jq .
```

Expected response:

```json
{
  "status": "healthy",
  "uptime": 123.456,
  "version": "1.0.0"
}
```

### 1.2 Initialize a Session

```bash
curl -s -i -X POST "https://catalyst-mcp-ec2.scootersoft.info/calc/mcp" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mcp-calc-secret-key-2026" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-03-26",
      "capabilities": {},
      "clientInfo": {"name": "curl-test", "version": "1.0.0"}
    }
  }'
```

**Important:** Copy the `Mcp-Session-Id` header from the response for subsequent requests.

### 1.3 List Available Tools

```bash
export SESSION_ID="14938f48-2065-44ed-9b56-a58dbcfe54b4"  # Replace with actual SESSION_ID from the previous response

# Replace SESSION_ID with the value from step 1.2
curl -s -X POST "https://catalyst-mcp-ec2.scootersoft.info/calc/mcp" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mcp-calc-secret-key-2026" \
  -H "Mcp-Session-Id: ${SESSION_ID}" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }' | jq '.result.tools[].name'
```

### 1.4 Simple Calculation (0.1 + 0.2)

```bash
curl -s -X POST "https://catalyst-mcp-ec2.scootersoft.info/calc/mcp" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mcp-calc-secret-key-2026" \
  -H "Mcp-Session-Id: ${SESSION_ID}" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "calculate",
      "arguments": {
        "calculations": [
          {"name": "result", "operation": "add", "args": [0.1, 0.2]}
        ]
      }
    }
  }' | jq '.result.content[0].text'
```

Expected: `{"results":{"result":0.3},"success":true}`

### 1.5 Batch Calculation with References

```bash
curl -s -X POST "https://catalyst-mcp-ec2.scootersoft.info/calc/mcp" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mcp-calc-secret-key-2026" \
  -H "Mcp-Session-Id: ${SESSION_ID}" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "calculate",
      "arguments": {
        "calculations": [
          {"name": "revenue", "operation": "sum", "args": [50000, 75000, 62000]},
          {"name": "costs", "operation": "sum", "args": [30000, 45000, 38000]},
          {"name": "profit", "operation": "subtract", "args": ["revenue", "costs"]},
          {"name": "margin_pct", "operation": "multiply", "args": [{"operation": "divide", "args": ["profit", "revenue"]}, 100]}
        ]
      }
    }
  }' | jq '.result.content[0].text'
```

### 1.6 All-in-One Test Script

Save as `~/test-mcp.sh`:

```bash
#!/bin/bash
SERVER="https://catalyst-mcp-ec2.scootersoft.info/calc"
API_KEY="mcp-calc-secret-key-2026"

echo "=== Health Check ==="
curl -s "$SERVER/health" -H "X-API-Key: $API_KEY" | jq .

echo -e "\n=== Initialize Session ==="
INIT=$(curl -s -i -X POST "$SERVER/mcp" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}')

SESSION_ID=$(echo "$INIT" | grep -i "Mcp-Session-Id:" | awk '{print $2}' | tr -d '\r')
echo "Session: $SESSION_ID"

echo -e "\n=== Precision Test: 0.1 + 0.2 ==="
curl -s -X POST "$SERVER/mcp" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"calculate","arguments":{"calculations":[{"name":"result","operation":"add","args":[0.1,0.2]}]}}}' \
  | jq '.result.content[0].text' -r

echo -e "\n=== Done ==="
```

Run with: `chmod +x ~/test-mcp.sh && ~/test-mcp.sh`

---

## 2. Testing with Claude Desktop

### 2.1 Configuration

Edit your Claude Desktop config file:

| OS | Config File Location |
|----|---------------------|
| __macOS__ | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| __Linux__ | `~/.config/claude-desktop/config.json` |
| __Windows__ | `%APPDATA%\Claude\claude_desktop_config.json` |

Add this configuration (Claude Desktop uses `mcp-remote` for HTTPS servers):

```json
{
  "mcpServers": {
    "calculator": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "https://catalyst-mcp-ec2.scootersoft.info/calc/mcp",
        "--header", "X-API-Key: mcp-calc-secret-key-2026"
      ],
      "env": {
        "PATH": "/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin"
      }
    }
  }
}
```

**Note:** Adjust the `PATH` to include your Node.js bin directory if npx isn't found. Uses `X-API-Key` header (nginx validates the key).

### 2.2 Restart Claude Desktop

After saving the config, fully quit and restart Claude Desktop for the MCP server to be recognized.

### 2.3 Verify Connection

In Claude Desktop, you should see "calculator" listed in the MCP servers (usually shown in the interface or via a tools menu).

### 2.4 Test Prompts

Copy and paste these prompts into Claude Desktop:

#### Precision Test

```sh
What is 0.1 + 0.2? Use the calculator tool.
```

*Expected: Claude uses the calculator and returns exactly 0.3*

#### Healthcare Metrics

```csv
Using the calculator tool, compute:
1. Sum these charges: 125000, 89500, 156000, 234000
2. Calculate the average of those same values
3. If we had 15000 in denials, what percentage is that of the total? (divide denials by total, multiply by 100)
```

#### Financial Calculation

```csv
Calculate compound interest: $10,000 invested at 7% annual rate, compounded monthly, for 5 years. Use the calculator's compound_interest operation.
```

#### Chained Calculations

```sh
Using the calculator with named results that reference each other:
1. Name "q1_revenue" = sum of 45000, 52000, 48000
2. Name "q1_costs" = sum of 28000, 31000, 29000
3. Name "q1_profit" = subtract q1_costs from q1_revenue
4. Name "profit_margin" = divide q1_profit by q1_revenue, then multiply by 100

Show me all the results.
```

#### Statistical Analysis

```sql
Calculate statistics for these AR days values: 42, 38, 45, 51, 33, 47, 39

Use the calculator to find:
- average
- min
- max
- stddev (standard deviation)
```

---

## 3. Available Operations

| Operation | Args | Description |
|-----------|------|-------------|
| `add` | [a, b] | a + b |
| `subtract` | [a, b] | a - b |
| `multiply` | [a, b] | a × b |
| `divide` | [a, b] | a ÷ b |
| `sum` | [a, b, c, ...] | Sum of all values |
| `average` | [a, b, c, ...] | Mean of all values |
| `min` | [a, b, c, ...] | Minimum value |
| `max` | [a, b, c, ...] | Maximum value |
| `median` | [a, b, c, ...] | Median value |
| `stddev` | [a, b, c, ...] | Standard deviation |
| `percentage` | [a, b] | Percentage change: ((a-b)/b) × 100 |
| `round` | [value, decimals] | Round to N decimal places |
| `abs` | [value] | Absolute value |
| `ceil` | [value] | Round up |
| `floor` | [value] | Round down |
| `compare` | [a, b] | Returns -1, 0, or 1 |
| `roi` | [gain, cost] | Return on investment |
| `compound_interest` | [principal, rate, years, frequency] | Future value |
| `present_value` | [future, rate, years, frequency] | Present value |

---

## 4. Troubleshooting

### "Invalid or missing API key"

You forgot the `X-API-Key` header. Add:

```sh
-H "X-API-Key: mcp-calc-secret-key-2026"
```

### "Session not found"

The session expired or you're using an invalid session ID. Re-run the `initialize` call to get a new session.

### Claude Desktop doesn't show the calculator

1. Check the config file syntax (valid JSON)
2. Ensure the server is running: `curl -s https://catalyst-mcp-ec2.scootersoft.info/calc/health -H "X-API-Key: mcp-calc-secret-key-2026"`
3. Restart Claude Desktop completely (quit, not just close window)

### Connection refused

The server may be down. Check EC2 instance status:

```bash
# Check EC2 status
AWS_PROFILE=ai-lab aws ec2 describe-instances --instance-ids i-07f420b768f425ed9 \
  --query 'Reservations[0].Instances[0].State.Name' --region us-east-2

# Start EC2 if stopped
AWS_PROFILE=ai-lab aws ec2 start-instances --instance-ids i-07f420b768f425ed9 --region us-east-2

# Check logs via SSM
AWS_PROFILE=ai-lab aws ssm send-command \
  --instance-ids i-07f420b768f425ed9 \
  --document-name "AWS-RunShellScript" \
  --parameters commands="docker logs mcp-calculator --tail 20" \
  --region us-east-2
```

**Local development:** For local testing with mini.local, use `catalyst-mcp.scootersoft.info` instead (requires MCP services running locally).
