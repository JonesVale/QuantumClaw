package service

import (
	"github.com/expr-lang/expr"
)

// init registers the EvaluateExpr function variable used by tiered_settle.go.
// This deferred import pattern avoids circular dependencies between billing packages.
func init() {
	EvaluateExpr = evaluateBillingExprImpl
}

func evaluateBillingExprImpl(exprStr string, env map[string]interface{}) (interface{}, error) {
	if exprStr == "" {
		return 0, nil
	}
	program, err := expr.Compile(exprStr, expr.Env(env))
	if err != nil {
		return nil, err
	}
	output, err := expr.Run(program, env)
	if err != nil {
		return nil, err
	}
	return output, nil
}
