package parse

import (
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
)

// struct to hold the result of a decoding operation
type DecodeResult struct {
	Diags   hcl.Diagnostics
	Depends map[string]*modconfig.ResourceDependency
}

func newDecodeResult() *DecodeResult {
	return &DecodeResult{Depends: make(map[string]*modconfig.ResourceDependency)}
}

// Merge merges this decode result with another
func (p *DecodeResult) Merge(other *DecodeResult) *DecodeResult {
	p.Diags = append(p.Diags, other.Diags...)
	for k, v := range other.Depends {
		p.Depends[k] = v
	}

	return p
}

// Success returns if the was parsing successful - true if there are no errors and no dependencies
func (p *DecodeResult) Success() bool {
	return !p.Diags.HasErrors() && len(p.Depends) == 0
}

// if the diags contains dependency errors, add dependencies to the result
// otherwise add diags to the result
func (p *DecodeResult) handleDecodeDiags(diags hcl.Diagnostics) {
	for _, diag := range diags {
		if dependency := diagsToDependency(diag); dependency != nil {
			p.Depends[dependency.String()] = dependency
		}
	}
	// only register errors if there are NOT any missing variables
	if len(p.Depends) == 0 {
		p.addDiags(diags)
	}
}

// determine whether the diag is a dependency error, and if so, return a dependency object
func diagsToDependency(diag *hcl.Diagnostic) *modconfig.ResourceDependency {
	if slices.Contains[[]string, string](missingVariableErrors, diag.Summary) {
		return &modconfig.ResourceDependency{Range: diag.Expression.Range(), Traversals: diag.Expression.Variables()}
	}
	return nil
}

func (p *DecodeResult) addDiags(diags hcl.Diagnostics) {
	p.Diags = append(p.Diags, diags...)
}
