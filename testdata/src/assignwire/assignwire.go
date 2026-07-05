package main

import cli "github.com/urfave/cli/v3"

var version = "dev"

var mixed bool

func pair() (string, bool) { return "x", true }

func main() {
	// post-construction wiring: the literal itself has no Version, but the
	// package invariant is satisfied by the assignment below — no diagnostic.
	cmd := &cli.Command{Name: "app"}
	cmd.Version = version
	// a tuple assignment (Lhs/Rhs lengths differ) is skipped, not judged.
	cmd.Version, mixed = pair()
}
