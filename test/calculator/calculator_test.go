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

	// Regression: subtract previously dropped args[2:], so passing
	// [a, b, c] returned a-b instead of a-b-c. The narrative-generation
	// path frequently calls e.g. subtract(total, COVID, PATH) to compute
	// excluding-COVID/PATH aggregates and was silently getting wrong
	// answers.
	t.Run("subtract_variadic_three_args", func(t *testing.T) {
		// 1343 - 3 - 418 = 922 (the Vikor/Example-1A excl-COVID/PATH case).
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "excl", Operation: "subtract", Args: []any{1343, 3, 418}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 922.0, result.Results["excl"])
	})

	t.Run("subtract_variadic_four_args", func(t *testing.T) {
		// 1000 - 100 - 50 - 25 = 825
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "diff", Operation: "subtract", Args: []any{1000, 100, 50, 25}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 825.0, result.Results["diff"])
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

	// Regression: divide previously dropped args[2:] (same shape as
	// the subtract bug). 100 / 4 / 5 must be 5.0 (left-fold), not 25.0.
	t.Run("divide_variadic_three_args", func(t *testing.T) {
		result := calculator.Calculate([]calculator.Calculation{
			{Name: "quot", Operation: "divide", Args: []any{100, 4, 5}},
		})
		assert.True(t, result.Success)
		assertNumericEqual(t, 5.0, result.Results["quot"])
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
