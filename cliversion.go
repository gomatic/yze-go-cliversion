// Package cliversion provides a go/analysis analyzer enforcing the gomatic CLI
// standard that a urfave/cli `package main` exposes --version: a urfave command
// literal sets its Version field to the package-level `var version` (the ldflags
// `-X main.version` target). It verifies the GO side only — the goreleaser
// ldflag↔symbol match is a separate stickler-level YAML check.
//
// Scope is the `cmd/<app>/main.go` package. The rule is a package-level
// invariant — "main builds a urfave command whose Version is `var version`" — so
// extra command literals (e.g. an inline subcommand, which legitimately has no
// Version) never produce a false positive; only the absence of ANY properly
// version-wired command is reported.
package cliversion

import (
	"go/ast"
	"go/types"

	goyze "github.com/gomatic/go-yze"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	urfaveV3 = "github.com/urfave/cli/v3"
	urfaveV2 = "github.com/urfave/cli/v2"

	messageMissing = "package main builds a urfave/cli command but none sets Version to the package-level `var version`; --version will not report the build version"
	messageNotVar  = "Version must reference the package-level `var version`, not a literal or a differently-named symbol"
)

// Analyzer reports a urfave/cli main package that does not wire a command's
// Version to the package-level `var version`.
var Analyzer = &analysis.Analyzer{
	Name:     "cliversion",
	Doc:      "reports a urfave/cli main package that does not wire a command Version to `var version`",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// Registration declares this analyzer to the yze framework.
var Registration = goyze.Registration{
	Name:       "cliversion",
	Categories: []goyze.Category{"cli"},
	URL:        "https://docs.gomatic.dev/yze/cliversion",
	Analyzer:   Analyzer,
}

// run collects the urfave command literals in a main package and reports when
// none is version-wired.
func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	var commands []*ast.CompositeLit
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		lit := n.(*ast.CompositeLit)
		if isURFaveCommand(pass, lit) {
			commands = append(commands, lit)
		}
	})
	reportCommands(pass, commands)
	return nil, nil
}

// reportCommands enforces the invariant: if any command wires Version to the
// package-level `var version`, the package is clean. Otherwise it reports once —
// at the first bad Version value if one exists, else at the first command.
func reportCommands(pass *analysis.Pass, commands []*ast.CompositeLit) {
	if len(commands) == 0 {
		return
	}
	var badValue ast.Expr
	for _, lit := range commands {
		value, ok := versionField(lit)
		switch {
		case !ok:
			continue
		case isPackageVersionVar(pass, value):
			return
		case badValue == nil:
			badValue = value
		}
	}
	reportFailure(pass, commands[0], badValue)
}

// reportFailure emits the single diagnostic for a package with no version-wired
// command: messageNotVar when a Version was set to the wrong thing, else
// messageMissing.
func reportFailure(pass *analysis.Pass, firstCommand *ast.CompositeLit, badValue ast.Expr) {
	if badValue != nil {
		pass.Reportf(badValue.Pos(), messageNotVar)
		return
	}
	pass.Reportf(firstCommand.Pos(), messageMissing)
}

// isURFaveCommand reports whether the literal constructs a urfave cli.Command or
// cli.App value (v3 or v2).
func isURFaveCommand(pass *analysis.Pass, lit *ast.CompositeLit) bool {
	named, ok := pass.TypesInfo.TypeOf(lit).(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	fromURFave := pkg != nil && (pkg.Path() == urfaveV3 || pkg.Path() == urfaveV2)
	return fromURFave && (obj.Name() == "Command" || obj.Name() == "App")
}

// versionField returns the value assigned to the literal's Version field, if any.
func versionField(lit *ast.CompositeLit) (ast.Expr, bool) {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if key, isIdent := kv.Key.(*ast.Ident); isIdent && key.Name == "Version" {
			return kv.Value, true
		}
	}
	return nil, false
}

// isPackageVersionVar reports whether expr is the package-level `var version`.
func isPackageVersionVar(pass *analysis.Pass, expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	v, ok := pass.TypesInfo.ObjectOf(ident).(*types.Var)
	return ok && v.Name() == "version" && v.Parent() == pass.Pkg.Scope()
}
