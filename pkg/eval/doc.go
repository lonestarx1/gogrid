// Package eval provides an evaluation framework for scoring GoGrid
// agent outputs. It includes built-in evaluators for exact match,
// substring containment, cost budgets, tool usage, and LLM-as-judge,
// plus a Suite for running multiple evaluators against a single result.
//
// Built-in evaluators:
//
//   - ExactMatch — output must equal an expected string
//   - Contains — output must include specified substrings
//   - CostWithin — run cost must stay within a USD budget
//   - ToolUse — conversation history must include expected tool calls
//   - LLMJudge — an LLM scores the output against a rubric (0-10)
//   - CompletedWithin — standalone function for duration checks
//
// Use Func to wrap any function as an Evaluator, and Suite to run
// multiple evaluators against a single agent.Result.
package eval
