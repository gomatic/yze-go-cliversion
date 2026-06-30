package main

import cli "github.com/urfave/cli/v3"

// version is the package-level ldflags target.
var version = "dev"

// newApp follows the gomatic dependency-injection idiom: the package `version`
// is threaded through a `version` parameter for testability (the yupsh pattern).
// `Version: version` here references the parameter — which carries the injected
// package value — and must NOT be flagged.
func newApp(version string) *cli.Command {
	return &cli.Command{Name: "app", Version: version}
}

func main() {
	_ = newApp(version)
}
