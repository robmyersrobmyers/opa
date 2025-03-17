// Copyright 2025 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"context"
	_ "embed"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown/cache"
)

// employeeGQLSchema is a simple schema defined in graphql_test.go
//
// If you don't have your own complex GQL schema you can use the
// [GitHub GraphQL Schema](https://docs.github.com/en/graphql) for more realistic performance testing
//
// Download the schema with curl -o complex-schema.gql -L https://docs.github.com/public/fpt/schema.docs.graphql
// and use embed so we don't include an extra 2+M in source tree.
//
//go:embed complex-schema.gql
var complexGQLSchema string

// graphqurl --introspect https://api.github.com/graphql -H "Authorization: bearer $GITHUB_TOKEN" --format=json > complex-schema.json
//
//go:embed complex-schema.json
var complexGQLSchemaJSON string

func BenchmarkGraphQLSchemaIsValid(b *testing.B) {

	// Share an InterQueryValueCache across multiple runs
	// Tune number of entries to exceed number of distinct GQL schemas
	in := `{"inter_query_builtin_value_cache": {"max_num_entries": 10},}`
	config, _ := cache.ParseCachingConfig([]byte(in))
	valueCache := cache.NewInterQueryValueCache(context.Background(), config)

	benches := []struct {
		desc   string
		schema *ast.Term
		cache  cache.InterQueryValueCache
	}{
		{"Trivial Schema - string", ast.NewTerm(ast.String(employeeGQLSchema)), nil},
		{"Trivial Schema with cache - string", ast.NewTerm(ast.String(employeeGQLSchema)), valueCache},
		{"Trivial Schema - object", ast.NewTerm(ast.MustParseTerm(employeeGQLSchemaAST).Value.(ast.Object)), nil},
		{"Trivial Schema with cache - object", ast.NewTerm(ast.MustParseTerm(employeeGQLSchemaAST).Value.(ast.Object)), valueCache},
		{"Trivial Schema - AST JSON string", ast.NewTerm(ast.String(employeeGQLSchemaAST)), nil},
		{"Trivial Schema with cache - AST JSON string", ast.NewTerm(ast.String(employeeGQLSchemaAST)), valueCache},
		{"Complex Schema - string", ast.NewTerm(ast.String(complexGQLSchema)), nil},
		{"Complex Schema with cache - string", ast.NewTerm(ast.String(complexGQLSchema)), valueCache},
		{"Complex Schema - JSON string", ast.NewTerm(ast.String(complexGQLSchemaJSON)), nil},
		{"Complex Schema with cache - JSON string", ast.NewTerm(ast.String(complexGQLSchemaJSON)), valueCache},
	}

	for _, bench := range benches {
		b.Run(bench.desc, func(b *testing.B) {
			for range b.N {
				err := builtinGraphQLSchemaIsValid(
					BuiltinContext{
						InterQueryBuiltinValueCache: bench.cache,
					},
					[]*ast.Term{bench.schema},
					func(term *ast.Term) error {
						return nil
					},
				)

				if err != nil {
					b.Fatalf("unexpected error: %s", err)
				}
			}
		})
	}
}
