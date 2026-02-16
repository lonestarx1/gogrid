package bench

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/memory/shared"
)

func BenchmarkSharedMemorySaveLoad(b *testing.B) {
	mem := shared.New()
	ctx := context.Background()
	msgs := []llm.Message{
		llm.NewUserMessage("hello"),
		llm.NewAssistantMessage("world"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = mem.Save(ctx, "bench-key", msgs)
		_, _ = mem.Load(ctx, "bench-key")
	}
}

func BenchmarkSharedMemoryContention(b *testing.B) {
	for _, writers := range []int{1, 2, 5, 10} {
		b.Run(fmt.Sprintf("writers=%d", writers), func(b *testing.B) {
			mem := shared.New()
			ctx := context.Background()
			msgs := []llm.Message{
				llm.NewAssistantMessage("response"),
			}

			b.ReportAllocs()
			b.ResetTimer()

			var wg sync.WaitGroup
			iterations := b.N / writers
			if iterations < 1 {
				iterations = 1
			}

			for w := range writers {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					key := fmt.Sprintf("writer-%d", id)
					for range iterations {
						_ = mem.Save(ctx, key, msgs)
						_, _ = mem.Load(ctx, key)
					}
				}(w)
			}
			wg.Wait()
		})
	}
}
