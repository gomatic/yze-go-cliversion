package main

import cli "github.com/urfave/cli/v3"

func getVersion() string { return "x" }

func main() {
	// the old getVersion() wrapper pattern — a call expr, must be flagged.
	_ = &cli.Command{Version: getVersion()} // want `must reference the package-level`
}
