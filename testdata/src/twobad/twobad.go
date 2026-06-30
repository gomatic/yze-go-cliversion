package main

import cli "github.com/urfave/cli/v3"

func main() {
	// two mis-wired commands: the package is reported once, at the first.
	_ = &cli.Command{Version: "1.0"} // want `must reference the package-level`
	_ = &cli.App{Version: "2.0"}
}
