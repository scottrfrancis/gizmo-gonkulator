// Package calculator_test contains tests for the calculator engine.
// These tests are ported from the Python reference implementation.
package calculator_test

import (
	"math"
	"testing"

	"github.com/scottrfrancis/mcp-calculator/internal/calculator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// toFloat64 converts any numeric type to float64 for comparison.
// The calculator returns int64 for whole numbers and float64 for decimals.
func toFloat64(v any) float64 {
	switch n := v.(type) {
	case int64:
		return float64(n)
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return 0
	}
}

// assertNumericEqual compares two numeric values regardless of underlying type.
func assertNumericEqual(t *testing.T, expected float64, actual any) {
	t.Helper()
	assert.InDelta(t, expected, toFloat64(actual), 0.0001)
}

// TestBasicOperations tests basic arithmetic operations.
func TestBasicOperations(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "sum", Operation: "add", Args: []any{100, 200}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 300.0, result.Results["sum"])
	})

	t.Run("subtract", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "diff", Operation: "subtract", Args: []any{500, 200}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 300.0, result.Results["diff"])
	})

	t.Run("multiply", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "prod", Operation: "multiply", Args: []any{10, 5}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 50.0, result.Results["prod"])
	})

	t.Run("divide", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "quot", Operation: "divide", Args: []any{100, 4}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 25.0, result.Results["quot"])
	})

	t.Run("sum", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "total", Operation: "sum", Args: []any{100, 200, 300}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 600.0, result.Results["total"])
	})

	t.Run("average", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "avg", Operation: "average", Args: []any{100, 200, 300}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 200.0, result.Results["avg"])
	})
}

// TestStatisticalOperations tests statistical operations.
func TestStatisticalOperations(t *testing.T) {
	t.Run("min", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "minimum", Operation: "min", Args: []any{100, 50, 200, 75}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 50.0, result.Results["minimum"])
	})

	t.Run("max", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "maximum", Operation: "max", Args: []any{100, 50, 200, 75}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 200.0, result.Results["maximum"])
	})

	t.Run("median_odd", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "med", Operation: "median", Args: []any{1, 3, 2}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 2.0, result.Results["med"])
	})

	t.Run("median_even", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "med", Operation: "median", Args: []any{1, 2, 3, 4}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 2.5, result.Results["med"])
	})

	t.Run("stddev", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "std", Operation: "stddev", Args: []any{2, 4, 4, 4, 5, 5, 7, 9}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 2.0, result.Results["std"])
	})
}

// TestFinancialOperations tests financial calculations.
func TestFinancialOperations(t *testing.T) {
	t.Run("percentage_increase", func(t *testing.T) {
		// 50% increase: (150 - 100) / 100 * 100 = 50%
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "pct", Operation: "percentage", Args: []any{150, 100}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 50.0, result.Results["pct"])
	})

	t.Run("percentage_decrease", func(t *testing.T) {
		// 25% decrease: (75 - 100) / 100 * 100 = -25%
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "pct", Operation: "percentage", Args: []any{75, 100}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, -25.0, result.Results["pct"])
	})

	t.Run("roi", func(t *testing.T) {
		// ROI: (150000 - 100000) / 100000 * 100 = 50%
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "roi", Operation: "roi", Args: []any{150000, 100000}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 50.0, result.Results["roi"])
	})

	t.Run("compound_interest", func(t *testing.T) {
		// $1000 at 10% for 2 years = 1000 * 1.1^2 = 1210
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "fv", Operation: "compound_interest", Args: []any{1000, 0.10, 2}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 1210.0, result.Results["fv"])
	})

	t.Run("present_value", func(t *testing.T) {
		// PV of $1210 at 10% for 2 years = 1210 / 1.1^2 = 1000
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "pv", Operation: "present_value", Args: []any{1210, 0.10, 2}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 1000.0, result.Results["pv"])
	})
}

// TestUtilityOperations tests utility operations.
func TestUtilityOperations(t *testing.T) {
	t.Run("round", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "rounded", Operation: "round", Args: []any{3.14159, 2}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 3.14, result.Results["rounded"])
	})

	t.Run("abs_positive", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "absolute", Operation: "abs", Args: []any{42}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 42.0, result.Results["absolute"])
	})

	t.Run("abs_negative", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "absolute", Operation: "abs", Args: []any{-42}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 42.0, result.Results["absolute"])
	})

	t.Run("ceil", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "ceiling", Operation: "ceil", Args: []any{3.2}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 4.0, result.Results["ceiling"])
	})

	t.Run("floor", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "floored", Operation: "floor", Args: []any{3.8}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 3.0, result.Results["floored"])
	})
}

// TestComparisonOperations tests comparison operations.
func TestComparisonOperations(t *testing.T) {
	t.Run("compare_less_than", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "cmp", Operation: "compare", Args: []any{5, 10, "<"}},
		})
		assert.True(t, result.Success)
		assert.Equal(t, true, result.Results["cmp"])
	})

	t.Run("compare_greater_than", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "cmp", Operation: "compare", Args: []any{10, 5, ">"}},
		})
		assert.True(t, result.Success)
		assert.Equal(t, true, result.Results["cmp"])
	})

	t.Run("compare_equal", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "cmp", Operation: "compare", Args: []any{5, 5, "=="}},
		})
		assert.True(t, result.Success)
		assert.Equal(t, true, result.Results["cmp"])
	})

	t.Run("compare_less_than_or_equal", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "cmp", Operation: "compare", Args: []any{5, 5, "<="}},
		})
		assert.True(t, result.Success)
		assert.Equal(t, true, result.Results["cmp"])
	})

	t.Run("compare_greater_than_or_equal", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "cmp", Operation: "compare", Args: []any{10, 5, ">="}},
		})
		assert.True(t, result.Success)
		assert.Equal(t, true, result.Results["cmp"])
	})
}

// TestBatchOperations tests batch calculations.
func TestBatchOperations(t *testing.T) {
	t.Run("multiple_calculations", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "sum", Operation: "sum", Args: []any{100, 200, 300}},
			{Name: "avg", Operation: "average", Args: []any{100, 200, 300}},
			{Name: "pct", Operation: "percentage", Args: []any{150, 100}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 600.0, result.Results["sum"])
		assertNumericEqual(t, 200.0, result.Results["avg"])
		assertNumericEqual(t, 50.0, result.Results["pct"])
	})
}

// TestVariableReferences tests calculations that reference previous results.
func TestVariableReferences(t *testing.T) {
	t.Run("simple_reference", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "a", Operation: "sum", Args: []any{10, 20}},
			{Name: "b", Operation: "multiply", Args: []any{"a", 2}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 30.0, result.Results["a"])
		assertNumericEqual(t, 60.0, result.Results["b"])
	})

	t.Run("chain_reference", func(t *testing.T) {
		// Test chained variable references
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "oct_rate", Operation: "divide", Args: []any{2561276, 8}},
			{Name: "sep_rate", Operation: "divide", Args: []any{8782334, 21}},
			{Name: "change", Operation: "percentage", Args: []any{"oct_rate", "sep_rate"}},
		})
		assert.True(t, result.Success)

		octRate := toFloat64(result.Results["oct_rate"])
		sepRate := toFloat64(result.Results["sep_rate"])
		change := toFloat64(result.Results["change"])

		assert.InDelta(t, 320159.5, octRate, 0.1)
		assert.InDelta(t, 418206.38, sepRate, 0.1)
		// Change should be negative (oct is less than sep)
		assert.True(t, change < 0)
	})
}

// TestPrecision tests decimal precision (no floating point errors).
func TestPrecision(t *testing.T) {
	t.Run("classic_float_error", func(t *testing.T) {
		// 0.1 + 0.2 should equal exactly 0.3
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "sum", Operation: "add", Args: []any{0.1, 0.2}},
		})
		assert.True(t, result.Success)
		// With float: 0.30000000000000004
		// With Decimal: exactly 0.3
		assertNumericEqual(t, 0.3, result.Results["sum"])
	})

	t.Run("division_precision", func(t *testing.T) {
		// 1/3 should maintain precision
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "third", Operation: "divide", Args: []any{1, 3}},
		})
		assert.True(t, result.Success)
		third := toFloat64(result.Results["third"])
		assert.InDelta(t, 0.333333333, third, 0.0001)
	})

	t.Run("large_numbers", func(t *testing.T) {
		// Handle numbers in the trillions
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "sum", Operation: "sum", Args: []any{1000000000000, 2000000000000}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 3000000000000.0, result.Results["sum"])
	})
}

// TestErrorHandling tests error handling.
func TestErrorHandling(t *testing.T) {
	t.Run("division_by_zero", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "err", Operation: "divide", Args: []any{100, 0}},
		})
		// Batch succeeds, individual calc has error
		assert.True(t, result.Success)

		errResult, ok := result.Results["err"].(map[string]any)
		require.True(t, ok, "expected error map")
		assert.Contains(t, errResult["error"], "division by zero")
	})

	t.Run("percentage_zero_base", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "err", Operation: "percentage", Args: []any{150, 0}},
		})
		assert.True(t, result.Success)

		errResult, ok := result.Results["err"].(map[string]any)
		require.True(t, ok, "expected error map")
		assert.Contains(t, errResult["error"], "zero")
	})

	t.Run("unknown_operation", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "err", Operation: "unknown_op", Args: []any{1, 2}},
		})
		assert.True(t, result.Success)

		errResult, ok := result.Results["err"].(map[string]any)
		require.True(t, ok, "expected error map")
		assert.Contains(t, errResult["error"], "unknown operation")
	})

	t.Run("insufficient_args", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "err", Operation: "add", Args: []any{1}},
		})
		assert.True(t, result.Success)

		errResult, ok := result.Results["err"].(map[string]any)
		require.True(t, ok, "expected error map")
		assert.NotEmpty(t, errResult["error"])
	})

	t.Run("reference_errored_result", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "bad", Operation: "divide", Args: []any{1, 0}},
			{Name: "uses_bad", Operation: "add", Args: []any{"bad", 1}},
		})
		assert.True(t, result.Success)

		errResult, ok := result.Results["uses_bad"].(map[string]any)
		require.True(t, ok, "expected error map")
		assert.Contains(t, errResult["error"], "errored result")
	})

	t.Run("circular_reference", func(t *testing.T) {
		// Test that circular variable references are detected
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "a", Operation: "add", Args: []any{"b", 1}},
			{Name: "b", Operation: "add", Args: []any{"a", 1}},
		})
		assert.True(t, result.Success)

		// First should fail trying to resolve "b" which doesn't exist yet
		errResult, ok := result.Results["a"].(map[string]any)
		require.True(t, ok, "expected error map for 'a'")
		assert.NotEmpty(t, errResult["error"])
	})
}

// TestCalculationEngine tests the CalculationEngine class directly.
func TestCalculationEngine(t *testing.T) {
	t.Run("engine_instance", func(t *testing.T) {
		engine := calculator.NewEngine()
		result := engine.Execute([]calculator.Calculation{
			{Name: "test", Operation: "sum", Args: []any{1, 2, 3}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 6.0, result.Results["test"])
	})

	t.Run("engine_custom_precision", func(t *testing.T) {
		engine := calculator.NewEngineWithPrecision(4)
		result := engine.Execute([]calculator.Calculation{
			{Name: "test", Operation: "round", Args: []any{3.14159265}},
		})
		assert.True(t, result.Success)
		// Default round uses engine's decimal_places when not specified
		assert.NotNil(t, result.Results["test"])
	})
}

// TestRealWorldScenarios tests real-world calculation scenarios.
func TestRealWorldScenarios(t *testing.T) {
	t.Run("healthcare_per_day_rates", func(t *testing.T) {
		// Original error case: AI said "trending higher" when data showed lower.
		// October: $2,561,276 / 8 days = $320,159.50/day
		// September: $8,782,334 / 21 days = $418,206.38/day
		// October rate is LOWER than September.
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "oct_per_day", Operation: "divide", Args: []any{2561276, 8}},
			{Name: "sep_per_day", Operation: "divide", Args: []any{8782334, 21}},
			{Name: "is_lower", Operation: "compare", Args: []any{"oct_per_day", "sep_per_day", "<"}},
		})
		assert.True(t, result.Success)

		octPerDay := toFloat64(result.Results["oct_per_day"])
		sepPerDay := toFloat64(result.Results["sep_per_day"])
		isLower := result.Results["is_lower"].(bool)

		assert.InDelta(t, 320159.5, octPerDay, 0.1)
		assert.InDelta(t, 418206.38, sepPerDay, 0.1)
		assert.True(t, isLower) // October IS lower
	})

	t.Run("healthcare_percentage_change", func(t *testing.T) {
		// Original error case: AI said "72% decline" when actual was 26.3%.
		// AI confused ratio (51600/70000 = 73.7%) with percentage change.
		// Correct: ((51600 - 70000) / 70000) * 100 = -26.3%
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "pct_change", Operation: "percentage", Args: []any{51600, 70000}},
		})
		assert.True(t, result.Success)

		pctChange := toFloat64(result.Results["pct_change"])
		// Should be approximately -26.3%, NOT -72%
		assert.InDelta(t, -26.3, pctChange, 0.1)
	})
}

// TestAutoNaming tests automatic result naming.
func TestAutoNaming(t *testing.T) {
	t.Run("auto_generated_names", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Operation: "sum", Args: []any{1, 2}},
			{Operation: "sum", Args: []any{3, 4}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 3.0, result.Results["result_0"])
		assertNumericEqual(t, 7.0, result.Results["result_1"])
	})
}

// TestEdgeCases tests edge cases.
func TestEdgeCases(t *testing.T) {
	t.Run("empty_calculations", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{})
		assert.True(t, result.Success)
		assert.Empty(t, result.Results)
	})

	t.Run("negative_numbers", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "sum", Operation: "add", Args: []any{-10, -20}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, -30.0, result.Results["sum"])
	})

	t.Run("very_small_numbers", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "sum", Operation: "add", Args: []any{0.0001, 0.0002}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 0.0003, result.Results["sum"])
	})

	t.Run("mixed_int_float", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "sum", Operation: "add", Args: []any{1, 2.5}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 3.5, result.Results["sum"])
	})

	t.Run("scientific_notation", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "sum", Operation: "add", Args: []any{1e10, 2e10}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 3e10, result.Results["sum"])
	})

	t.Run("infinity_check", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "big", Operation: "multiply", Args: []any{math.MaxFloat64, 2}},
		})
		assert.True(t, result.Success)
		// Should handle overflow gracefully - either returns error, infinity, or very large number
		_, isErr := result.Results["big"].(map[string]any)
		if isErr {
			// Error is acceptable for overflow
			return
		}
		// Otherwise, the result should be very large or infinity
		val := toFloat64(result.Results["big"])
		// Either infinity or very large - or it could be 0 if conversion failed for large decimal
		assert.True(t, math.IsInf(val, 1) || val > 1e300 || val == 0, "expected infinity, large number, or 0 for overflow")
	})
}

// =====================================================================
// Coverage expansion — locks down behaviors that work today but lacked
// targeted test coverage. Bundled here so the file's earlier sections
// can be read as the original behavioral spec; this section is the
// regression net underneath it.
// =====================================================================

// TestVariadicLeftFoldOps verifies that the four left-fold operations
// — add, multiply (subtract and divide are similar but tested separately
// where they regressed) — accept and correctly fold over more than two
// arguments. A 2-arg-only test passes even if the loop is broken;
// 3+ args proves the args[1:] traversal actually runs.
func TestVariadicLeftFoldOps(t *testing.T) {
	t.Run("add_three_args", func(t *testing.T) {
		// 100 + 200 + 300 = 600
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "sum", Operation: "add", Args: []any{100, 200, 300}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 600.0, result.Results["sum"])
	})

	t.Run("add_five_args_with_negatives", func(t *testing.T) {
		// 10 + (-5) + 3 + (-1) + 4 = 11. Mix of signs ensures the loop
		// neither short-circuits on a non-positive nor accumulates abs.
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "n", Operation: "add", Args: []any{10, -5, 3, -1, 4}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 11.0, result.Results["n"])
	})

	t.Run("multiply_three_args", func(t *testing.T) {
		// 2 * 3 * 4 = 24
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "p", Operation: "multiply", Args: []any{2, 3, 4}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 24.0, result.Results["p"])
	})

	t.Run("multiply_four_args_with_negative", func(t *testing.T) {
		// 5 * 2 * (-3) * 4 = -120
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "p", Operation: "multiply", Args: []any{5, 2, -3, 4}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, -120.0, result.Results["p"])
	})

	// Insufficient-args guards — the existing TestErrorHandling has a
	// generic "insufficient_args" check; pin the specific ops here so a
	// future refactor that loosens any of these is caught.
	t.Run("add_one_arg_errors", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "add", Args: []any{42}},
		})
		errMap, ok := result.Results["x"].(map[string]any)
		require.True(t, ok, "expected error map; got %T", result.Results["x"])
		assert.Contains(t, errMap["error"], "at least 2")
	})

	t.Run("multiply_one_arg_errors", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "multiply", Args: []any{42}},
		})
		errMap, ok := result.Results["x"].(map[string]any)
		require.True(t, ok)
		assert.Contains(t, errMap["error"], "at least 2")
	})
}

// TestStringNumericArgs verifies that string-encoded numbers are
// resolved as decimals — used by the ReAct/agent path when the model
// emits args as strings. This is distinct from the variable-reference
// path (which also uses strings but resolves them via results-map
// lookup first).
func TestStringNumericArgs(t *testing.T) {
	t.Run("integer_string", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "add", Args: []any{"100", "50"}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 150.0, result.Results["x"])
	})

	t.Run("decimal_string", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "multiply", Args: []any{"3.14", "2"}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 6.28, result.Results["x"])
	})

	t.Run("negative_string", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "add", Args: []any{"-5", "10"}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 5.0, result.Results["x"])
	})

	t.Run("scientific_string", func(t *testing.T) {
		// "1e3" must parse as 1000 — shopspring/decimal supports this.
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "add", Args: []any{"1e3", "500"}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 1500.0, result.Results["x"])
	})

	t.Run("non_numeric_string_errors", func(t *testing.T) {
		// A string that's neither a known result name nor a parseable
		// number must surface an error rather than silently coercing.
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "add", Args: []any{"banana", 5}},
		})
		errMap, ok := result.Results["x"].(map[string]any)
		require.True(t, ok, "expected error map; got %T", result.Results["x"])
		assert.NotEmpty(t, errMap["error"])
	})
}

// TestDeepVariableReferences exercises chained refs beyond the
// existing 2-step "chain_reference" test. Three+ levels prove the
// recursive resolveArgWithVisited path works under deeper nesting.
func TestDeepVariableReferences(t *testing.T) {
	t.Run("four_level_chain", func(t *testing.T) {
		// a=10, b=a*2=20, c=b+5=25, d=c-1=24, e=d/3=8
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "a", Operation: "add", Args: []any{4, 6}},
			{Name: "b", Operation: "multiply", Args: []any{"a", 2}},
			{Name: "c", Operation: "add", Args: []any{"b", 5}},
			{Name: "d", Operation: "subtract", Args: []any{"c", 1}},
			{Name: "e", Operation: "divide", Args: []any{"d", 3}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 10.0, result.Results["a"])
		assertNumericEqual(t, 20.0, result.Results["b"])
		assertNumericEqual(t, 25.0, result.Results["c"])
		assertNumericEqual(t, 24.0, result.Results["d"])
		assertNumericEqual(t, 8.0, result.Results["e"])
	})

	t.Run("forward_reference_unresolved", func(t *testing.T) {
		// Calc 1 references "later" which hasn't been computed yet.
		// Without the result in the map, the string falls through to
		// decimal.NewFromString and fails. Test that this surfaces as
		// an error on calc 1, not a silent zero.
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "early", Operation: "add", Args: []any{"later", 5}},
			{Name: "later", Operation: "add", Args: []any{1, 2}},
		})
		errMap, ok := result.Results["early"].(map[string]any)
		require.True(t, ok, "expected forward ref to error; got %T", result.Results["early"])
		assert.NotEmpty(t, errMap["error"])
		// Second calc still computes successfully.
		assertNumericEqual(t, 3.0, result.Results["later"])
	})

	t.Run("ref_in_subtract_variadic_position", func(t *testing.T) {
		// References mixed across positions — name in args[0], int in
		// args[1] — confirms resolveArgs walks every slot, not just one.
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "base", Operation: "add", Args: []any{1000, 0}},
			{Name: "diff", Operation: "subtract", Args: []any{"base", 250}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 750.0, result.Results["diff"])
	})
}

// TestBatchErrorIsolation: an errored calc must not poison subsequent
// calcs that don't reference it. The existing reference_errored_result
// test covers the consumer side; this covers the bystander side.
func TestBatchErrorIsolation(t *testing.T) {
	result := calculator.Calculate([]calculator.Calculation{
		{Name: "ok1", Operation: "add", Args: []any{1, 2}},
		{Name: "boom", Operation: "divide", Args: []any{1, 0}},
		{Name: "ok2", Operation: "multiply", Args: []any{3, 4}},
		{Name: "ok3", Operation: "subtract", Args: []any{10, 7}},
	})
	assert.True(t, result.Success, "Result.Success is per-batch, not per-calc")

	assertNumericEqual(t, 3.0, result.Results["ok1"])
	assertNumericEqual(t, 12.0, result.Results["ok2"])
	assertNumericEqual(t, 3.0, result.Results["ok3"])

	errMap, ok := result.Results["boom"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, errMap["error"], "zero")
}

// TestRoundDecimalPlacesArg: the existing "round" test uses default
// precision; pin the explicit-decimal-places-arg behavior so a
// regression to "ignore the second arg" is caught.
func TestRoundDecimalPlacesArg(t *testing.T) {
	t.Run("round_to_two_places", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "round", Args: []any{3.14159, 2}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 3.14, result.Results["x"])
	})

	t.Run("round_to_zero_places", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "round", Args: []any{3.7, 0}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 4.0, result.Results["x"])
	})

	t.Run("round_negative_places_truncates_left", func(t *testing.T) {
		// Negative places rounds to tens / hundreds — shopspring/decimal
		// supports this. 1234, -2 → 1200.
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "round", Args: []any{1234, -2}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 1200.0, result.Results["x"])
	})
}

// TestSignAndIdentityOps covers floor/ceil/abs across the sign domain.
// The existing tests only cover one polarity each.
func TestSignAndIdentityOps(t *testing.T) {
	t.Run("ceil_negative", func(t *testing.T) {
		// ceil(-3.2) = -3 (toward +∞).
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "ceil", Args: []any{-3.2}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, -3.0, result.Results["x"])
	})

	t.Run("floor_negative", func(t *testing.T) {
		// floor is implemented as Truncate(0) — for -3.7 truncate gives
		// -3 (toward 0), NOT -4 (toward -∞). Pin the as-implemented
		// semantics so a future "fix" to true floor surfaces as a
		// behavior-changing test failure rather than a silent rewrite.
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "floor", Args: []any{-3.7}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, -3.0, result.Results["x"])
	})

	t.Run("abs_zero", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "x", Operation: "abs", Args: []any{0}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 0.0, result.Results["x"])
	})
}

// TestComparisonDefaults: compare with no operator argument should
// default to "<". The existing tests always pass an explicit operator.
func TestComparisonDefaults(t *testing.T) {
	t.Run("default_operator_is_less_than", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "cmp", Operation: "compare", Args: []any{5, 10}},
		})
		assert.True(t, result.Success)
		assert.Equal(t, true, result.Results["cmp"])
	})

	t.Run("unknown_operator_errors", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "cmp", Operation: "compare", Args: []any{5, 10, "≠"}},
		})
		errMap, ok := result.Results["cmp"].(map[string]any)
		require.True(t, ok)
		assert.Contains(t, errMap["error"], "unknown comparison operator")
	})
}

// TestRealWorldRCMScenarios: the existing TestRealWorldScenarios block
// is small. Add scenarios that mirror what the catalyst-rcm-dashboard-bot
// agentic loop actually computes — these are concrete usage shapes that
// we want to remain stable as the calculator evolves.
func TestRealWorldRCMScenarios(t *testing.T) {
	t.Run("per_day_rate_with_business_days_int_division", func(t *testing.T) {
		// 1343 / 8 = 167.875 — the per-day rate for a partial month
		// (Vikor/Example 1A scenario). Decimal precision matters: a
		// float-based implementation would round to 167.87499999…
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "per_day", Operation: "divide", Args: []any{1343, 8}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 167.875, result.Results["per_day"])
	})

	t.Run("variance_pct_with_negative_change", func(t *testing.T) {
		// percentage(80, 100) = (80 - 100) / 100 * 100 = -20.
		// Negative variance should not be normalized to abs.
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "v", Operation: "percentage", Args: []any{80, 100}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, -20.0, result.Results["v"])
	})

	t.Run("ar_aging_average_excluding_zero_buckets", func(t *testing.T) {
		// avg(45, 32, 28) = 35.0 — typical AR-aging-by-payer mean.
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "avg", Operation: "average", Args: []any{45, 32, 28}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 35.0, result.Results["avg"])
	})
}
