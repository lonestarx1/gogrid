// Package mock provides a configurable mock LLM provider for testing
// GoGrid agents, teams, pipelines, and graphs. It supports sequential
// responses, error injection, latency simulation, and call recording.
//
// Options:
//
//   - WithResponses — queue a sequence of pre-programmed responses
//   - WithFallback — response returned after the sequence is exhausted
//   - WithError — inject a constant or count-limited error
//   - WithFailCount — fail only the first N calls
//   - WithDelay — simulate LLM latency (respects context cancellation)
//   - WithCallback — observe call arguments in tests
//
// The mock provider implements llm.Provider and is safe for concurrent use.
// All examples in the examples/ directory use the mock provider so they
// run without API keys.
package mock
