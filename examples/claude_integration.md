# Integrating mcp-calculator with Claude

This guide shows how to configure mcp-calculator as an MCP server for Claude Desktop or Claude Code.

## Claude Desktop Configuration

Add to your Claude Desktop MCP configuration file:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "calculator": {
      "command": "python",
      "args": ["-m", "mcp_calculator"]
    }
  }
}
```

## Claude Code Configuration

Add to `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "calculator": {
      "command": "python3",
      "args": ["-m", "mcp_calculator"]
    }
  }
}
```

## Usage in Prompts

Once configured, you can instruct Claude to use the calculator:

```
When performing any arithmetic in your response, use the calculate tool
instead of computing values yourself. This ensures 100% accuracy.
```

### Example Prompt

```
Analyze these collections figures and calculate the per-day rate and percentage change:
- October: $2,561,276 over 8 business days
- September: $8,782,334 over 21 business days

Use the calculate tool for all arithmetic.
```

### Expected Tool Call

Claude will invoke:

```json
{
  "name": "calculate",
  "arguments": {
    "calculations": [
      {"name": "oct_rate", "operation": "divide", "args": [2561276, 8]},
      {"name": "sep_rate", "operation": "divide", "args": [8782334, 21]},
      {"name": "change", "operation": "percentage", "args": ["oct_rate", "sep_rate"]}
    ]
  }
}
```

### Tool Response

```json
{
  "success": true,
  "results": {
    "oct_rate": 320159.5,
    "sep_rate": 418206.38095238095,
    "change": -23.44...
  }
}
```

## Best Practices

1. **Always use the tool for financial calculations** - Even simple arithmetic can have floating-point errors
2. **Name your results** - Makes it easier to reference in follow-up calculations
3. **Batch calculations** - Send multiple related calculations in one call
4. **Reference previous results** - Use variable names to build complex calculations

## Troubleshooting

### Tool not appearing

1. Verify mcp-calculator is installed: `pip show mcp-calculator`
2. Test the server: `python -m mcp_calculator --test`
3. Check your MCP configuration file syntax

### Permission errors

Ensure the python executable is in your PATH and accessible.

### Debug mode

Set logging level to DEBUG in the server for troubleshooting:

```python
import logging
logging.getLogger("mcp_calculator").setLevel(logging.DEBUG)
```
