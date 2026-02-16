package cli

import (
	"flag"
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/lonestarx1/gogrid/internal/config"
)

func (a *App) runList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(a.stderr)
	configPath := fs.String("config", "gogrid.yaml", "path to gogrid.yaml")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		a.errf("Error: %v\n", err)
		return 1
	}

	// Sort agent names for stable output.
	names := make([]string, 0, len(cfg.Agents))
	for name := range cfg.Agents {
		names = append(names, name)
	}
	sort.Strings(names)

	w := tabwriter.NewWriter(a.stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tPROVIDER\tMODEL")
	for _, name := range names {
		agent := cfg.Agents[name]
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", name, agent.Provider, agent.Model)
	}
	_ = w.Flush()

	return 0
}
