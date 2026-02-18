// Package cli implements the GoGrid command-line interface.
package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/lonestarx1/gogrid/pkg/llm"
)

// ProviderFactory creates an LLM provider by name.
// The default implementation resolves API keys from environment variables.
type ProviderFactory func(ctx context.Context, name string) (llm.Provider, error)

// App is the GoGrid CLI application.
type App struct {
	stdout          io.Writer
	stderr          io.Writer
	providerFactory ProviderFactory
}

// New creates a CLI application that writes to the given writers.
func New(stdout, stderr io.Writer) *App {
	return &App{
		stdout:          stdout,
		stderr:          stderr,
		providerFactory: defaultProviderFactory,
	}
}

// SetProviderFactory overrides the default provider factory (for testing).
func (a *App) SetProviderFactory(f ProviderFactory) {
	a.providerFactory = f
}

// Run dispatches to the appropriate subcommand and returns an exit code.
func (a *App) Run(args []string) int {
	if len(args) == 0 {
		a.printUsage()
		return 0
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "version":
		return a.runVersion()
	case "init":
		return a.runInit(cmdArgs)
	case "list":
		return a.runList(cmdArgs)
	case "run":
		return a.runRun(cmdArgs)
	case "trace":
		return a.runTrace(cmdArgs)
	case "cost":
		return a.runCost(cmdArgs)
	case "help", "-h", "--help":
		a.printUsage()
		return 0
	default:
		a.errf("unknown command: %s\n\n", cmd)
		a.printUsage()
		return 1
	}
}

func (a *App) printUsage() {
	a.outf(`gogrid — Develop and orchestrate AI agents

Usage: gogrid <command> [flags]

Commands:
  init      Scaffold a new GoGrid project
  list      List agents defined in gogrid.yaml
  run       Execute an agent
  trace     Inspect execution traces
  cost      View cost breakdown
  version   Print version information
  help      Show this help message

Run 'gogrid <command> -h' for command-specific help.
`)
}

// outf writes to stdout, ignoring write errors (terminal I/O).
func (a *App) outf(format string, args ...any) {
	_, _ = fmt.Fprintf(a.stdout, format, args...)
}

// errf writes to stderr, ignoring write errors (terminal I/O).
func (a *App) errf(format string, args ...any) {
	_, _ = fmt.Fprintf(a.stderr, format, args...)
}
