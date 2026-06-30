package main

import cli "github.com/urfave/cli/v3"

var version = "dev"

func main() {
	// canonical root: pointer literal, Version wired to the package var, with an
	// inline subcommand that legitimately has no Version (must NOT be flagged).
	_ = &cli.Command{
		Name:    "app",
		Version: version,
		Commands: []*cli.Command{
			{Name: "sub"},
		},
	}
	// App is also a recognized command type.
	_ = &cli.App{Name: "alt", Version: version}
	// A urfave non-command type is ignored.
	_ = cli.StringFlag{Name: "f"}
}
