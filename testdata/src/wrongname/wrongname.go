package main

import cli "github.com/urfave/cli/v3"

// appVersion is the wrong (non-canonical) variable name.
var appVersion = "dev"

func main() {
	_ = &cli.Command{Version: appVersion} // want `must reference the package-level`
}
