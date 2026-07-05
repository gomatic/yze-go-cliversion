package main

import cli "github.com/urfave/cli/v3"

// Cmd aliases the urfave command type; the alias must still be recognized.
type Cmd = cli.Command

func main() {
	_ = Cmd{Name: "app"} // want `none sets Version`
}
