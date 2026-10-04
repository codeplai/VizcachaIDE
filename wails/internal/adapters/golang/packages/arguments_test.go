package packages

import (
	"errors"
	"reflect"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

func TestGoCommandArgumentsAreValidated(t *testing.T) {
	if got, err := modInitArguments(" example.com/x "); err != nil || !reflect.DeepEqual(got, []string{"mod", "init", "example.com/x"}) {
		t.Errorf("ModInitArguments = %v, %v", got, err)
	}
	if got, err := getArguments("github.com/google/uuid"); err != nil || !reflect.DeepEqual(got, []string{"get", "github.com/google/uuid"}) {
		t.Errorf("GetArguments = %v, %v", got, err)
	}
	for _, bad := range []string{"", "  ", "-u", "two words"} {
		if _, err := getArguments(bad); !errors.Is(err, app.ErrInvalidArgument) {
			t.Errorf("getArguments(%q) error = %v", bad, err)
		}
		if _, err := modInitArguments(bad); !errors.Is(err, app.ErrInvalidArgument) {
			t.Errorf("modInitArguments(%q) error = %v", bad, err)
		}
	}
}
