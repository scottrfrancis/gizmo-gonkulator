# mcp-calculator

**Precise Arithmetic for AI Agents**

A Model Context Protocol (MCP) server that eliminates AI math errors by providing deterministic arithmetic operations using Python's Decimal module.

## The Problem

LLMs confidently produce wrong arithmetic. This isn't ignorance—language models are text prediction engines, not calculators.

**Real-world examples:**

| Context | AI Output | Correct Value | Error |
|---------|-----------|---------------|-------|
| Financial report | "72% decline" | 26.3% decline | Confused ratio with % change |
| Trend analysis | "trending higher" | trending lower | Incorrect comparison |

**Observed error rate:** 10-20% when AI does math vs. 0% with this tool.

## Installation

```bash
pip install mcp-calculator
```

## Quick Start

### As an MCP Server

Add to your MCP configuration (e.g., Claude Desktop):

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

### As a Python Library

```python
from mcp_calculator import calculate

result = calculate([
    {"name": "total", "operation": "sum", "args": [100, 200, 300]},
    {"name": "average", "operation": "average", "args": [100, 200, 300]},
    {"name": "growth", "operation": "percentage", "args": [150, 100]}
])

print(result)
# {
#   "success": True,
#   "results": {
#     "total": 600.0,
#     "average": 200.0,
#     "growth": 50.0
#   }
# }
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

```python
result = calculate([
    {"name": "oct_rate", "operation": "divide", "args": [2561276, 8]},
    {"name": "sep_rate", "operation": "divide", "args": [8782334, 21]},
    {"name": "change", "operation": "percentage", "args": ["oct_rate", "sep_rate"]}
])
# oct_rate and sep_rate are computed, then used in the percentage calculation
```

## Why Decimal Precision?

```python
# Float (IEEE 754) - imprecise
>>> 0.1 + 0.2
0.30000000000000004

# mcp-calculator (Decimal) - exact
>>> calculate([{"operation": "add", "args": [0.1, 0.2]}])
{"results": {"result_0": 0.3}}
```

For financial applications, floating-point errors are unacceptable.

## Error Handling

Errors are returned per-calculation, not as exceptions:

```python
result = calculate([
    {"name": "valid", "operation": "sum", "args": [1, 2, 3]},
    {"name": "error", "operation": "divide", "args": [100, 0]}
])
# {
#   "success": True,
#   "results": {
#     "valid": 6.0,
#     "error": {"error": "division by zero"}
#   }
# }
```

## Self-Test

```bash
python -m mcp_calculator --test
```

## Development

```bash
# Clone and install
git clone https://github.com/your-org/mcp-calculator
cd mcp-calculator
pip install -e ".[dev]"

# Run tests
pytest tests/

# Run self-test
python -m mcp_calculator --test
```

## License

MIT

***REMOVED***


