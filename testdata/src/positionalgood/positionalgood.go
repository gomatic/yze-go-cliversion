package main

import cli "github.com/urfave/cli/v3"

var version = "dev"

func main() {
	// positional (non-keyed) literal with Version wired to the package-level
	// `var version` — no diagnostic.
	_ = cli.Command{"app", version, nil}
}
