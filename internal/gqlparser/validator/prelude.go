package validator

import (
	_ "embed"

	"github.com/open-policy-agent/opa/internal/gqlparser/ast"
)

//go:embed imported/prelude.graphql
var preludeGraphql string

var Prelude = &ast.Source{
	Name:    "prelude.graphql",
	Input:   preludeGraphql,
	BuiltIn: true,
}
