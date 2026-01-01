"""
Tests for mcp-calculator.

Run with: pytest tests/
"""

import pytest
from mcp_calculator import calculate, CalculationEngine


class TestBasicOperations:
    """Test basic arithmetic operations."""

    def test_add(self):
        result = calculate([{"name": "sum", "operation": "add", "args": [100, 200]}])
        assert result["success"] is True
        assert result["results"]["sum"] == 300.0

    def test_subtract(self):
        result = calculate([{"name": "diff", "operation": "subtract", "args": [500, 200]}])
        assert result["success"] is True
        assert result["results"]["diff"] == 300.0

    def test_multiply(self):
        result = calculate([{"name": "prod", "operation": "multiply", "args": [10, 5]}])
        assert result["success"] is True
        assert result["results"]["prod"] == 50.0

    def test_divide(self):
        result = calculate([{"name": "quot", "operation": "divide", "args": [100, 4]}])
        assert result["success"] is True
        assert result["results"]["quot"] == 25.0

    def test_sum(self):
        result = calculate([{"name": "total", "operation": "sum", "args": [100, 200, 300]}])
        assert result["success"] is True
        assert result["results"]["total"] == 600.0

    def test_average(self):
        result = calculate([{"name": "avg", "operation": "average", "args": [100, 200, 300]}])
        assert result["success"] is True
        assert result["results"]["avg"] == 200.0


class TestStatisticalOperations:
    """Test statistical operations."""

    def test_min(self):
        result = calculate([{"name": "minimum", "operation": "min", "args": [100, 50, 200, 75]}])
        assert result["success"] is True
        assert result["results"]["minimum"] == 50.0

    def test_max(self):
        result = calculate([{"name": "maximum", "operation": "max", "args": [100, 50, 200, 75]}])
        assert result["success"] is True
        assert result["results"]["maximum"] == 200.0

    def test_median_odd(self):
        result = calculate([{"name": "med", "operation": "median", "args": [1, 3, 2]}])
        assert result["success"] is True
        assert result["results"]["med"] == 2.0

    def test_median_even(self):
        result = calculate([{"name": "med", "operation": "median", "args": [1, 2, 3, 4]}])
        assert result["success"] is True
        assert result["results"]["med"] == 2.5

    def test_stddev(self):
        result = calculate([{"name": "std", "operation": "stddev", "args": [2, 4, 4, 4, 5, 5, 7, 9]}])
        assert result["success"] is True
        assert abs(result["results"]["std"] - 2.0) < 0.01  # stddev of this set is 2


class TestFinancialOperations:
    """Test financial calculations."""

    def test_percentage_increase(self):
        """50% increase: (150 - 100) / 100 * 100 = 50%"""
        result = calculate([{"name": "pct", "operation": "percentage", "args": [150, 100]}])
        assert result["success"] is True
        assert result["results"]["pct"] == 50.0

    def test_percentage_decrease(self):
        """25% decrease: (75 - 100) / 100 * 100 = -25%"""
        result = calculate([{"name": "pct", "operation": "percentage", "args": [75, 100]}])
        assert result["success"] is True
        assert result["results"]["pct"] == -25.0

    def test_roi(self):
        """ROI: (150000 - 100000) / 100000 * 100 = 50%"""
        result = calculate([{"name": "roi", "operation": "roi", "args": [150000, 100000]}])
        assert result["success"] is True
        assert result["results"]["roi"] == 50.0

    def test_compound_interest(self):
        """$1000 at 10% for 2 years = 1000 * 1.1^2 = 1210"""
        result = calculate([{"name": "fv", "operation": "compound_interest", "args": [1000, 0.10, 2]}])
        assert result["success"] is True
        assert abs(result["results"]["fv"] - 1210.0) < 0.01

    def test_present_value(self):
        """PV of $1210 at 10% for 2 years = 1210 / 1.1^2 = 1000"""
        result = calculate([{"name": "pv", "operation": "present_value", "args": [1210, 0.10, 2]}])
        assert result["success"] is True
        assert abs(result["results"]["pv"] - 1000.0) < 0.01


class TestUtilityOperations:
    """Test utility operations."""

    def test_round(self):
        result = calculate([{"name": "rounded", "operation": "round", "args": [3.14159, 2]}])
        assert result["success"] is True
        assert result["results"]["rounded"] == 3.14

    def test_abs_positive(self):
        result = calculate([{"name": "absolute", "operation": "abs", "args": [42]}])
        assert result["success"] is True
        assert result["results"]["absolute"] == 42.0

    def test_abs_negative(self):
        result = calculate([{"name": "absolute", "operation": "abs", "args": [-42]}])
        assert result["success"] is True
        assert result["results"]["absolute"] == 42.0

    def test_ceil(self):
        result = calculate([{"name": "ceiling", "operation": "ceil", "args": [3.2]}])
        assert result["success"] is True
        assert result["results"]["ceiling"] == 4.0

    def test_floor(self):
        result = calculate([{"name": "floored", "operation": "floor", "args": [3.8]}])
        assert result["success"] is True
        assert result["results"]["floored"] == 3.0


class TestComparisonOperations:
    """Test comparison operations."""

    def test_compare_less_than(self):
        result = calculate([{"name": "cmp", "operation": "compare", "args": [5, 10, "<"]}])
        assert result["success"] is True
        assert result["results"]["cmp"] is True

    def test_compare_greater_than(self):
        result = calculate([{"name": "cmp", "operation": "compare", "args": [10, 5, ">"]}])
        assert result["success"] is True
        assert result["results"]["cmp"] is True

    def test_compare_equal(self):
        result = calculate([{"name": "cmp", "operation": "compare", "args": [5, 5, "=="]}])
        assert result["success"] is True
        assert result["results"]["cmp"] is True


class TestBatchOperations:
    """Test batch calculations."""

    def test_multiple_calculations(self):
        result = calculate([
            {"name": "sum", "operation": "sum", "args": [100, 200, 300]},
            {"name": "avg", "operation": "average", "args": [100, 200, 300]},
            {"name": "pct", "operation": "percentage", "args": [150, 100]}
        ])
        assert result["success"] is True
        assert result["results"]["sum"] == 600.0
        assert result["results"]["avg"] == 200.0
        assert result["results"]["pct"] == 50.0


class TestVariableReferences:
    """Test calculations that reference previous results."""

    def test_simple_reference(self):
        result = calculate([
            {"name": "a", "operation": "sum", "args": [10, 20]},
            {"name": "b", "operation": "multiply", "args": ["a", 2]}
        ])
        assert result["success"] is True
        assert result["results"]["a"] == 30.0
        assert result["results"]["b"] == 60.0

    def test_chain_reference(self):
        """Test chained variable references."""
        result = calculate([
            {"name": "oct_rate", "operation": "divide", "args": [2561276, 8]},
            {"name": "sep_rate", "operation": "divide", "args": [8782334, 21]},
            {"name": "change", "operation": "percentage", "args": ["oct_rate", "sep_rate"]}
        ])
        assert result["success"] is True
        assert abs(result["results"]["oct_rate"] - 320159.5) < 0.1
        assert abs(result["results"]["sep_rate"] - 418206.38) < 0.1
        # Change should be negative (oct is less than sep)
        assert result["results"]["change"] < 0


class TestPrecision:
    """Test decimal precision (no floating point errors)."""

    def test_classic_float_error(self):
        """0.1 + 0.2 should equal exactly 0.3"""
        result = calculate([{"name": "sum", "operation": "add", "args": [0.1, 0.2]}])
        assert result["success"] is True
        # With float: 0.30000000000000004
        # With Decimal: exactly 0.3
        assert result["results"]["sum"] == 0.3

    def test_division_precision(self):
        """1/3 should maintain precision"""
        result = calculate([{"name": "third", "operation": "divide", "args": [1, 3]}])
        assert result["success"] is True
        assert abs(result["results"]["third"] - 0.333333333) < 0.0001

    def test_large_numbers(self):
        """Handle numbers in the trillions"""
        result = calculate([
            {"name": "sum", "operation": "sum", "args": [1000000000000, 2000000000000]}
        ])
        assert result["success"] is True
        assert result["results"]["sum"] == 3000000000000.0


class TestErrorHandling:
    """Test error handling."""

    def test_division_by_zero(self):
        result = calculate([{"name": "err", "operation": "divide", "args": [100, 0]}])
        assert result["success"] is True  # Batch succeeds, individual calc has error
        assert "error" in result["results"]["err"]
        assert "division by zero" in result["results"]["err"]["error"]

    def test_percentage_zero_base(self):
        result = calculate([{"name": "err", "operation": "percentage", "args": [150, 0]}])
        assert result["success"] is True
        assert "error" in result["results"]["err"]
        assert "zero base" in result["results"]["err"]["error"]

    def test_unknown_operation(self):
        result = calculate([{"name": "err", "operation": "unknown_op", "args": [1, 2]}])
        assert result["success"] is True
        assert "error" in result["results"]["err"]
        assert "unknown operation" in result["results"]["err"]["error"]

    def test_insufficient_args(self):
        result = calculate([{"name": "err", "operation": "add", "args": [1]}])
        assert result["success"] is True
        assert "error" in result["results"]["err"]


class TestCalculationEngine:
    """Test CalculationEngine class directly."""

    def test_engine_instance(self):
        engine = CalculationEngine()
        result = engine.execute([{"name": "test", "operation": "sum", "args": [1, 2, 3]}])
        assert result["success"] is True
        assert result["results"]["test"] == 6.0

    def test_engine_custom_precision(self):
        engine = CalculationEngine(decimal_places=4)
        result = engine.execute([{"name": "test", "operation": "round", "args": [3.14159265]}])
        # Default round uses engine's decimal_places when not specified
        assert result["success"] is True


class TestRealWorldScenarios:
    """Test real-world calculation scenarios from the original project."""

    def test_healthcare_per_day_rates(self):
        """
        Original error case: AI said "trending higher" when data showed lower.
        October: $2,561,276 / 8 days = $320,159.50/day
        September: $8,782,334 / 21 days = $418,206.38/day
        October rate is LOWER than September.
        """
        result = calculate([
            {"name": "oct_per_day", "operation": "divide", "args": [2561276, 8]},
            {"name": "sep_per_day", "operation": "divide", "args": [8782334, 21]},
            {"name": "is_lower", "operation": "compare", "args": ["oct_per_day", "sep_per_day", "<"]}
        ])
        assert result["success"] is True
        assert abs(result["results"]["oct_per_day"] - 320159.5) < 0.1
        assert abs(result["results"]["sep_per_day"] - 418206.38) < 0.1
        assert result["results"]["is_lower"] is True  # October IS lower

    def test_healthcare_percentage_change(self):
        """
        Original error case: AI said "72% decline" when actual was 26.3%.
        AI confused ratio (51600/70000 = 73.7%) with percentage change.
        Correct: ((51600 - 70000) / 70000) * 100 = -26.3%
        """
        result = calculate([
            {"name": "pct_change", "operation": "percentage", "args": [51600, 70000]}
        ])
        assert result["success"] is True
        # Should be approximately -26.3%, NOT -72%
        assert abs(result["results"]["pct_change"] - (-26.3)) < 0.1
