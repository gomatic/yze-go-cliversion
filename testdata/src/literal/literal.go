package main

import cli "github.com/urfave/cli/v3"

func main() {
	_ = &cli.Command{Version: "1.0.0"} // want `must reference the package-level`
}
