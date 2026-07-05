package main

import cli "github.com/urfave/cli/v3"

func main() {
	// post-construction wiring to a literal does not satisfy the invariant and
	// is reported as the wrong-value failure, at the assigned value.
	cmd := cli.Command{Name: "app"}
	cmd.Version = "1.0" // want `must reference the package-level`
}
