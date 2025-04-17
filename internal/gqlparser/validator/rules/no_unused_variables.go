package rules

import (
	"github.com/open-policy-agent/opa/internal/gqlparser/ast"

	//nolint:staticcheck // Validator rules each use dot imports for convenience.
	. "github.com/open-policy-agent/opa/internal/gqlparser/validator"
)

var NoUnusedVariablesRule = Rule{
	Name: "NoUnusedVariables",
	RuleFunc: func(observers *Events, addError AddErrFunc) {
		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			for _, varDef := range operation.VariableDefinitions {
				if varDef.Used {
					continue
				}

				if operation.Name != "" {
					addError(
						Message(`Variable "$%s" is never used in operation "%s".`, varDef.Variable, operation.Name),
						At(varDef.Position),
					)
				} else {
					addError(
						Message(`Variable "$%s" is never used.`, varDef.Variable),
						At(varDef.Position),
					)
				}
			}
		})
	},
}

func init() {
	AddRule(NoUnusedVariablesRule.Name, NoUnusedVariablesRule.RuleFunc)
}
