package alert

import (
	"github.com/expr-lang/expr"
)

func Evaluate(expression string, data map[string]interface{}) bool {
	// Parse the expression
	program, err := expr.Compile(expression, expr.Env(data))
	if err != nil {
		return false
	}

	// Evaluate the expression
	result, err := expr.Run(program, data)
	if err != nil {
		return false
	}

	// Convert the result to a boolean
	boolResult, ok := result.(bool)
	if !ok {
		return false
	}

	return boolResult
}
