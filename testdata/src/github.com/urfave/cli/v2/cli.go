// Package cli is a minimal stub of urfave/cli v2 for analysistest fixtures.
package cli

// Command is a stand-in for the v2 Command type.
type Command struct {
	Name, Version string
	Commands      []*Command
}

// App is a stand-in for the v2 App type.
type App struct{ Name, Version string }

// StringFlag is a urfave type that is NOT a command (exercises the non-command path).
type StringFlag struct{ Name string }
