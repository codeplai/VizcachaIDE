package packages

// Scaffold is the "Hello + keyboard input" project of Rust: Cargo.toml, src/main.rs and a
// .gitignore for target/. It implements app.ProjectScaffold and touches no disk.
type Scaffold struct{}

const mainTemplate = `use std::io::{self, Write};

fn main() {
    print!("¿Cómo te llamas? ");
    io::stdout().flush().unwrap();
    let mut nombre = String::new();
    io::stdin().read_line(&mut nombre).unwrap();
    println!("Hola, {}", nombre.trim());
}
`

// Scaffold returns the files of a crate named after the project (see CrateName).
func (Scaffold) Scaffold(name string) (map[string]string, string) {
	manifest := "[package]\nname = \"" + CrateName(name) + "\"\nversion = \"0.1.0\"\nedition = \"2021\"\n\n[dependencies]\n"
	return map[string]string{
		"Cargo.toml":  manifest,
		"src/main.rs": mainTemplate,
		".gitignore":  "/target\n",
	}, "src/main.rs"
}
