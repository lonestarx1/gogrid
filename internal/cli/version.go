package cli

import "runtime"

// Version is set at build time via -ldflags.
var Version = "dev"

func (a *App) runVersion() int {
	a.outf("gogrid %s (%s/%s, %s)\n",
		Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
	return 0
}
