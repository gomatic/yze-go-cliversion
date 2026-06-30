package main

import cli "github.com/urfave/cli/v3"

func main() {
	_ = &cli.Command{Name: "app"} // want `none sets Version`
}
