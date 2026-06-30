package main

import cli "github.com/urfave/cli/v3"

func main() {
	_ = &cli.Command{} // want `none sets Version`
}
