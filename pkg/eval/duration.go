package eval

import (
	"fmt"
	"time"
)

// CompletedWithin returns a Score indicating whether the observed
// duration is within the given limit. This is a standalone function
// rather than an Evaluator because wall-clock time is not part of
// agent.Result. Use Func to integrate it into a Suite.
func CompletedWithin(observed, limit time.Duration) Score {
	if observed <= limit {
		return Score{
			Pass:   true,
			Value:  1.0,
			Reason: fmt.Sprintf("completed in %s (limit %s)", observed, limit),
		}
	}
	var value float64
	if observed > 0 {
		value = float64(limit) / float64(observed)
	}
	return Score{
		Pass:   false,
		Value:  value,
		Reason: fmt.Sprintf("took %s, exceeds limit %s", observed, limit),
	}
}
