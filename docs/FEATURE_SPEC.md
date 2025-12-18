# mcp-calculator: Precise Arithmetic for AI Agents

**A Model Context Protocol (MCP) server that eliminates AI math errors**

## Vision

Every AI developer building agentic workflows has encountered this problem: LLMs confidently produce wrong arithmetic. The issue isn't ignorance—it's that language models are fundamentally text prediction engines, not calculators.

`mcp-calculator` provides a production-ready solution: a lightweight MCP server that AI agents invoke for deterministic arithmetic, ensuring 100% accuracy in financial, scientific, and analytical applications.

## The Problem

LLMs make systematic arithmetic errors when performing calculations during text generation:

| Context | AI Output | Correct Value | Error Type |
|---------|-----------|---------------|------------|
| Financial decline | "72% decline" | 26.3% decline | Confused ratio with % change |
| Collections trend | "trending higher" | trending lower | Incorrect comparison |
| Per-day rates | Wrong values | $320k vs $418k | Division errors |

**Observed error rate:** 10-20% in sections requiring AI math vs. 0% with tool-assisted calculations.

## Core Value Proposition

| Without Tool | With mcp-calculator |
|--------------|-------------------|
| 10-20% math error rate | 0% error rate |
| Floating-point imprecision | Decimal precision |
| Inconsistent formatting | Deterministic output |
| No audit trail | Named, traceable results |

---

## Feature Specification

### 1. Core Operations

**Basic Arithmetic**
- `add(a, b, ...)` - Sum any number of values
- `subtract(a, b)` - Difference
- `multiply(a, b, ...)` - Product of values
- `divide(a, b)` - Division with zero-check

**Statistical**
- `sum(values)` - Total of array
- `average(values)` - Arithmetic mean
- `min(values)` - Minimum value
- `max(values)` - Maximum value
- `median(values)` - Median value
- `stddev(values)` - Standard deviation

**Financial**
- `percentage(new, old)` - Percent change: ((new-old)/old)×100
- `compound_interest(principal, rate, periods)` - P(1+r)^n
- `present_value(future_value, rate, periods)` - FV/(1+r)^n
- `roi(gain, cost)` - (gain-cost)/cost × 100

**Comparison**
- `compare(a, b, operator)` - Returns boolean for >, <, >=, <=, ==
- `rank(values)` - Return sorted indices

**Utility**
- `round(value, decimals)` - Round with banker's rounding
- `ceil(value)` - Round up
- `floor(value)` - Round down
- `abs(value)` - Absolute value
- `format_currency(value, currency, locale)` - "$1,234.56"
- `format_percentage(value, decimals)` - "12.34%"

### 2. Batch Operations

Execute multiple calculations in a single call for efficiency:

```json
{
  "calculations": [
    {"name": "total_revenue", "operation": "sum", "args": [1000, 2000, 3000]},
    {"name": "avg_revenue", "operation": "average", "args": [1000, 2000, 3000]},
    {"name": "growth", "operation": "percentage", "args": [3000, 1000]}
  ]
}
```

**Response:**
```json
{
  "success": true,
  "results": {
    "total_revenue": 6000,
    "avg_revenue": 2000,
    "growth": 200.0
  }
}
```

### 3. Variable References

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

This enables complex multi-step calculations without multiple round-trips.

### 4. Error Handling

Graceful error handling with per-calculation error reporting:

```json
{
  "success": true,
  "results": {
    "valid_calc": 100.0,
    "div_by_zero": {"error": "division by zero"},
    "bad_percentage": {"error": "cannot calculate percentage with zero base"}
  }
}
```

### 5. Precision Modes

Support different precision requirements:

```json
{
  "precision": "decimal",     // Default: Python Decimal (arbitrary precision)
  "precision": "float64",     // IEEE 754 double (15-17 significant digits)
  "precision": "currency",    // Fixed 2 decimal places, banker's rounding
  "decimal_places": 10        // Custom precision
}
```

### 6. Transport Support

**STDIO (default)** - For local MCP integrations
```bash
python3 -m mcp_calculator
```

**HTTP/SSE** - For remote/cloud deployments
```bash
python3 -m mcp_calculator --transport http --port 8080
```

**Streamable HTTP** - For modern MCP clients
```bash
python3 -m mcp_calculator --transport streamable-http --port 8080
```

### 7. Unit Conversions (Extension)

Optional module for scientific applications:

```json
{
  "operation": "convert",
  "args": [100, "celsius", "fahrenheit"]
}
// Returns: 212
```

Categories: temperature, length, weight, volume, time, currency (with live rates)

### 8. Expression Evaluation (Safe)

Parse and evaluate mathematical expressions safely (no code execution):

```json
{
  "operation": "eval",
  "expression": "(revenue - cost) / cost * 100",
  "variables": {"revenue": 150000, "cost": 100000}
}
// Returns: 50.0
```

**Safety features:**
- Whitelist of allowed operators
- No function calls except built-in math
- Variable injection only (no arbitrary evaluation)
- Recursion/complexity limits

---

## API Design

### MCP Tool Definition

```json
{
  "name": "calculate",
  "description": "Perform precise arithmetic calculations. Use this tool for ALL math operations - never calculate in your response. Supports: add, subtract, multiply, divide, sum, average, percentage, round, min, max, median, stddev, compare.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "calculations": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "name": {"type": "string", "description": "Label for result"},
            "operation": {"type": "string", "enum": ["add", "subtract", "multiply", "divide", "sum", "average", "percentage", "round", "min", "max", "median", "stddev", "compare", "abs", "ceil", "floor"]},
            "args": {"type": "array", "items": {"oneOf": [{"type": "number"}, {"type": "string"}]}}
          },
          "required": ["operation", "args"]
        }
      },
      "precision": {"type": "string", "enum": ["decimal", "float64", "currency"]},
      "decimal_places": {"type": "integer", "minimum": 0, "maximum": 50}
    },
    "required": ["calculations"]
  }
}
```

---

## Implementation Notes

### Why Decimal, Not Float?

```python
# Float (IEEE 754) - imprecise
>>> 0.1 + 0.2
0.30000000000000004

# Decimal - exact
>>> Decimal('0.1') + Decimal('0.2')
Decimal('0.3')
```

For financial applications, floating-point errors are unacceptable.

### Performance Considerations

- Batch operations minimize round-trip latency
- Variable references avoid redundant calculations
- STDIO transport has ~1ms overhead per call
- HTTP adds ~5-10ms network overhead

### Security

- No arbitrary code execution
- Input validation on all operations
- Bounded output size
- Rate limiting for HTTP transport

---

## Use Cases

1. **Financial Reports** - Revenue analysis, variance calculations, trend detection
2. **Scientific Computing** - Unit conversions, statistical analysis
3. **E-commerce** - Pricing calculations, discount percentages, tax computation
4. **Healthcare Analytics** - Per-day rates, percentage changes, comparisons
5. **Any AI Agent** - Wherever deterministic math is required

---

## Success Metrics

1. **Adoption**: Downloads, GitHub stars, MCP registry usage
2. **Accuracy**: Zero reported calculation errors in production
3. **Performance**: <5ms p99 latency for batch operations
4. **Integration**: Support in Claude Desktop, Claude Code, third-party agents
