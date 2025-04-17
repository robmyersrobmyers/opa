package rules

import (
	"github.com/open-policy-agent/opa/internal/gqlparser/ast"

	//nolint:staticcheck // Validator rules each use dot imports for convenience.
	. "github.com/open-policy-agent/opa/internal/gqlparser/validator"
)

var UniqueFragmentNamesRule = Rule{
	Name: "UniqueFragmentNames",
	RuleFunc: func(observers *Events, addError AddErrFunc) {
		seenFragments := map[string]bool{}

		observers.OnFragment(func(walker *Walker, fragment *ast.FragmentDefinition) {
			if seenFragments[fragment.Name] {
				addError(
					Message(`There can be only one fragment named "%s".`, fragment.Name),
					At(fragment.Position),
				)
			}
			seenFragments[fragment.Name] = true
		})
	},
}

func init() {
	AddRule(UniqueFragmentNamesRule.Name, UniqueFragmentNamesRule.RuleFunc)
}
