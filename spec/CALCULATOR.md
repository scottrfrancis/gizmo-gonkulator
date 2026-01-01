# Calculator Engine Specification

## Overview

The calculator engine provides precise arithmetic calculations using arbitrary-precision decimal arithmetic. It eliminates floating-point errors common in IEEE 754 arithmetic, making it suitable for financial and analytical calculations.

## Core Requirements

### CR-1: Decimal Precision

- MUST use arbitrary-precision decimal arithmetic (not IEEE 754 floats)
- MUST handle the classic float error: `0.1 + 0.2` MUST equal exactly `0.3`
- MUST maintain at least 10 decimal places of precision by default
- MUST handle numbers up to 10^15 without precision loss

### CR-2: Batch Operations

- MUST support executing multiple calculations in a single call
- MUST return all results in a single response
- MUST continue processing remaining calculations if one fails
- Individual calculation errors MUST NOT fail the entire batch

### CR-3: Variable References

- Calculations MAY reference results of previous calculations by name
- References MUST be resolved in order of calculation execution
- Referencing an errored result MUST produce an error
- Circular references are not possible (forward references only)

## Operations

### Basic Arithmetic

| Operation | Args | Description | Error Conditions |
|-----------|------|-------------|------------------|
| `add` | `a, b, ...` | Sum of all arguments | Requires >= 2 args |
| `subtract` | `a, b` | `a - b` | Requires exactly 2 args |
| `multiply` | `a, b, ...` | Product of all arguments | Requires >= 2 args |
| `divide` | `a, b` | `a / b` | Requires 2 args, b != 0 |
| `sum` | `a, b, ...` | Sum of all arguments | Requires >= 1 arg |
| `average` | `a, b, ...` | Arithmetic mean | Requires >= 1 arg |

### Statistical Operations

| Operation | Args | Description | Error Conditions |
|-----------|------|-------------|------------------|
| `min` | `a, b, ...` | Minimum value | Requires >= 1 arg |
| `max` | `a, b, ...` | Maximum value | Requires >= 1 arg |
| `median` | `a, b, ...` | Median value | Requires >= 1 arg |
| `stddev` | `a, b, ...` | Population standard deviation | Requires >= 2 args |

### Financial Operations

| Operation | Args | Description | Formula |
|-----------|------|-------------|---------|
| `percentage` | `new, old` | Percentage change | `((new - old) / old) * 100` |
| `roi` | `gain, cost` | Return on investment | `((gain - cost) / cost) * 100` |
| `compound_interest` | `principal, rate, periods` | Future value | `principal * (1 + rate)^periods` |
| `present_value` | `fv, rate, periods` | Present value | `fv / (1 + rate)^periods` |

### Utility Operations

| Operation | Args | Description | Error Conditions |
|-----------|------|-------------|------------------|
| `round` | `value[, places]` | Round to decimal places | Default: engine precision |
| `abs` | `value` | Absolute value | Requires 1 arg |
| `ceil` | `value` | Round up to integer | Requires 1 arg |
| `floor` | `value` | Round down to integer | Requires 1 arg |
| `compare` | `a, b[, op]` | Boolean comparison | ops: `<`, `>`, `<=`, `>=`, `==` |

## Input Format

```json
{
  "calculations": [
    {
      "name": "optional_result_name",
      "operation": "operation_name",
      "args": [1, 2, 3]
    }
  ]
}
```

### Fields

- `name` (optional): Label for the result. Auto-generated as `result_N` if omitted.
- `operation` (required): One of the supported operations.
- `args` (required): Array of numbers or string references to previous results.

## Output Format

```json
{
  "success": true,
  "results": {
    "result_name": 123.45,
    "errored_result": {"error": "division by zero"}
  }
}
```

### Fields

- `success`: Always `true` for batch operations (individual errors reported per-result)
- `results`: Map of result names to values or error objects

## Error Handling

### Error Response Format

```json
{
  "error": "descriptive error message"
}
```

### Error Categories

1. **Validation Errors**: Invalid operation, insufficient arguments
2. **Arithmetic Errors**: Division by zero, invalid percentage base
3. **Reference Errors**: Unknown variable reference, errored reference

## Test Cases (Ported from Python)

### Basic Operations
- `add(100, 200)` = 300.0
- `subtract(500, 200)` = 300.0
- `multiply(10, 5)` = 50.0
- `divide(100, 4)` = 25.0
- `sum(100, 200, 300)` = 600.0
- `average(100, 200, 300)` = 200.0

### Statistical Operations
- `min(100, 50, 200, 75)` = 50.0
- `max(100, 50, 200, 75)` = 200.0
- `median(1, 3, 2)` = 2.0
- `median(1, 2, 3, 4)` = 2.5
- `stddev(2, 4, 4, 4, 5, 5, 7, 9)` ≈ 2.0

### Financial Operations
- `percentage(150, 100)` = 50.0 (50% increase)
- `percentage(75, 100)` = -25.0 (25% decrease)
- `roi(150000, 100000)` = 50.0
- `compound_interest(1000, 0.10, 2)` ≈ 1210.0
- `present_value(1210, 0.10, 2)` ≈ 1000.0

### Precision Tests
- `add(0.1, 0.2)` = 0.3 (exactly, not 0.30000000000000004)
- `divide(1, 3)` ≈ 0.333333333
- `sum(1000000000000, 2000000000000)` = 3000000000000

### Error Cases
- `divide(100, 0)` = error: "division by zero"
- `percentage(150, 0)` = error: contains "zero base"
- Unknown operation = error: contains "unknown operation"
- `add(1)` = error: insufficient args

### Variable References
- Given: `a = sum(10, 20)` then `multiply("a", 2)` = 60.0
- Chain: `oct_rate = divide(2561276, 8)`, `sep_rate = divide(8782334, 21)`,
  `percentage("oct_rate", "sep_rate")` < 0 (negative change)

## Implementation Notes

1. Use `shopspring/decimal` or `github.com/cockroachdb/apd` for Go
2. Thread-safe: Multiple goroutines may call Calculate concurrently
3. No shared state between Calculate calls
4. Result names are case-sensitive
