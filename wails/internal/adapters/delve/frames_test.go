package delve

import (
	"testing"

	"github.com/google/go-dap"
)

func TestScopeReferencesFindsArgumentsAndLocals(t *testing.T) {
	scopes := []dap.Scope{
		{Name: "Arguments", VariablesReference: 1001},
		{Name: "Locals (warning: optimized function)", VariablesReference: 1002},
		{Name: "Globals", VariablesReference: 1003},
	}

	arguments, locals := scopeReferences(scopes)

	if arguments != 1001 || locals != 1002 {
		t.Errorf("references = %d, %d; want 1001, 1002", arguments, locals)
	}
	if a, l := scopeReferences(nil); a != 0 || l != 0 {
		t.Error("no scopes means references 0")
	}
}
