// Package calculator provides precise arithmetic calculations using decimal arithmetic.
// It eliminates floating-point errors common in IEEE 754 arithmetic.
package calculator

import (
	"fmt"
	"math"
	"sort"

	"github.com/shopspring/decimal"
)

// Calculation represents a single calculation to perform.
type Calculation struct {
	Name      string `json:"name,omitempty"`
	Operation string `json:"operation"`
	Args      []any  `json:"args"`
}

// Result represents the result of a batch calculation.
type Result struct {
	Success bool           `json:"success"`
	Results map[string]any `json:"results"`
}

// Engine performs precise arithmetic calculations.
type Engine struct {
	decimalPlaces int32
	results       map[string]any
}

// NewEngine creates a new calculation engine with default precision (10 decimal places).
func NewEngine() *Engine {
	return &Engine{
		decimalPlaces: 10,
		results:       make(map[string]any),
	}
}

// NewEngineWithPrecision creates a new calculation engine with custom precision.
func NewEngineWithPrecision(places int32) *Engine {
	return &Engine{
		decimalPlaces: places,
		results:       make(map[string]any),
	}
}

// Execute performs a batch of calculations.
func (e *Engine) Execute(calculations []Calculation) Result {
	e.results = make(map[string]any)

	for i, calc := range calculations {
		name := calc.Name
		if name == "" {
			name = fmt.Sprintf("result_%d", i)
		}

		result, err := e.executeOne(calc.Operation, calc.Args)
		if err != nil {
			e.results[name] = map[string]any{"error": err.Error()}
		} else {
			e.results[name] = result
		}
	}

	return Result{
		Success: true,
		Results: e.results,
	}
}

// resolveArg resolves a single argument, which may be a number or a reference.
// The visited map tracks already-resolved variable names to detect circular references.
func (e *Engine) resolveArg(arg any) (decimal.Decimal, error) {
	return e.resolveArgWithVisited(arg, make(map[string]bool))
}

// resolveArgWithVisited resolves an argument with cycle detection.
func (e *Engine) resolveArgWithVisited(arg any, visited map[string]bool) (decimal.Decimal, error) {
	switch v := arg.(type) {
	case string:
		// Check for circular reference
		if visited[v] {
			return decimal.Zero, fmt.Errorf("circular reference detected: '%s'", v)
		}

		// Variable reference
		if result, ok := e.results[v]; ok {
			if errMap, isErr := result.(map[string]any); isErr {
				if _, hasError := errMap["error"]; hasError {
					return decimal.Zero, fmt.Errorf("cannot use errored result '%s'", v)
				}
			}
			// Mark as visited before recursing
			visited[v] = true
			return e.resolveArgWithVisited(result, visited)
		}
		// Try to parse as number string
		return decimal.NewFromString(v)
	case int:
		return decimal.NewFromInt(int64(v)), nil
	case int32:
		return decimal.NewFromInt(int64(v)), nil
	case int64:
		return decimal.NewFromInt(v), nil
	case float32:
		return decimal.NewFromFloat32(v), nil
	case float64:
		return decimal.NewFromFloat(v), nil
	case decimal.Decimal:
		return v, nil
	default:
		return decimal.Zero, fmt.Errorf("unsupported argument type: %T", arg)
	}
}

// resolveArgs resolves all arguments in a list.
func (e *Engine) resolveArgs(args []any) ([]decimal.Decimal, error) {
	resolved := make([]decimal.Decimal, len(args))
	for i, arg := range args {
		val, err := e.resolveArg(arg)
		if err != nil {
			return nil, err
		}
		resolved[i] = val
	}
	return resolved, nil
}

// executeOne executes a single operation.
func (e *Engine) executeOne(operation string, args []any) (any, error) {
	// Special handling for compare - only resolve first 2 args as decimals
	if operation == "compare" {
		return e.opCompare(args)
	}

	resolved, err := e.resolveArgs(args)
	if err != nil {
		return nil, err
	}

	switch operation {
	case "add":
		return e.opAdd(resolved)
	case "subtract":
		return e.opSubtract(resolved)
	case "multiply":
		return e.opMultiply(resolved)
	case "divide":
		return e.opDivide(resolved)
	case "sum":
		return e.opSum(resolved)
	case "average":
		return e.opAverage(resolved)
	case "min":
		return e.opMin(resolved)
	case "max":
		return e.opMax(resolved)
	case "median":
		return e.opMedian(resolved)
	case "stddev":
		return e.opStddev(resolved)
	case "percentage":
		return e.opPercentage(resolved)
	case "roi":
		return e.opROI(resolved)
	case "compound_interest":
		return e.opCompoundInterest(resolved)
	case "present_value":
		return e.opPresentValue(resolved)
	case "round":
		return e.opRound(resolved, args)
	case "abs":
		return e.opAbs(resolved)
	case "ceil":
		return e.opCeil(resolved)
	case "floor":
		return e.opFloor(resolved)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

// decimalToNumber converts a Decimal to a numeric value.
// For integers, returns int64 to avoid floating point representation issues.
// For decimals, returns the float64 representation.
func decimalToNumber(d decimal.Decimal) any {
	// If it's an integer, return as int64 for cleaner JSON output
	if d.Equal(d.Truncate(0)) {
		return d.IntPart()
	}
	// Otherwise return as float64
	f, _ := d.Float64()
	return f
}

func (e *Engine) opAdd(args []decimal.Decimal) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("add requires at least 2 arguments")
	}
	result := args[0]
	for _, arg := range args[1:] {
		result = result.Add(arg)
	}
	return decimalToNumber(result), nil
}

func (e *Engine) opSubtract(args []decimal.Decimal) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("subtract requires at least 2 arguments")
	}
	// Variadic: subtract every arg after the first from the running result.
	// Matches add/multiply's left-fold semantics so subtract(a, b, c) is a-b-c.
	result := args[0]
	for _, arg := range args[1:] {
		result = result.Sub(arg)
	}
	return decimalToNumber(result), nil
}

func (e *Engine) opMultiply(args []decimal.Decimal) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("multiply requires at least 2 arguments")
	}
	result := args[0]
	for _, arg := range args[1:] {
		result = result.Mul(arg)
	}
	return decimalToNumber(result), nil
}

func (e *Engine) opDivide(args []decimal.Decimal) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("divide requires at least 2 arguments")
	}
	// Variadic: divide by every arg after the first (left-fold). Matches
	// the symmetry with subtract's variadic semantics.
	result := args[0]
	for _, arg := range args[1:] {
		if arg.IsZero() {
			return nil, fmt.Errorf("division by zero")
		}
		result = result.Div(arg)
	}
	return decimalToNumber(result), nil
}

func (e *Engine) opSum(args []decimal.Decimal) (any, error) {
	if len(args) == 0 {
		return int64(0), nil
	}
	result := decimal.Zero
	for _, arg := range args {
		result = result.Add(arg)
	}
	return decimalToNumber(result), nil
}

func (e *Engine) opAverage(args []decimal.Decimal) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("average requires at least 1 argument")
	}
	sum := decimal.Zero
	for _, arg := range args {
		sum = sum.Add(arg)
	}
	result := sum.Div(decimal.NewFromInt(int64(len(args))))
	return decimalToNumber(result), nil
}

func (e *Engine) opMin(args []decimal.Decimal) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("min requires at least 1 argument")
	}
	result := args[0]
	for _, arg := range args[1:] {
		if arg.LessThan(result) {
			result = arg
		}
	}
	return decimalToNumber(result), nil
}

func (e *Engine) opMax(args []decimal.Decimal) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("max requires at least 1 argument")
	}
	result := args[0]
	for _, arg := range args[1:] {
		if arg.GreaterThan(result) {
			result = arg
		}
	}
	return decimalToNumber(result), nil
}

func (e *Engine) opMedian(args []decimal.Decimal) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("median requires at least 1 argument")
	}

	// Sort a copy
	sorted := make([]decimal.Decimal, len(args))
	copy(sorted, args)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LessThan(sorted[j])
	})

	n := len(sorted)
	mid := n / 2

	var result decimal.Decimal
	if n%2 == 0 {
		result = sorted[mid-1].Add(sorted[mid]).Div(decimal.NewFromInt(2))
	} else {
		result = sorted[mid]
	}

	return decimalToNumber(result), nil
}

func (e *Engine) opStddev(args []decimal.Decimal) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("stddev requires at least 2 arguments")
	}

	// Calculate mean
	sum := decimal.Zero
	for _, arg := range args {
		sum = sum.Add(arg)
	}
	mean := sum.Div(decimal.NewFromInt(int64(len(args))))

	// Calculate variance
	variance := decimal.Zero
	for _, arg := range args {
		diff := arg.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(args))))

	// Calculate stddev (sqrt of variance)
	varianceFloat, _ := variance.Float64()
	return math.Sqrt(varianceFloat), nil
}

func (e *Engine) opPercentage(args []decimal.Decimal) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("percentage requires 2 arguments: new, old")
	}
	newVal, oldVal := args[0], args[1]
	if oldVal.IsZero() {
		return nil, fmt.Errorf("cannot calculate percentage with zero base")
	}
	// ((new - old) / old) * 100
	result := newVal.Sub(oldVal).Div(oldVal).Mul(decimal.NewFromInt(100))
	return decimalToNumber(result), nil
}

func (e *Engine) opROI(args []decimal.Decimal) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("roi requires 2 arguments: gain, cost")
	}
	gain, cost := args[0], args[1]
	if cost.IsZero() {
		return nil, fmt.Errorf("cannot calculate ROI with zero cost")
	}
	// ((gain - cost) / cost) * 100
	result := gain.Sub(cost).Div(cost).Mul(decimal.NewFromInt(100))
	return decimalToNumber(result), nil
}

func (e *Engine) opCompoundInterest(args []decimal.Decimal) (any, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("compound_interest requires 3 arguments: principal, rate, periods")
	}
	principal, rate := args[0], args[1]
	periods, _ := args[2].Float64()

	// principal * (1 + rate)^periods
	onePlusRate := decimal.NewFromInt(1).Add(rate)
	onePlusRateFloat, _ := onePlusRate.Float64()
	multiplier := math.Pow(onePlusRateFloat, periods)
	principalFloat, _ := principal.Float64()

	return principalFloat * multiplier, nil
}

func (e *Engine) opPresentValue(args []decimal.Decimal) (any, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("present_value requires 3 arguments: future_value, rate, periods")
	}
	fv, rate := args[0], args[1]
	periods, _ := args[2].Float64()

	// fv / (1 + rate)^periods
	onePlusRate := decimal.NewFromInt(1).Add(rate)
	onePlusRateFloat, _ := onePlusRate.Float64()
	divisor := math.Pow(onePlusRateFloat, periods)
	fvFloat, _ := fv.Float64()

	return fvFloat / divisor, nil
}

func (e *Engine) opRound(args []decimal.Decimal, rawArgs []any) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("round requires at least 1 argument")
	}

	places := e.decimalPlaces
	if len(rawArgs) > 1 {
		if p, err := e.resolveArg(rawArgs[1]); err == nil {
			places = int32(p.IntPart())
		}
	}

	result := args[0].Round(places)
	return decimalToNumber(result), nil
}

func (e *Engine) opAbs(args []decimal.Decimal) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("abs requires 1 argument")
	}
	result := args[0].Abs()
	return decimalToNumber(result), nil
}

func (e *Engine) opCeil(args []decimal.Decimal) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("ceil requires 1 argument")
	}
	result := args[0].Ceil()
	return decimalToNumber(result), nil
}

func (e *Engine) opFloor(args []decimal.Decimal) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("floor requires 1 argument")
	}
	result := args[0].Truncate(0)
	return decimalToNumber(result), nil
}

func (e *Engine) opCompare(rawArgs []any) (bool, error) {
	if len(rawArgs) < 2 {
		return false, fmt.Errorf("compare requires at least 2 arguments")
	}

	// Resolve only the first two arguments as decimals
	a, err := e.resolveArg(rawArgs[0])
	if err != nil {
		return false, err
	}
	b, err := e.resolveArg(rawArgs[1])
	if err != nil {
		return false, err
	}

	op := "<"
	if len(rawArgs) > 2 {
		if opStr, ok := rawArgs[2].(string); ok {
			op = opStr
		}
	}

	switch op {
	case "<":
		return a.LessThan(b), nil
	case ">":
		return a.GreaterThan(b), nil
	case "<=":
		return a.LessThanOrEqual(b), nil
	case ">=":
		return a.GreaterThanOrEqual(b), nil
	case "==", "=":
		return a.Equal(b), nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %s", op)
	}
}

// Calculate is a convenience function to perform calculations without instantiating an engine.
func Calculate(calculations []Calculation) Result {
	engine := NewEngine()
	return engine.Execute(calculations)
}
