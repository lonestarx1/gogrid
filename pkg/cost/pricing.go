package cost

// DefaultPricing contains per-model pricing as of February 2026.
// Prices are in USD per 1 million tokens. These are configurable via
// Tracker.SetPricing and should be updated as providers change their rates.
var DefaultPricing = map[string]ModelPricing{
	// --- OpenAI ---
	"gpt-4o":       {PromptPer1M: 2.50, CompletionPer1M: 10.00},
	"gpt-4o-mini":  {PromptPer1M: 0.15, CompletionPer1M: 0.60},
	"gpt-4.1":      {PromptPer1M: 2.00, CompletionPer1M: 8.00},
	"gpt-4.1-mini": {PromptPer1M: 0.40, CompletionPer1M: 1.60},
	"gpt-4.1-nano": {PromptPer1M: 0.10, CompletionPer1M: 0.40},
	"o3":           {PromptPer1M: 2.00, CompletionPer1M: 8.00},
	"o4-mini":      {PromptPer1M: 1.10, CompletionPer1M: 4.40},

	// --- Anthropic ---
	"claude-opus-4-6-20250827":   {PromptPer1M: 5.00, CompletionPer1M: 25.00},
	"claude-opus-4-5-20250620":   {PromptPer1M: 5.00, CompletionPer1M: 25.00},
	"claude-sonnet-4-5-20250929": {PromptPer1M: 3.00, CompletionPer1M: 15.00},
	"claude-sonnet-4-0-20250514": {PromptPer1M: 3.00, CompletionPer1M: 15.00},
	"claude-haiku-4-5-20251001":  {PromptPer1M: 1.00, CompletionPer1M: 5.00},

	// --- Google Gemini ---
	"gemini-3-pro":     {PromptPer1M: 2.00, CompletionPer1M: 12.00},
	"gemini-3-flash":   {PromptPer1M: 0.50, CompletionPer1M: 3.00},
	"gemini-2.5-pro":   {PromptPer1M: 1.25, CompletionPer1M: 10.00},
	"gemini-2.5-flash": {PromptPer1M: 0.15, CompletionPer1M: 0.60},
	"gemini-2.0-flash": {PromptPer1M: 0.10, CompletionPer1M: 0.40},
}
