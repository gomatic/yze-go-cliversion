package lib

import cli "github.com/urfave/cli/v3"

// A non-main package is out of scope even though this command omits Version.
var Root = &cli.Command{Name: "x"}
