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

// Convert the complex schema string into an Object
var complexGQLSchemaObj *ast.Term = ast.MustParseTerm(complexGQLSchemaJSON).Value.(ast.Object).Get(ast.NewTerm(ast.String("__schema")))

const complexGQLQueryASTString = `{"Operations":[{"Name":"","Operation":"query","SelectionSet":[{"Alias":"securityAdvisories","Name":"securityAdvisories","SelectionSet":[{"Alias":"totalCount","Name":"totalCount"}]}]}]}`

var complexGQLQueryASTObj *ast.Term = ast.NewTerm(ast.String(complexGQLQueryASTString))

//go:embed employee-schema.json
var employeeGQLSchemaJSON string
var employeeGQLSchemaObj *ast.Term = ast.MustParseTerm(employeeGQLSchemaJSON).Value.(ast.Object).Get(ast.NewTerm(ast.String("__schema")))

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
		result *ast.Term
	}{
		{"Trivial Schema - string", ast.NewTerm(ast.String(employeeGQLSchema)), nil, ast.BooleanTerm(true)},
		{"Trivial Schema with cache - string", ast.NewTerm(ast.String(employeeGQLSchema)), valueCache, ast.BooleanTerm(true)},
		{"Trivial Schema - object", employeeGQLSchemaObj, nil, ast.BooleanTerm(true)},
		{"Trivial Schema with cache - object", employeeGQLSchemaObj, valueCache, ast.BooleanTerm(true)},
		{"Trivial Schema - AST object", ast.NewTerm(ast.MustParseTerm(employeeGQLSchemaAST).Value.(ast.Object)), nil, ast.BooleanTerm(true)},
		{"Trivial Schema with cache - AST object", ast.NewTerm(ast.MustParseTerm(employeeGQLSchemaAST).Value.(ast.Object)), valueCache, ast.BooleanTerm(true)},
		{"Complex Schema - string", ast.NewTerm(ast.String(complexGQLSchema)), nil, ast.BooleanTerm(true)},
		{"Complex Schema with cache - string", ast.NewTerm(ast.String(complexGQLSchema)), valueCache, ast.BooleanTerm(true)},
		{"Complex Schema - object", complexGQLSchemaObj, nil, ast.BooleanTerm(true)},
		{"Complex Schema with cache - object", complexGQLSchemaObj, valueCache, ast.BooleanTerm(true)},
	}

	for _, bench := range benches {
		b.Run(bench.desc, func(b *testing.B) {
			for range b.N {
				var result *ast.Term
				b.StartTimer()
				err := builtinGraphQLSchemaIsValid(
					BuiltinContext{
						InterQueryBuiltinValueCache: bench.cache,
					},
					[]*ast.Term{bench.schema},
					func(term *ast.Term) error {
						result = term
						return nil
					},
				)
				b.StopTimer()
				if err != nil {
					b.Fatalf("unexpected error: %s", err)
				}
				if !bench.result.Equal(result) {
					b.Fatalf("unexpected result: wanted: %#v got: %#v", bench.result, result)
				}
			}
		})
	}
}

func BenchmarkGraphQLParseSchema(b *testing.B) {

	// Share an InterQueryValueCache across multiple runs
	// Tune number of entries to exceed number of distinct GQL schemas
	in := `{"inter_query_builtin_value_cache": {"max_num_entries": 10},}`
	config, _ := cache.ParseCachingConfig([]byte(in))
	valueCache := cache.NewInterQueryValueCache(context.Background(), config)

	benches := []struct {
		desc   string
		schema *ast.Term
		cache  cache.InterQueryValueCache
		result *ast.Term
	}{
		{"Trivial Schema - string", ast.NewTerm(ast.String(employeeGQLSchema)), nil, ast.NewTerm(employeeGQLSchemaASTObj)},
		{"Trivial Schema with cache - string", ast.NewTerm(ast.String(employeeGQLSchema)), valueCache, ast.NewTerm(employeeGQLSchemaASTObj)},
		// TODO: InterfaceToValue allocates a ton of memory and never comes back in this case
		// {"Complex Schema - string", ast.NewTerm(ast.String(complexGQLSchema)), nil, complexGQLSchemaObj},
		// {"Complex Schema with cache - string", ast.NewTerm(ast.String(complexGQLSchema)), valueCache, complexGQLSchemaObj},
	}
	for _, bench := range benches {
		b.Run(bench.desc, func(b *testing.B) {
			for range b.N {
				var result *ast.Term
				b.StartTimer()
				err := builtinGraphQLParseSchema(
					BuiltinContext{
						InterQueryBuiltinValueCache: bench.cache,
					},
					[]*ast.Term{bench.schema},
					func(term *ast.Term) error {
						result = term
						return nil
					},
				)
				b.StopTimer()
				if err != nil {
					b.Fatalf("unexpected error: %s", err)
				}
				if !bench.result.Equal(result) {
					b.Errorf("Unexpected result, expected %#v, got %#v", bench.result, result)
					return
				}
			}
		})
	}
}

func BenchmarkGraphQLParseQuery(b *testing.B) {

	// Share an InterQueryValueCache across multiple runs
	// Tune number of entries to exceed number of distinct GQL schemas
	in := `{"inter_query_builtin_value_cache": {"max_num_entries": 10},}`
	config, _ := cache.ParseCachingConfig([]byte(in))
	valueCache := cache.NewInterQueryValueCache(context.Background(), config)

	benches := []struct {
		desc   string
		query  *ast.Term
		cache  cache.InterQueryValueCache
		result *ast.Term
	}{
		{"Trivial Query - string", ast.NewTerm(ast.String(`{ employeeByID(id: "alice") { salary } }`)), nil, ast.NewTerm(employeeGQLQueryASTObj)},
		{"Trivial Query with cache - string", ast.NewTerm(ast.String(`{ employeeByID(id: "alice") { salary } }`)), valueCache, ast.NewTerm(employeeGQLQueryASTObj)},
	}
	for _, bench := range benches {
		b.Run(bench.desc, func(b *testing.B) {
			for range b.N {
				var result *ast.Term
				b.StartTimer()
				err := builtinGraphQLParseQuery(
					BuiltinContext{
						InterQueryBuiltinValueCache: bench.cache,
					},
					[]*ast.Term{bench.query},
					func(term *ast.Term) error {
						result = term
						return nil
					},
				)
				b.StopTimer()
				if err != nil {
					b.Fatalf("unexpected error: %s", err)
				}
				if !bench.result.Equal(result) {
					b.Errorf("Unexpected result, expected %#v, got %#v", bench.result, result)
					return
				}
			}
		})
	}
}

func BenchmarkGraphQLIsValid(b *testing.B) {

	// Share an InterQueryValueCache across multiple runs
	// Tune number of entries to exceed number of distinct GQL schemas
	in := `{"inter_query_builtin_value_cache": {"max_num_entries": 10},}`
	config, _ := cache.ParseCachingConfig([]byte(in))
	valueCache := cache.NewInterQueryValueCache(context.Background(), config)

	benches := []struct {
		desc   string
		schema *ast.Term
		cache  cache.InterQueryValueCache
		query  *ast.Term
		result *ast.Term
	}{
		{
			desc:   "Trivial Schema - string",
			cache:  nil,
			query:  ast.NewTerm(ast.String(`{ employeeByID(id: "alice") { salary } }`)),
			schema: ast.NewTerm(ast.String(employeeGQLSchema)),
			result: ast.BooleanTerm(true),
		},
		{
			desc:   "Trivial Schema with cache - string",
			cache:  valueCache,
			query:  ast.NewTerm(ast.String(`{ employeeByID(id: "alice") { salary } }`)),
			schema: ast.NewTerm(ast.String(employeeGQLSchema)),
			result: ast.BooleanTerm(true),
		},
		{
			desc:   "Complex Schema - string",
			cache:  nil,
			query:  ast.NewTerm(ast.String(`{ securityAdvisories { totalCount } }`)),
			schema: ast.NewTerm(ast.String(complexGQLSchema)),
			result: ast.BooleanTerm(true),
		},
		{
			desc:   "Complex Schema with cache - string",
			cache:  valueCache,
			query:  ast.NewTerm(ast.String(`{ securityAdvisories { totalCount } }`)),
			schema: ast.NewTerm(ast.String(complexGQLSchema)),
			result: ast.BooleanTerm(true),
		},
		// TODO add objects
	}
	for _, bench := range benches {
		b.Run(bench.desc, func(b *testing.B) {
			for range b.N {
				var result *ast.Term
				b.StartTimer()
				err := builtinGraphQLIsValid(
					BuiltinContext{
						InterQueryBuiltinValueCache: bench.cache,
					},
					[]*ast.Term{bench.query, bench.schema},
					func(term *ast.Term) error {
						result = term
						return nil
					},
				)
				b.StopTimer()
				if err != nil {
					b.Fatalf("unexpected error: %s", err)
				}
				if !bench.result.Equal(result) {
					b.Errorf("Unexpected result, expected %#v, got %#v", bench.result, result)
					return
				}
			}
		})
	}
}

func BenchmarkGraphQLParse(b *testing.B) {

	// Share an InterQueryValueCache across multiple runs
	// Tune number of entries to exceed number of distinct GQL schemas
	in := `{"inter_query_builtin_value_cache": {"max_num_entries": 10},}`
	config, _ := cache.ParseCachingConfig([]byte(in))
	valueCache := cache.NewInterQueryValueCache(context.Background(), config)

	// Use this to map result item position to purpose for better error messages
	resultItemDescription := []string{"query_ast", "schema_ast"}

	benches := []struct {
		desc   string
		schema *ast.Term
		cache  cache.InterQueryValueCache
		query  *ast.Term
		result *ast.Term
	}{
		{
			desc:   "Trivial Schema - string",
			cache:  nil,
			query:  ast.NewTerm(ast.String(`{ employeeByID(id: "alice") { salary } }`)),
			schema: ast.NewTerm(ast.String(employeeGQLSchema)),
			result: ast.ArrayTerm(
				ast.NewTerm(employeeGQLQueryASTObj),
				ast.NewTerm(employeeGQLSchemaASTObj),
			),
		},
		{
			desc:   "Trivial Schema with cache - string",
			cache:  valueCache,
			query:  ast.NewTerm(ast.String(`{ employeeByID(id: "alice") { salary } }`)),
			schema: ast.NewTerm(ast.String(employeeGQLSchema)),
			result: ast.ArrayTerm(
				ast.NewTerm(employeeGQLQueryASTObj),
				ast.NewTerm(employeeGQLSchemaASTObj),
			),
		},
		// {
		// 	desc:   "Complex Schema - string",
		// 	cache:  nil,
		// 	query:  ast.NewTerm(ast.String(`{ securityAdvisories { totalCount } }`)),
		// 	schema: ast.NewTerm(ast.String(complexGQLSchema)),
		// 	result: ast.ArrayTerm(
		// 		complexGQLQueryASTObj,
		// 		complexGQLSchemaObj,
		// 	),
		// },
		// {
		// 	desc:   "Complex Schema with cache - string",
		// 	cache:  valueCache,
		// 	query:  ast.NewTerm(ast.String(`{ securityAdvisories { totalCount } }`)),
		// 	schema: ast.NewTerm(ast.String(complexGQLSchema)),
		// 	result: ast.ArrayTerm(
		// 		complexGQLQueryASTObj,
		// 		complexGQLSchemaObj,
		// 	),
		// },
		// TODO add objects
	}
	for _, bench := range benches {
		b.Run(bench.desc, func(b *testing.B) {
			for range b.N {
				var result *ast.Term
				b.StartTimer()
				err := builtinGraphQLParse(
					BuiltinContext{
						InterQueryBuiltinValueCache: bench.cache,
					},
					[]*ast.Term{bench.query, bench.schema},
					func(term *ast.Term) error {
						result = term
						return nil
					},
				)
				b.StopTimer()
				if err != nil {
					b.Fatalf("unexpected error: %s", err)
				}
				if !bench.result.Equal(result) {
					b.Errorf("Unexpected result, expected %#v, got %#v", bench.result, result)
					return
				}
				// Check each item in array result
				for i := range bench.result.Value.(*ast.Array).Len() {
					expected := bench.result.Value.(*ast.Array).Elem(i)
					actual := result.Value.(*ast.Array).Elem(i)
					if !expected.Equal(actual) {
						b.Errorf("Unexpected value at result[%d] (%s), expected %#v, got %#v", i, resultItemDescription[i], expected, actual)
						return
					}
				}
			}
		})
	}
}

func BenchmarkGraphQLParseAndVerify(b *testing.B) {

	// Share an InterQueryValueCache across multiple runs
	// Tune number of entries to exceed number of distinct GQL schemas
	in := `{"inter_query_builtin_value_cache": {"max_num_entries": 10},}`
	config, _ := cache.ParseCachingConfig([]byte(in))
	valueCache := cache.NewInterQueryValueCache(context.Background(), config)

	// Use this to map result item position to purpose for better error messages
	resultItemDescription := []string{"is_valid", "query_ast", "schema_ast"}

	benches := []struct {
		desc   string
		schema *ast.Term
		cache  cache.InterQueryValueCache
		query  *ast.Term
		result *ast.Term
	}{
		{
			desc:   "Trivial Schema - string",
			cache:  nil,
			query:  ast.NewTerm(ast.String(`{ employeeByID(id: "alice") { salary } }`)),
			schema: ast.NewTerm(ast.String(employeeGQLSchema)),
			result: ast.ArrayTerm(
				ast.BooleanTerm(true),
				ast.NewTerm(employeeGQLQueryASTObj),
				ast.NewTerm(employeeGQLSchemaASTObj),
			),
		},
		{
			desc:   "Trivial Schema with cache - string",
			cache:  valueCache,
			query:  ast.NewTerm(ast.String(`{ employeeByID(id: "alice") { salary } }`)),
			schema: ast.NewTerm(ast.String(employeeGQLSchema)),
			result: ast.ArrayTerm(
				ast.BooleanTerm(true),
				ast.NewTerm(employeeGQLQueryASTObj),
				ast.NewTerm(employeeGQLSchemaASTObj),
			),
		},
		// TODO add objects and complex schema
	}
	for _, bench := range benches {
		b.Run(bench.desc, func(b *testing.B) {
			for range b.N {
				var result *ast.Term
				b.StartTimer()
				err := builtinGraphQLParseAndVerify(
					BuiltinContext{
						InterQueryBuiltinValueCache: bench.cache,
					},
					[]*ast.Term{bench.query, bench.schema},
					func(term *ast.Term) error {
						result = term
						return nil
					},
				)
				b.StopTimer()
				if err != nil {
					b.Fatalf("unexpected error: %s", err)
				}
				if !bench.result.Equal(result) {
					b.Errorf("Unexpected result, expected %#v, got %#v", bench.result, result)
					return
				}
				// Check each item in array result
				for i := range bench.result.Value.(*ast.Array).Len() {
					expected := bench.result.Value.(*ast.Array).Elem(i)
					actual := result.Value.(*ast.Array).Elem(i)
					if !expected.Equal(actual) {
						b.Errorf("Unexpected value at result[%d] (%s), expected %#v, got %#v", i, resultItemDescription[i], expected, actual)
						return
					}
				}
			}
		})
	}
}
