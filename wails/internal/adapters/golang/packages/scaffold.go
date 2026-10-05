package packages

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Scaffold is the "Hello + keyboard input" project of Go: a module and a main.go that asks for
// the student's name. It implements app.ProjectScaffold and touches no disk.
type Scaffold struct{}

const mainTemplate = `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Print("¿Cómo te llamas? ")
	nombre, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	fmt.Printf("Hola, %s\n", strings.TrimSpace(nombre))
}
`

// Scaffold returns go.mod and main.go for a project called name.
func (Scaffold) Scaffold(name string) (map[string]string, string) {
	return map[string]string{
		"go.mod":  "module " + ModuleName(name) + "\n\ngo 1.21\n",
		"main.go": mainTemplate,
	}, "main.go"
}

// ModuleName turns a project name into a module path: lower case, no accents, spaces as "-"
// and only a-z, 0-9, ".", "_" and "-" ("Mi Programa Ñandú" gives "mi-programa-nandu").
func ModuleName(name string) string {
	plain, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC), name)
	if err != nil {
		plain = name
	}
	var module strings.Builder
	for _, letter := range strings.ToLower(plain) {
		switch {
		case letter == ' ':
			module.WriteRune('-')
		case letter < unicode.MaxASCII && (unicode.IsLower(letter) || unicode.IsDigit(letter) || strings.ContainsRune("._-", letter)):
			module.WriteRune(letter)
		}
	}
	clean := strings.Trim(module.String(), "-._")
	if clean == "" {
		return "proyecto"
	}
	return clean
}
