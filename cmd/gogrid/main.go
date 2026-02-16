// GoGrid (G2) CLI entry point.
package main

import (
	"os"

	"github.com/lonestarx1/gogrid/internal/cli"
)

func main() {
	app := cli.New(os.Stdout, os.Stderr)
	os.Exit(app.Run(os.Args[1:]))
}
