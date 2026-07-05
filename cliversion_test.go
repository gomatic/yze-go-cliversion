package cliversion_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"

	cliversion "github.com/gomatic/yze-go-cliversion"
)

func TestVersionWiringIsEnforced(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), cliversion.Analyzer,
		"good", "diparam", "missing", "empty", "positional", "positionalgood",
		"literal", "wrongname", "constver", "selector", "callexpr", "twobad",
		"nonurfave", "lib", "aliased", "assignwire", "assignbad", "v2app")
}

func TestRegistrationIsWellFormed(t *testing.T) {
	assert.NoError(t, cliversion.Registration.Validate())
	assert.Equal(t, "yze/cliversion", cliversion.Registration.RuleID())
	assert.Same(t, cliversion.Analyzer, cliversion.Registration.Analyzer)
}
