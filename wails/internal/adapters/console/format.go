package console

import (
	"fmt"
	"reflect"
)

// format prints a value like a REPL: strings quoted, the rest as %v.
// It returns "" for an invalid value (a statement) and for nil.
func format(value reflect.Value) string {
	if !value.IsValid() {
		return ""
	}
	if value.Kind() == reflect.Interface && value.IsNil() {
		return ""
	}
	if !value.CanInterface() {
		return fmt.Sprint(value)
	}
	v := value.Interface()
	if _, ok := v.(error); ok {
		return fmt.Sprintf("%v", v)
	}
	if _, ok := v.(fmt.Stringer); ok {
		return fmt.Sprintf("%v", v)
	}
	if value.Kind() == reflect.String {
		return fmt.Sprintf("%q", value.String())
	}
	return fmt.Sprintf("%v", v)
}
