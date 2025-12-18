#!/usr/bin/env python3
"""
MCP Server for precise arithmetic calculations.

This module implements the Model Context Protocol (MCP) server that exposes
the calculator as a tool for AI agents. Supports STDIO transport.

Usage:
    python3 -m mcp_calculator          # Run as MCP server (JSON-RPC over STDIO)
    python3 -m mcp_calculator --test   # Run self-test

Protocol: JSON-RPC 2.0 over STDIO (newline-delimited)
"""

import json
import sys
import logging
from typing import Any, Dict, Optional

from .calculator import calculate

# Configure logging to stderr (stdout is for JSON-RPC)
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    stream=sys.stderr
)
logger = logging.getLogger(__name__)

# Server metadata
SERVER_NAME = "mcp-calculator"
SERVER_VERSION = "0.1.0"
PROTOCOL_VERSION = "2024-11-05"

# Tool definition
CALCULATE_TOOL = {
    "name": "calculate",
    "description": (
        "Perform precise arithmetic calculations using Decimal precision. "
        "Use this tool for ALL math operations - never calculate in your response. "
        "Supports batch operations: add, subtract, multiply, divide, sum, average, "
        "percentage, round, min, max, median, stddev, compare, abs, ceil, floor, "
        "roi, compound_interest, present_value. "
        "Results can reference previous calculations by name."
    ),
    "inputSchema": {
        "type": "object",
        "properties": {
            "calculations": {
                "type": "array",
                "description": "List of calculations to perform",
                "items": {
                    "type": "object",
                    "properties": {
                        "name": {
                            "type": "string",
                            "description": "Label for the result (can be referenced by later calculations)"
                        },
                        "operation": {
                            "type": "string",
                            "enum": [
                                "add", "subtract", "multiply", "divide",
                                "sum", "average", "percentage", "round",
                                "min", "max", "median", "stddev",
                                "compare", "abs", "ceil", "floor",
                                "roi", "compound_interest", "present_value"
                            ],
                            "description": "Operation to perform"
                        },
                        "args": {
                            "type": "array",
                            "items": {
                                "oneOf": [
                                    {"type": "number"},
                                    {"type": "string", "description": "Reference to previous result by name"}
                                ]
                            },
                            "description": "Numeric arguments or references to previous results"
                        }
                    },
                    "required": ["operation", "args"]
                }
            }
        },
        "required": ["calculations"]
    }
}


class MCPCalculatorServer:
    """
    MCP Server that exposes the calculate tool.

    Handles JSON-RPC 2.0 protocol over STDIO for integration with
    MCP clients like Claude Desktop and Claude Code.
    """

    def handle_request(self, request: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """
        Handle incoming JSON-RPC request.

        Supports:
        - initialize: Server handshake
        - notifications/initialized: Client acknowledgment
        - tools/list: Return available tools
        - tools/call: Execute calculate tool
        """
        method = request.get("method", "")
        params = request.get("params", {})
        request_id = request.get("id")

        logger.debug(f"Handling method: {method}")

        try:
            if method == "initialize":
                result = {
                    "protocolVersion": PROTOCOL_VERSION,
                    "capabilities": {
                        "tools": {}
                    },
                    "serverInfo": {
                        "name": SERVER_NAME,
                        "version": SERVER_VERSION
                    }
                }

            elif method == "notifications/initialized":
                # Client acknowledgment - no response needed
                return None

            elif method == "tools/list":
                result = {
                    "tools": [CALCULATE_TOOL]
                }

            elif method == "tools/call":
                tool_name = params.get("name")
                tool_args = params.get("arguments", {})

                if tool_name == "calculate":
                    calculations = tool_args.get("calculations", [])
                    tool_result = calculate(calculations)
                else:
                    raise ValueError(f"Unknown tool: {tool_name}")

                # Wrap result in MCP CallToolResult format
                result = {
                    "content": [
                        {
                            "type": "text",
                            "text": json.dumps(tool_result, indent=2)
                        }
                    ]
                }

            else:
                raise ValueError(f"Unknown method: {method}")

            # Build successful response
            if request_id is not None:
                return {
                    "jsonrpc": "2.0",
                    "id": request_id,
                    "result": result
                }
            return None  # Notification, no response

        except Exception as e:
            logger.error(f"Error handling {method}: {e}")
            if request_id is not None:
                return {
                    "jsonrpc": "2.0",
                    "id": request_id,
                    "error": {
                        "code": -32603,
                        "message": str(e)
                    }
                }
            return None

    def run_stdio(self):
        """Run the server using STDIO transport."""
        logger.info(f"Starting {SERVER_NAME} v{SERVER_VERSION}")

        for line in sys.stdin:
            line = line.strip()
            if not line:
                continue

            try:
                request = json.loads(line)
                logger.debug(f"Received: {request.get('method', 'unknown')}")

                response = self.handle_request(request)

                if response is not None:
                    print(json.dumps(response), flush=True)

            except json.JSONDecodeError as e:
                logger.error(f"JSON decode error: {e}")
                error_response = {
                    "jsonrpc": "2.0",
                    "id": None,
                    "error": {
                        "code": -32700,
                        "message": f"Parse error: {e}"
                    }
                }
                print(json.dumps(error_response), flush=True)

            except Exception as e:
                logger.error(f"Unexpected error: {e}")


def run_self_test() -> bool:
    """Run self-test to verify calculate function works correctly."""
    print("Running mcp-calculator self-test...", file=sys.stderr)

    test_cases = [
        # Basic operations
        {
            "name": "sum",
            "input": [{"name": "sum", "operation": "sum", "args": [1, 2, 3]}],
            "expected": {"sum": 6.0}
        },
        {
            "name": "average",
            "input": [{"name": "avg", "operation": "average", "args": [10, 20, 30]}],
            "expected": {"avg": 20.0}
        },
        {
            "name": "percentage",
            "input": [{"name": "pct", "operation": "percentage", "args": [150, 100]}],
            "expected": {"pct": 50.0}
        },
        {
            "name": "divide",
            "input": [{"name": "div", "operation": "divide", "args": [100, 4]}],
            "expected": {"div": 25.0}
        },
        # Variable references
        {
            "name": "variable_refs",
            "input": [
                {"name": "a", "operation": "sum", "args": [10, 20]},
                {"name": "b", "operation": "multiply", "args": ["a", 2]}
            ],
            "expected": {"a": 30.0, "b": 60.0}
        },
        # Precision test
        {
            "name": "precision",
            "input": [{"name": "precise", "operation": "divide", "args": [1, 3]}],
            "check": lambda r: abs(r["results"]["precise"] - 0.333333) < 0.001
        },
        # Error handling
        {
            "name": "div_by_zero",
            "input": [{"name": "err", "operation": "divide", "args": [100, 0]}],
            "check": lambda r: "error" in r["results"]["err"]
        },
    ]

    all_passed = True
    for tc in test_cases:
        result = calculate(tc["input"])

        if "expected" in tc:
            if result["results"] != tc["expected"]:
                print(f"  FAIL {tc['name']}: expected {tc['expected']}, got {result['results']}", file=sys.stderr)
                all_passed = False
            else:
                print(f"  PASS {tc['name']}", file=sys.stderr)
        elif "check" in tc:
            if tc["check"](result):
                print(f"  PASS {tc['name']}", file=sys.stderr)
            else:
                print(f"  FAIL {tc['name']}: check failed, got {result}", file=sys.stderr)
                all_passed = False

    if all_passed:
        print("All tests passed!", file=sys.stderr)
    return all_passed


def main():
    """Main entry point."""
    # Check for --test flag
    if len(sys.argv) > 1 and sys.argv[1] == "--test":
        success = run_self_test()
        sys.exit(0 if success else 1)

    # Run STDIO server
    server = MCPCalculatorServer()
    server.run_stdio()


if __name__ == "__main__":
    main()
