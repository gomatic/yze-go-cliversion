package main

import cli "github.com/urfave/cli/v2"

var version = "dev"

func main() {
	// the v2 module path is recognized: a wired v2 App is clean, and an unwired
	// inline v2 subcommand does not break the package invariant.
	_ = &cli.App{Name: "app", Version: version}
	_ = &cli.Command{Name: "sub"}
}
