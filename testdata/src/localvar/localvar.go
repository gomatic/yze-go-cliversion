package main

import cli "github.com/urfave/cli/v3"

func main() {
	// a function-local var named version is not the package-level ldflag target.
	version := "dev"
	_ = &cli.Command{Version: version} // want `must reference the package-level`
}
