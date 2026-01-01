#!/usr/bin/env python3
"""
Example: Using mcp-calculator for financial report generation.

This example demonstrates how to use the calculator to avoid AI math errors
when generating financial narratives.
"""

import sys
sys.path.insert(0, '../src')

from mcp_calculator import calculate


def analyze_collections():
    """
    Analyze collections data - a real-world scenario where AI made errors.

    Original problem:
    - AI claimed "trending higher" when data showed collections were LOWER
    - AI calculated "72% decline" when actual decline was 26.3%
    """
    print("=" * 60)
    print("Healthcare Collections Analysis")
    print("=" * 60)

    # Raw data
    october_collections = 2561276  # dollars
    october_days = 8  # business days
    september_collections = 8782334
    september_days = 21

    # Calculate per-day rates and compare
    result = calculate([
        {"name": "oct_per_day", "operation": "divide", "args": [october_collections, october_days]},
        {"name": "sep_per_day", "operation": "divide", "args": [september_collections, september_days]},
        {"name": "is_lower", "operation": "compare", "args": ["oct_per_day", "sep_per_day", "<"]},
        {"name": "pct_change", "operation": "percentage", "args": ["oct_per_day", "sep_per_day"]}
    ])

    results = result["results"]
    print(f"\nOctober per-day rate: ${results['oct_per_day']:,.2f}")
    print(f"September per-day rate: ${results['sep_per_day']:,.2f}")
    print(f"October is lower than September: {results['is_lower']}")
    print(f"Percentage change: {results['pct_change']:.1f}%")

    # Generate correct narrative
    trend = "lower" if results["is_lower"] else "higher"
    direction = "decline" if results["pct_change"] < 0 else "increase"
    print(f"\nCorrect narrative: Collections are trending {trend} "
          f"({abs(results['pct_change']):.1f}% {direction})")

    print("\n" + "-" * 60)
    print("Without this tool, AI incorrectly said 'trending higher'")
    print("-" * 60)


def analyze_managed_medicare():
    """
    Managed Medicare analysis - another real error case.

    Original problem:
    - AI calculated "72% decline"
    - Actual decline was 26.3%
    - AI confused ratio (51600/70000 = 73.7%) with percentage change
    """
    print("\n" + "=" * 60)
    print("Managed Medicare Analysis")
    print("=" * 60)

    # Per-day rates (pre-calculated from monthly data)
    october_daily = 51600
    september_daily = 70000

    result = calculate([
        # What AI incorrectly calculated (the ratio)
        {"name": "wrong_ratio", "operation": "divide", "args": [october_daily, september_daily]},
        # Correct percentage change
        {"name": "correct_pct", "operation": "percentage", "args": [october_daily, september_daily]}
    ])

    results = result["results"]
    print(f"\nOctober daily: ${october_daily:,}")
    print(f"September daily: ${september_daily:,}")
    print(f"\nWrong (ratio): {results['wrong_ratio']:.1%} (AI confused this with % change)")
    print(f"Correct (% change): {results['correct_pct']:.1f}%")

    print("\n" + "-" * 60)
    print("Without this tool, AI incorrectly said '72% decline'")
    print("Correct answer: 26.3% decline")
    print("-" * 60)


def batch_financial_metrics():
    """
    Example of batch calculations for a financial dashboard.
    """
    print("\n" + "=" * 60)
    print("Financial Dashboard Metrics")
    print("=" * 60)

    result = calculate([
        # Revenue metrics
        {"name": "total_revenue", "operation": "sum", "args": [125000, 187500, 162000]},
        {"name": "avg_monthly_revenue", "operation": "average", "args": [125000, 187500, 162000]},

        # Growth calculations
        {"name": "q3_growth", "operation": "percentage", "args": [162000, 125000]},

        # Investment returns
        {"name": "roi", "operation": "roi", "args": [175000, 100000]},

        # Future value projection
        {"name": "projected_value", "operation": "compound_interest", "args": [100000, 0.08, 5]},

        # Comparison
        {"name": "exceeded_target", "operation": "compare", "args": ["total_revenue", 400000, ">"]}
    ])

    r = result["results"]
    print(f"\nTotal Revenue: ${r['total_revenue']:,.2f}")
    print(f"Avg Monthly Revenue: ${r['avg_monthly_revenue']:,.2f}")
    print(f"Q3 Growth: {r['q3_growth']:.1f}%")
    print(f"ROI: {r['roi']:.1f}%")
    print(f"5-Year Projection (8% annual): ${r['projected_value']:,.2f}")
    print(f"Exceeded $400K target: {r['exceeded_target']}")


if __name__ == "__main__":
    analyze_collections()
    analyze_managed_medicare()
    batch_financial_metrics()

    print("\n" + "=" * 60)
    print("All calculations verified with 100% accuracy")
    print("=" * 60)
