package python

// Scaffold is the "Hello + keyboard input" project of Python: one main.py, and no virtual
// environment (the student can add one later). It implements app.ProjectScaffold and touches
// no disk.
type Scaffold struct{}

const mainTemplate = `nombre = input("¿Cómo te llamas? ")
print(f"Hola, {nombre}")
`

// Scaffold returns main.py; the name does not change it.
func (Scaffold) Scaffold(string) (map[string]string, string) {
	return map[string]string{"main.py": mainTemplate}, "main.py"
}
