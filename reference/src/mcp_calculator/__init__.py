"""
mcp-calculator: Precise Arithmetic for AI Agents

A Model Context Protocol (MCP) server that eliminates AI math errors
by providing deterministic arithmetic operations using Python's Decimal module.
"""

__version__ = "0.1.0"
__author__ = "MCP Calculator Contributors"
__license__ = "MIT"

from .calculator import calculate, CalculationEngine
from .server import MCPCalculatorServer

__all__ = ["calculate", "CalculationEngine", "MCPCalculatorServer", "__version__"]
