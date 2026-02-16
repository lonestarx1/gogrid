package eval

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
)

const defaultPromptTemplate = `You are evaluating an AI agent's output against a rubric.

Rubric:
{rubric}

Agent output:
{output}

Evaluate the output against the rubric. Respond with exactly two lines:
SCORE: <integer from 0 to 10>
REASON: <brief explanation>`

// LLMJudge uses an LLM provider to evaluate agent output against a rubric.
// The judge sends the output and rubric to the LLM and parses a structured
// SCORE/REASON response. Scores >= 7 (out of 10) are considered passing.
type LLMJudge struct {
	provider       llm.Provider
	model          string
	rubric         string
	promptTemplate string
}

// NewLLMJudge creates an LLMJudge evaluator.
func NewLLMJudge(provider llm.Provider, model, rubric string) *LLMJudge {
	return &LLMJudge{
		provider:       provider,
		model:          model,
		rubric:         rubric,
		promptTemplate: defaultPromptTemplate,
	}
}

// WithPromptTemplate sets a custom prompt template. The template should
// contain {rubric} and {output} placeholders.
func (j *LLMJudge) WithPromptTemplate(tmpl string) *LLMJudge {
	j.promptTemplate = tmpl
	return j
}

// Name returns "llm_judge".
func (j *LLMJudge) Name() string { return "llm_judge" }

// Evaluate sends the agent output and rubric to the LLM provider,
// then parses the SCORE/REASON response.
func (j *LLMJudge) Evaluate(ctx context.Context, result *agent.Result) (Score, error) {
	prompt := j.promptTemplate
	prompt = strings.ReplaceAll(prompt, "{rubric}", j.rubric)
	prompt = strings.ReplaceAll(prompt, "{output}", result.Message.Content)

	resp, err := j.provider.Complete(ctx, llm.Params{
		Model: j.model,
		Messages: []llm.Message{
			llm.NewUserMessage(prompt),
		},
	})
	if err != nil {
		return Score{}, fmt.Errorf("llm_judge: provider error: %w", err)
	}

	rawScore, reason, err := parseJudgeResponse(resp.Message.Content)
	if err != nil {
		return Score{}, fmt.Errorf("llm_judge: %w", err)
	}

	value := float64(rawScore) / 10.0
	pass := rawScore >= 7

	return Score{
		Pass:   pass,
		Value:  value,
		Reason: reason,
	}, nil
}

// parseJudgeResponse extracts SCORE and REASON from the LLM response.
// It uses case-insensitive prefix matching.
func parseJudgeResponse(content string) (int, string, error) {
	var scoreVal int
	var reason string
	foundScore := false
	foundReason := false

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		upper := strings.ToUpper(trimmed)

		if strings.HasPrefix(upper, "SCORE:") {
			s := strings.TrimSpace(trimmed[len("SCORE:"):])
			val, err := strconv.Atoi(s)
			if err != nil {
				return 0, "", fmt.Errorf("invalid score %q: %w", s, err)
			}
			if val < 0 || val > 10 {
				return 0, "", fmt.Errorf("score %d out of range [0, 10]", val)
			}
			scoreVal = val
			foundScore = true
		}

		if strings.HasPrefix(upper, "REASON:") {
			reason = strings.TrimSpace(trimmed[len("REASON:"):])
			foundReason = true
		}
	}

	if !foundScore {
		return 0, "", fmt.Errorf("no SCORE line found in response: %q", content)
	}
	if !foundReason {
		reason = "no reason provided"
	}

	return scoreVal, reason, nil
}
