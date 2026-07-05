package main

import cli "github.com/urfave/cli/v3"

func main() {
	// positional (non-keyed) literal: Version is positionally set to a string
	// literal, not `var version` — the wrong-value failure.
	_ = cli.Command{"app", "1.0", nil} // want `must reference the package-level`
}
