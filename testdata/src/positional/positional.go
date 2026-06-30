package main

import cli "github.com/urfave/cli/v3"

func main() {
	// positional (non-keyed) literal: no keyed Version field — must not crash.
	_ = cli.Command{"app", "1.0", nil} // want `none sets Version`
}
