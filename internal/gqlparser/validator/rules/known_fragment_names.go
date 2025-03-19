package rules

import (
	"github.com/open-policy-agent/opa/internal/gqlparser/ast"

	//nolint:revive // Validator rules each use dot imports for convenience.
	. "github.com/open-policy-agent/opa/internal/gqlparser/validator"
)

var KnownFragmentNamesRule = Rule{
	Name: "KnownFragmentNames",
	RuleFunc: func(observers *Events, addError AddErrFunc) {
		observers.OnFragmentSpread(func(walker *Walker, fragmentSpread *ast.FragmentSpread) {
			if fragmentSpread.Definition == nil {
				addError(
					Message(`Unknown fragment "%s".`, fragmentSpread.Name),
					At(fragmentSpread.Position),
				)
			}
		})
	},
}

func init() {
	AddRule(KnownFragmentNamesRule.Name, KnownFragmentNamesRule.RuleFunc)
}
