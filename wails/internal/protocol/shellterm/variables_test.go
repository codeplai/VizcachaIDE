package shellterm

import (
	"slices"
	"testing"
)

func TestWithVariablesReplacesWhateverTheCase(t *testing.T) {
	env := []string{"Path=C:\\bin", "vcpkg_root=old", "HOME=/h"}
	got := WithVariables(env, []string{"VCPKG_ROOT=D:\\vcpkg"})
	want := []string{"Path=C:\\bin", "HOME=/h", "VCPKG_ROOT=D:\\vcpkg"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if same := WithVariables(env, nil); !slices.Equal(same, env) {
		t.Errorf("no variables must change nothing, got %v", same)
	}
}
