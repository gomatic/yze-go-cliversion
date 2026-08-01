package cliversion

// White-box test for the assertion commandStruct makes by construction.

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
)

// TestCommandStructGuaranteeHoldsForEveryShapeThatReachesIt names
// commandStruct's claim. It asserts twice without checking — `named` is
// dereferenced and its underlying is type-asserted to *types.Struct — on the
// stated guarantee that the caller already established isURFaveCommand. A
// guarantee nothing verifies is a nil dereference or a failed type assertion
// waiting for the first shape nobody considered, inside a linter, on a user's
// code.
//
// The shapes below are the ones isURFaveCommand admits: the type itself, a
// pointer to it, and an alias of either. Each must reach commandStruct and come
// back with the struct — which is what makes the two bare assertions safe.
func TestCommandStructGuaranteeHoldsForEveryShapeThatReachesIt(t *testing.T) {
	t.Parallel()
	pass, lits := literalsIn(t, `package p

type Command struct {
	Name    string
	Version string
}

type Aliased = Command

var (
	direct  = Command{}
	aliased = Aliased{}
	pointed = &Command{}
)
`)
	require.Len(t, lits, 3, "the fixture must supply every admitted shape")

	for i, lit := range lits {
		require.True(t, isCommandLike(pass, lit), "literal %d must be command-shaped for this test to mean anything", i)

		var got *types.Struct
		assert.NotPanics(t, func() { got = commandStruct(pass, lit) },
			"literal %d must not panic: the guarantee is what makes the assertions bare", i)
		require.NotNil(t, got)
		assert.Equal(t, 2, got.NumFields(), "the struct that comes back is the command's own")
	}
}

// isCommandLike is the precondition commandStruct rests on, expressed against
// the fixture's own type rather than urfave's: commandNamed must resolve, which
// is exactly what isURFaveCommand establishes for a real cli.Command.
func isCommandLike(pass *analysis.Pass, lit *ast.CompositeLit) bool {
	named, ok := commandNamed(pass.TypesInfo.TypeOf(lit))
	if !ok {
		return false
	}
	_, isStruct := named.Underlying().(*types.Struct)
	return isStruct
}

// literalsIn type-checks src and returns a pass plus every composite literal it
// declares, in source order.
func literalsIn(t *testing.T, src string) (*analysis.Pass, []*ast.CompositeLit) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", src, 0)
	require.NoError(t, err)
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}, Defs: map[*ast.Ident]types.Object{}}
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("example.test/p", fset, []*ast.File{file}, info)
	require.NoError(t, err)

	var lits []*ast.CompositeLit
	ast.Inspect(file, func(n ast.Node) bool {
		if lit, ok := n.(*ast.CompositeLit); ok {
			lits = append(lits, lit)
		}
		return true
	})
	return &analysis.Pass{Fset: fset, Files: []*ast.File{file}, Pkg: pkg, TypesInfo: info}, lits
}
