package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/lonestarx1/gogrid/internal/runrecord"
	"github.com/lonestarx1/gogrid/pkg/llm"
)

func (a *App) runCost(args []string) int {
	fs := flag.NewFlagSet("cost", flag.ContinueOnError)
	fs.SetOutput(a.stderr)
	jsonOutput := fs.Bool("json", false, "output as JSON")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	// No run-id: list all runs with cost.
	if fs.NArg() == 0 {
		return a.listRunCosts(*jsonOutput)
	}

	runID := fs.Arg(0)
	rec, err := runrecord.Load(".", runID)
	if err != nil {
		a.errf("Error: %v\n", err)
		return 1
	}

	if *jsonOutput {
		return a.costJSON(rec)
	}

	a.renderCostTable(rec)
	return 0
}

func (a *App) listRunCosts(jsonOut bool) int {
	ids, err := runrecord.List(".")
	if err != nil {
		a.errf("Error: %v\n", err)
		return 1
	}
	if len(ids) == 0 {
		a.outf("No runs found. Run 'gogrid run <agent>' first.\n")
		return 0
	}

	type runSummary struct {
		RunID string  `json:"run_id"`
		Agent string  `json:"agent"`
		Model string  `json:"model"`
		Cost  float64 `json:"cost"`
	}

	var summaries []runSummary
	for _, id := range ids {
		rec, err := runrecord.Load(".", id)
		if err != nil {
			continue
		}
		summaries = append(summaries, runSummary{
			RunID: rec.RunID,
			Agent: rec.Agent,
			Model: rec.Model,
			Cost:  rec.Cost,
		})
	}

	if jsonOut {
		data, _ := json.MarshalIndent(summaries, "", "  ")
		a.outf("%s\n", data)
		return 0
	}

	w := tabwriter.NewWriter(a.stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "RUN ID\tAGENT\tMODEL\tCOST")
	for _, s := range summaries {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t$%.6f\n", s.RunID, s.Agent, s.Model, s.Cost)
	}
	_ = w.Flush()
	return 0
}

type modelCost struct {
	Model string    `json:"model"`
	Calls int       `json:"calls"`
	Usage llm.Usage `json:"usage"`
	Cost  float64   `json:"cost"`
}

func (a *App) costJSON(rec *runrecord.Record) int {
	models := aggregateByModel(rec)
	data, err := json.MarshalIndent(models, "", "  ")
	if err != nil {
		a.errf("Error: %v\n", err)
		return 1
	}
	a.outf("%s\n", data)
	return 0
}

func (a *App) renderCostTable(rec *runrecord.Record) {
	a.outf("Run: %s\n\n", rec.RunID)

	models := aggregateByModel(rec)

	w := tabwriter.NewWriter(a.stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "MODEL\tCALLS\tPROMPT\tCOMPLETION\tCOST")

	var totalCalls int
	var totalPrompt, totalCompletion int
	var totalCost float64
	for _, m := range models {
		_, _ = fmt.Fprintf(w, "%s\t%d\t%d\t%d\t$%.6f\n",
			m.Model, m.Calls, m.Usage.PromptTokens, m.Usage.CompletionTokens, m.Cost)
		totalCalls += m.Calls
		totalPrompt += m.Usage.PromptTokens
		totalCompletion += m.Usage.CompletionTokens
		totalCost += m.Cost
	}

	_, _ = fmt.Fprintln(w, strings.Repeat("\u2500", 60)+"\t\t\t\t")
	_, _ = fmt.Fprintf(w, "TOTAL\t%d\t%d\t%d\t$%.6f\n",
		totalCalls, totalPrompt, totalCompletion, totalCost)
	_ = w.Flush()
}

func aggregateByModel(rec *runrecord.Record) []modelCost {
	byModel := make(map[string]*modelCost)

	if len(rec.CostRecords) > 0 {
		for _, cr := range rec.CostRecords {
			mc, ok := byModel[cr.Model]
			if !ok {
				mc = &modelCost{Model: cr.Model}
				byModel[cr.Model] = mc
			}
			mc.Calls++
			mc.Usage.PromptTokens += cr.Usage.PromptTokens
			mc.Usage.CompletionTokens += cr.Usage.CompletionTokens
			mc.Usage.TotalTokens += cr.Usage.TotalTokens
			mc.Cost += cr.Cost
		}
	} else {
		// Fallback: use the record's aggregate usage.
		byModel[rec.Model] = &modelCost{
			Model: rec.Model,
			Calls: rec.Turns,
			Usage: rec.Usage,
			Cost:  rec.Cost,
		}
	}

	// Sort by model name.
	names := make([]string, 0, len(byModel))
	for name := range byModel {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]modelCost, 0, len(names))
	for _, name := range names {
		result = append(result, *byModel[name])
	}
	return result
}
