// Package pipeline implements the Pipeline (Linear) orchestration pattern.
//
// A pipeline is a sequential chain of stages where each stage's agent
// processes input and hands off state to the next stage. State ownership
// is enforced via the memory/transfer package — once a stage completes,
// its handle is invalidated and only the next stage can access the data.
//
// Pipelines support stage-level retry policies, checkpointing for
// recovery, per-stage and pipeline-level timeouts and cost budgets,
// and progress reporting via trace spans and an optional callback.
package pipeline
