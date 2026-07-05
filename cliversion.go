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
// version-wired command is reported. Wiring is recognized in a keyed literal
// field, a positional literal element, or a post-construction assignment
// (`cmd.Version = version`), and through type aliases of the command types.
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

	versionFieldName = "Version"

	messageMissing = "package main builds a urfave/cli command but none sets Version to the package-level `var version`; --version will not report the build version"
	messageNotVar  = "Version must reference the package-level `var version` (or a value threaded from it through a `version` parameter), not a literal, call, const, or differently-named symbol"
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

// run collects the urfave command literals and the values assigned to command
// Version fields in a main package, and reports when none is version-wired.
func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	reportCommands(pass, collectCommands(pass, insp), collectAssignedVersions(pass, insp))
	return nil, nil
}

// collectCommands gathers every composite literal constructing a urfave command.
func collectCommands(pass *analysis.Pass, insp *inspector.Inspector) []*ast.CompositeLit {
	var commands []*ast.CompositeLit
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		lit := n.(*ast.CompositeLit)
		if isURFaveCommand(pass, lit) {
			commands = append(commands, lit)
		}
	})
	return commands
}

// collectAssignedVersions gathers every value assigned post-construction to a
// urfave command's Version field (`cmd.Version = <value>`).
func collectAssignedVersions(pass *analysis.Pass, insp *inspector.Inspector) []ast.Expr {
	var values []ast.Expr
	insp.Preorder([]ast.Node{(*ast.AssignStmt)(nil)}, func(n ast.Node) {
		values = append(values, versionAssignments(pass, n.(*ast.AssignStmt))...)
	})
	return values
}

// versionAssignments returns the right-hand values of an assignment whose
// left-hand sides select a urfave command's Version field. Tuple assignments
// (Lhs/Rhs lengths differ, e.g. `cmd.Version, ok = f()`) carry no directly
// judgeable value and are skipped.
func versionAssignments(pass *analysis.Pass, assign *ast.AssignStmt) []ast.Expr {
	if len(assign.Lhs) != len(assign.Rhs) {
		return nil
	}
	var values []ast.Expr
	for i, lhs := range assign.Lhs {
		if isCommandVersionSelector(pass, lhs) {
			values = append(values, assign.Rhs[i])
		}
	}
	return values
}

// isCommandVersionSelector reports whether expr selects the Version field of a
// urfave command value or pointer.
func isCommandVersionSelector(pass *analysis.Pass, expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != versionFieldName {
		return false
	}
	return isCommandType(pass.TypesInfo.TypeOf(sel.X))
}

// reportCommands enforces the invariant: if any command wires Version to the
// package-level `var version` — in a literal or by assignment — the package is
// clean. Otherwise it reports once — at the first bad Version value if one
// exists, else at the first command.
func reportCommands(pass *analysis.Pass, commands []*ast.CompositeLit, assigned []ast.Expr) {
	if len(commands) == 0 {
		return
	}
	var badValue ast.Expr
	for _, value := range versionValues(pass, commands, assigned) {
		if isVersionVar(pass, value) {
			return
		}
		if badValue == nil {
			badValue = value
		}
	}
	reportFailure(pass, commands[0], badValue)
}

// versionValues gathers every expression the package wires to a command
// Version: literal field values first, then post-construction assignments.
func versionValues(pass *analysis.Pass, commands []*ast.CompositeLit, assigned []ast.Expr) []ast.Expr {
	var values []ast.Expr
	for _, lit := range commands {
		if value, ok := versionField(pass, lit); ok {
			values = append(values, value)
		}
	}
	return append(values, assigned...)
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
	return isCommandType(pass.TypesInfo.TypeOf(lit))
}

// isCommandType reports whether t is (an alias of, or a pointer to) the urfave
// cli.Command or cli.App type, v3 or v2.
func isCommandType(t types.Type) bool {
	named, ok := commandNamed(t)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	fromURFave := pkg != nil && (pkg.Path() == urfaveV3 || pkg.Path() == urfaveV2)
	return fromURFave && (obj.Name() == "Command" || obj.Name() == "App")
}

// commandNamed resolves t to its named type, unaliasing (`type Cmd =
// cli.Command`) and dereferencing one pointer level (a `cmd := &cli.Command{}`
// selector base, or an elided &-literal inside a []*cli.Command slice, types as
// *Command).
func commandNamed(t types.Type) (*types.Named, bool) {
	if ptr, ok := types.Unalias(t).(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := types.Unalias(t).(*types.Named)
	return named, ok
}

// versionField returns the value the literal gives its Version field — by key
// or by position.
func versionField(pass *analysis.Pass, lit *ast.CompositeLit) (ast.Expr, bool) {
	if value, ok := keyedVersion(lit); ok {
		return value, true
	}
	return positionalVersion(pass, lit)
}

// keyedVersion returns the value of the literal's keyed Version element, if any.
func keyedVersion(lit *ast.CompositeLit) (ast.Expr, bool) {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if key, isIdent := kv.Key.(*ast.Ident); isIdent && key.Name == versionFieldName {
			return kv.Value, true
		}
	}
	return nil, false
}

// positionalVersion returns the element at the Version field's position in a
// positional (non-keyed) literal, if any. A positional struct literal must
// initialize every field, so element index i is field index i.
func positionalVersion(pass *analysis.Pass, lit *ast.CompositeLit) (ast.Expr, bool) {
	strct := commandStruct(pass, lit)
	for i, elt := range lit.Elts {
		if _, keyed := elt.(*ast.KeyValueExpr); keyed {
			return nil, false
		}
		if strct.Field(i).Name() == versionFieldName {
			return elt, true
		}
	}
	return nil, false
}

// commandStruct returns the struct underlying the literal's command type. The
// caller guarantees the literal passed isURFaveCommand, so both assertions hold
// by construction.
func commandStruct(pass *analysis.Pass, lit *ast.CompositeLit) *types.Struct {
	named, _ := commandNamed(pass.TypesInfo.TypeOf(lit))
	return named.Underlying().(*types.Struct)
}

// isVersionVar reports whether expr references a variable named `version` — the
// package-level `var version` (the ldflags target) or a value threaded from it
// through parameters (the gomatic dependency-injection idiom: `var version` ->
// run(version) -> newApp(version) -> Version: version). It deliberately accepts
// any `*types.Var` named `version`, not only the package-scope one, so the
// testable DI pattern is not a false positive; a const, literal, call (e.g.
// getVersion()), selector, or differently-named symbol is still rejected.
func isVersionVar(pass *analysis.Pass, expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	v, ok := pass.TypesInfo.ObjectOf(ident).(*types.Var)
	return ok && v.Name() == "version"
}
