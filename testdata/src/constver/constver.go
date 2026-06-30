package main

import cli "github.com/urfave/cli/v3"

// version is a const, not a var — ldflags -X cannot set it, so it is wrong.
const version = "dev"

func main() {
	_ = &cli.Command{Version: version} // want `must reference the package-level`
}
