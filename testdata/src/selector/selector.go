package main

import cli "github.com/urfave/cli/v3"

var cfg = struct{ Version string }{Version: "x"}

func main() {
	// Version set via a selector expression — must not crash, must be flagged.
	_ = &cli.Command{Version: cfg.Version} // want `must reference the package-level`
}
