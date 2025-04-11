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
	"runtime"
	"time"
	_ "unsafe" // needed for go:linkname

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown"
)

// Complex GQL schema to test with
const gqlSchema = `https://docs.github.com/public/fpt/schema.docs.graphql`

// Seconds between memory dump information
const memDumpSeconds = 3

//go:linkname  builtinGraphQLParseSchema github.com/open-policy-agent/opa/v1/topdown.builtinGraphQLParseSchema
func builtinGraphQLParseSchema(topdown.BuiltinContext, []*ast.Term, func(*ast.Term) error) error

func main() {
	b, err := downloadSchema(gqlSchema)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading: %s\n", err)
	}

	// Dump the memory usage every few seconds
	ticker := time.NewTicker(memDumpSeconds * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				PrintMemUsage()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

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

	// Stop printing memory usage
	done <- true
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

// From: https://gist.github.com/j33ty/79e8b736141be19687f565ea4c6f4226
// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
