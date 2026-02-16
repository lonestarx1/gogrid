package cli

import (
	"flag"
	"os"
	"path/filepath"
	"text/template"

	"github.com/lonestarx1/gogrid/internal/cli/templates"
)

func (a *App) runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(a.stderr)
	tmplName := fs.String("template", "single", "project template (single, team, pipeline)")
	projName := fs.String("name", "", "project name (defaults to directory name)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}

	// Validate template.
	tmpl := templates.Get(*tmplName)
	if tmpl == nil {
		a.errf("Error: unknown template %q (valid: single, team, pipeline)\n", *tmplName)
		return 1
	}

	// Resolve project name.
	name := *projName
	if name == "" {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			a.errf("Error: %v\n", err)
			return 1
		}
		name = filepath.Base(absDir)
	}

	// Check target directory.
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			a.errf("Error: %v\n", err)
			return 1
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		a.errf("Error: %v\n", err)
		return 1
	}
	// Allow if only hidden files exist.
	for _, e := range entries {
		if e.Name()[0] != '.' {
			a.errf("Error: directory %q is not empty\n", dir)
			return 1
		}
	}

	// Render template files.
	data := templates.Data{
		Name:   name,
		Module: "github.com/example/" + name,
	}

	for _, f := range tmpl.Files {
		path := filepath.Join(dir, f.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			a.errf("Error: %v\n", err)
			return 1
		}

		t, err := template.New(f.Path).Parse(f.Content)
		if err != nil {
			a.errf("Error: parsing template %s: %v\n", f.Path, err)
			return 1
		}

		out, err := os.Create(path)
		if err != nil {
			a.errf("Error: %v\n", err)
			return 1
		}

		if err := t.Execute(out, data); err != nil {
			_ = out.Close()
			a.errf("Error: rendering %s: %v\n", f.Path, err)
			return 1
		}
		_ = out.Close()
	}

	a.outf("Created GoGrid project %q with %s template in %s\n", name, *tmplName, dir)
	a.outf("\nNext steps:\n")
	a.outf("  cd %s\n", dir)
	a.outf("  go mod tidy\n")
	a.outf("  export OPENAI_API_KEY=sk-...  # or ANTHROPIC_API_KEY / GEMINI_API_KEY\n")
	a.outf("  gogrid run <agent-name> -input \"hello\"\n")

	return 0
}
