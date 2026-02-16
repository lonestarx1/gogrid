// Package bench provides benchmarks for GoGrid's core patterns using
// the mock LLM provider. It measures agent execution, pipeline
// scaling, team concurrency, and shared memory contention to track
// performance regressions across releases.
//
// Run benchmarks with:
//
//	go test -bench=. ./pkg/eval/bench/
package bench
