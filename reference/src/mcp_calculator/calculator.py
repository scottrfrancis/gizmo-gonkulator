"""
Core calculation engine using Python's Decimal module for precision.

This module provides the calculation logic independent of the MCP transport layer,
enabling use as a library or through the MCP server.
"""

from decimal import Decimal, ROUND_HALF_UP, InvalidOperation
from typing import Any, Dict, List, Optional, Union
import math


class CalculationEngine:
    """
    Precise arithmetic calculation engine using Python's Decimal module.

    Supports batch operations with variable references, enabling complex
    multi-step calculations in a single call.

    Example:
        engine = CalculationEngine()
        result = engine.execute([
            {"name": "oct_rate", "operation": "divide", "args": [2561276, 8]},
            {"name": "sep_rate", "operation": "divide", "args": [8782334, 21]},
            {"name": "change", "operation": "percentage", "args": ["oct_rate", "sep_rate"]}
        ])
        # Returns: {"oct_rate": 320159.5, "sep_rate": 418206.38..., "change": -23.44...}
    """

    def __init__(self, decimal_places: int = 10):
        """
        Initialize the calculation engine.

        Args:
            decimal_places: Default precision for rounding operations
        """
        self.decimal_places = decimal_places
        self._results: Dict[str, Any] = {}

    def execute(self, calculations: List[Dict[str, Any]]) -> Dict[str, Any]:
        """
        Execute a batch of calculations.

        Args:
            calculations: List of calculation specs, each with:
                - name: Optional label for the result
                - operation: Operation to perform
                - args: Arguments (numbers or references to previous results)

        Returns:
            Dict with "success" flag and "results" dict of named values
        """
        self._results = {}

        for i, calc in enumerate(calculations):
            name = calc.get("name", f"result_{i}")
            operation = calc.get("operation", "").lower()
            args = calc.get("args", [])

            try:
                # Resolve variable references
                resolved_args = self._resolve_args(args)

                # Convert to Decimal for precision
                decimal_args = [Decimal(str(a)) for a in resolved_args]

                # Execute operation
                result = self._execute_operation(operation, decimal_args)

                # Store result (convert to float for JSON serialization)
                self._results[name] = float(result) if isinstance(result, Decimal) else result

            except (InvalidOperation, ValueError, ZeroDivisionError) as e:
                self._results[name] = {"error": str(e)}

        return {
            "success": True,
            "results": self._results
        }

    def _resolve_args(self, args: List[Any]) -> List[Union[int, float]]:
        """Resolve variable references in arguments."""
        resolved = []
        for arg in args:
            if isinstance(arg, str) and arg in self._results:
                value = self._results[arg]
                if isinstance(value, dict) and "error" in value:
                    raise ValueError(f"Cannot use errored result '{arg}'")
                resolved.append(value)
            else:
                resolved.append(arg)
        return resolved

    def _execute_operation(self, operation: str, args: List[Decimal]) -> Union[Decimal, bool]:
        """Execute a single operation."""

        if operation == "add":
            if len(args) < 2:
                raise ValueError("add requires at least 2 arguments")
            return sum(args)

        elif operation == "subtract":
            if len(args) < 2:
                raise ValueError("subtract requires at least 2 arguments")
            return args[0] - args[1]

        elif operation == "multiply":
            if len(args) < 2:
                raise ValueError("multiply requires at least 2 arguments")
            result = args[0]
            for arg in args[1:]:
                result *= arg
            return result

        elif operation == "divide":
            if len(args) < 2:
                raise ValueError("divide requires at least 2 arguments")
            if args[1] == 0:
                raise ValueError("division by zero")
            return args[0] / args[1]

        elif operation == "sum":
            return sum(args)

        elif operation == "average":
            if not args:
                raise ValueError("average requires at least 1 argument")
            return sum(args) / len(args)

        elif operation == "min":
            if not args:
                raise ValueError("min requires at least 1 argument")
            return min(args)

        elif operation == "max":
            if not args:
                raise ValueError("max requires at least 1 argument")
            return max(args)

        elif operation == "percentage":
            # percentage(new, old) = ((new - old) / old) * 100
            if len(args) < 2:
                raise ValueError("percentage requires 2 arguments: new, old")
            new_val, old_val = args[0], args[1]
            if old_val == 0:
                raise ValueError("cannot calculate percentage with zero base")
            return ((new_val - old_val) / old_val) * 100

        elif operation == "round":
            if not args:
                raise ValueError("round requires at least 1 argument")
            decimals = int(args[1]) if len(args) > 1 else self.decimal_places
            return args[0].quantize(
                Decimal(10) ** -decimals, rounding=ROUND_HALF_UP
            )

        elif operation == "abs":
            if not args:
                raise ValueError("abs requires 1 argument")
            return abs(args[0])

        elif operation == "ceil":
            if not args:
                raise ValueError("ceil requires 1 argument")
            return Decimal(math.ceil(args[0]))

        elif operation == "floor":
            if not args:
                raise ValueError("floor requires 1 argument")
            return Decimal(math.floor(args[0]))

        elif operation == "compare":
            if len(args) < 2:
                raise ValueError("compare requires at least 2 arguments")
            a, b = args[0], args[1]
            op = str(args[2]) if len(args) > 2 else "<"

            if op == "<":
                return a < b
            elif op == ">":
                return a > b
            elif op == "<=":
                return a <= b
            elif op == ">=":
                return a >= b
            elif op == "==" or op == "=":
                return a == b
            else:
                raise ValueError(f"unknown comparison operator: {op}")

        elif operation == "median":
            if not args:
                raise ValueError("median requires at least 1 argument")
            sorted_args = sorted(args)
            n = len(sorted_args)
            mid = n // 2
            if n % 2 == 0:
                return (sorted_args[mid - 1] + sorted_args[mid]) / 2
            return sorted_args[mid]

        elif operation == "stddev":
            if len(args) < 2:
                raise ValueError("stddev requires at least 2 arguments")
            mean = sum(args) / len(args)
            variance = sum((x - mean) ** 2 for x in args) / len(args)
            return Decimal(str(math.sqrt(float(variance))))

        elif operation == "roi":
            # roi(gain, cost) = (gain - cost) / cost * 100
            if len(args) < 2:
                raise ValueError("roi requires 2 arguments: gain, cost")
            gain, cost = args[0], args[1]
            if cost == 0:
                raise ValueError("cannot calculate ROI with zero cost")
            return ((gain - cost) / cost) * 100

        elif operation == "compound_interest":
            # compound_interest(principal, rate, periods) = principal * (1 + rate) ^ periods
            if len(args) < 3:
                raise ValueError("compound_interest requires 3 arguments: principal, rate, periods")
            principal, rate, periods = args[0], args[1], int(args[2])
            return principal * ((1 + rate) ** periods)

        elif operation == "present_value":
            # present_value(future_value, rate, periods) = fv / (1 + rate) ^ periods
            if len(args) < 3:
                raise ValueError("present_value requires 3 arguments: future_value, rate, periods")
            fv, rate, periods = args[0], args[1], int(args[2])
            return fv / ((1 + rate) ** periods)

        else:
            raise ValueError(f"unknown operation: {operation}")


def calculate(calculations: List[Dict[str, Any]]) -> Dict[str, Any]:
    """
    Convenience function to perform calculations without instantiating an engine.

    Args:
        calculations: List of calculation specs

    Returns:
        Dict with "success" flag and "results" dict

    Example:
        result = calculate([
            {"name": "total", "operation": "sum", "args": [100, 200, 300]},
            {"name": "avg", "operation": "average", "args": [100, 200, 300]}
        ])
        # Returns: {"success": True, "results": {"total": 600, "avg": 200}}
    """
    engine = CalculationEngine()
    return engine.execute(calculations)
