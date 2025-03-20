package main

// Run with:
// go run reproducer.go
// or
// go build -o reproducer reproducer.go && ./reproducer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	_ "unsafe" // needed for go:linkname

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown"
)

// Complex GQL schema to test with
const gqlSchema = `https://docs.github.com/public/fpt/schema.docs.graphql`

//go:linkname  builtinGraphQLParseSchema github.com/open-policy-agent/opa/v1/topdown.builtinGraphQLParseSchema
func builtinGraphQLParseSchema(topdown.BuiltinContext, []*ast.Term, func(*ast.Term) error) error

func main() {
	b, err := downloadSchema(gqlSchema)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading: %s\n", err)
	}

	// This should allocate a bunch of memory and not return
	fmt.Fprintf(os.Stderr, "Now passing %d bytes to builtinGraphQLParseSchema() to reproduce the issue in ast.InterfaceToValue()\n", len(b))
	_ = builtinGraphQLParseSchema(
		topdown.BuiltinContext{Context: context.Background()},
		[]*ast.Term{
			ast.NewTerm(ast.String(string(b))),
		},
		func(term *ast.Term) error {
			return nil
		},
	)
}

func downloadSchema(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while downloading %s - %s\n", url, err)
		return []byte{}, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}
