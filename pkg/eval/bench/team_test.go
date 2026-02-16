package bench

import (
	"context"
	"fmt"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/orchestrator/team"
)

func BenchmarkTeamTwoMembers(b *testing.B) {
	t := team.New("bench-team",
		team.WithMembers(
			team.Member{Agent: newTestAgent("agent-a", "response-a")},
			team.Member{Agent: newTestAgent("agent-b", "response-b")},
		),
	)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, err := t.Run(ctx, "team input")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTeamScaling(b *testing.B) {
	for _, n := range []int{1, 2, 5, 10, 20} {
		b.Run(fmt.Sprintf("members=%d", n), func(b *testing.B) {
			members := make([]team.Member, n)
			for i := range n {
				members[i] = team.Member{
					Agent: newTestAgent(fmt.Sprintf("agent-%d", i), fmt.Sprintf("response-%d", i)),
				}
			}
			t := team.New("bench-scaling-team", team.WithMembers(members...))
			ctx := context.Background()

			b.ResetTimer()
			for b.Loop() {
				_, err := t.Run(ctx, "team input")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
