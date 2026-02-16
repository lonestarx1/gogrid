package bench

import (
	"context"
	"fmt"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/orchestrator/pipeline"
)

func BenchmarkPipelineThreeStages(b *testing.B) {
	p := pipeline.New("bench-pipeline",
		pipeline.WithStages(
			pipeline.Stage{Name: "stage-1", Agent: newTestAgent("s1", "output-1")},
			pipeline.Stage{Name: "stage-2", Agent: newTestAgent("s2", "output-2")},
			pipeline.Stage{Name: "stage-3", Agent: newTestAgent("s3", "output-3")},
		),
	)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, err := p.Run(ctx, "pipeline input")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPipelineScaling(b *testing.B) {
	for _, n := range []int{1, 3, 5, 10} {
		b.Run(fmt.Sprintf("stages=%d", n), func(b *testing.B) {
			stages := make([]pipeline.Stage, n)
			for i := range n {
				stages[i] = pipeline.Stage{
					Name:  fmt.Sprintf("stage-%d", i),
					Agent: newTestAgent(fmt.Sprintf("s%d", i), fmt.Sprintf("output-%d", i)),
				}
			}
			p := pipeline.New("bench-scaling-pipeline", pipeline.WithStages(stages...))
			ctx := context.Background()

			b.ResetTimer()
			for b.Loop() {
				_, err := p.Run(ctx, "pipeline input")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
