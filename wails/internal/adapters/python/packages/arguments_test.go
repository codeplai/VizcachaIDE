package packages

import (
	"slices"
	"testing"
)

func TestEveryPipCommandSkipsTheVersionCheck(t *testing.T) {
	prefix := []string{"-m", "pip", "--disable-pip-version-check"}
	cases := map[string][]string{
		"install":   pipArguments("install", "requests"),
		"uninstall": pipArguments("uninstall", "-y", "requests"),
		"list":      pipArguments("list"),
	}
	for name, args := range cases {
		if !slices.Equal(args[:3], prefix) || args[3] != name {
			t.Errorf("%s = %v", name, args)
		}
	}
}
