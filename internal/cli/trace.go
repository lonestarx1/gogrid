package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/lonestarx1/gogrid/internal/runrecord"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

func (a *App) runTrace(args []string) int {
	fs := flag.NewFlagSet("trace", flag.ContinueOnError)
	fs.SetOutput(a.stderr)
	jsonOutput := fs.Bool("json", false, "output as JSON")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	// No run-id: list recent runs.
	if fs.NArg() == 0 {
		return a.listRecentRuns()
	}

	runID := fs.Arg(0)
	rec, err := runrecord.Load(".", runID)
	if err != nil {
		a.errf("Error: %v\n", err)
		return 1
	}

	if *jsonOutput {
		data, err := json.MarshalIndent(rec.Spans, "", "  ")
		if err != nil {
			a.errf("Error: %v\n", err)
			return 1
		}
		a.outf("%s\n", data)
		return 0
	}

	a.renderSpanTree(rec)
	return 0
}

func (a *App) listRecentRuns() int {
	ids, err := runrecord.List(".")
	if err != nil {
		a.errf("Error: %v\n", err)
		return 1
	}
	if len(ids) == 0 {
		a.outf("No runs found. Run 'gogrid run <agent>' first.\n")
		return 0
	}

	a.outf("Recent runs:\n")
	limit := 10
	if len(ids) < limit {
		limit = len(ids)
	}
	for _, id := range ids[:limit] {
		rec, err := runrecord.Load(".", id)
		if err != nil {
			a.outf("  %s (error loading)\n", id)
			continue
		}
		errMark := ""
		if rec.Error != "" {
			errMark = " [ERROR]"
		}
		a.outf("  %s  %s  %s  %s%s\n",
			id, rec.Agent, rec.Model, formatDuration(rec.Duration), errMark)
	}
	return 0
}

func (a *App) renderSpanTree(rec *runrecord.Record) {
	a.outf("Run: %s\n", rec.RunID)
	a.outf("Agent: %s | Model: %s | Duration: %s\n\n",
		rec.Agent, rec.Model, formatDuration(rec.Duration))

	if len(rec.Spans) == 0 {
		a.outf("(no spans recorded)\n")
		return
	}

	// Build parent-child map.
	children := make(map[string][]*trace.Span)
	var roots []*trace.Span
	for _, s := range rec.Spans {
		if s.ParentID == "" {
			roots = append(roots, s)
		} else {
			children[s.ParentID] = append(children[s.ParentID], s)
		}
	}

	for _, root := range roots {
		a.printSpan(root, children, "", true)
	}
}

func (a *App) printSpan(s *trace.Span, children map[string][]*trace.Span, prefix string, isLast bool) {
	connector := "\u251c\u2500\u2500 "
	if isLast {
		connector = "\u2514\u2500\u2500 "
	}
	if prefix == "" && isLast {
		connector = ""
	}

	dur := formatDuration(s.EndTime.Sub(s.StartTime))
	detail := spanDetail(s)
	a.outf("%s%s%s (%s)%s\n", prefix, connector, s.Name, dur, detail)

	childPrefix := prefix
	if prefix != "" || !isLast {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "\u2502   "
		}
	}

	kids := children[s.ID]
	for i, child := range kids {
		a.printSpan(child, children, childPrefix, i == len(kids)-1)
	}
}

func spanDetail(s *trace.Span) string {
	var parts []string
	if v, ok := s.Attributes["llm.prompt_tokens"]; ok {
		parts = append(parts, "prompt: "+v)
	}
	if v, ok := s.Attributes["llm.completion_tokens"]; ok {
		parts = append(parts, "completion: "+v)
	}
	if v, ok := s.Attributes["tool.name"]; ok {
		parts = append(parts, "\""+v+"\"")
	}
	if s.Error != "" {
		parts = append(parts, "ERROR: "+s.Error)
	}
	if len(parts) == 0 {
		return ""
	}
	return " [" + strings.Join(parts, ", ") + "]"
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%d\u00b5s", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
