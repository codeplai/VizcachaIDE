package cpp

// Scaffold is the "Hello + keyboard input" project of C++: main.cpp and a compile_flags.txt so
// clangd uses the same flags as the compiler. It implements app.ProjectScaffold and touches no
// disk.
type Scaffold struct{}

const mainTemplate = `#include <iostream>
#include <string>

int main() {
    std::cout << "¿Cómo te llamas? ";
    std::string nombre;
    std::getline(std::cin, nombre);
    std::cout << "Hola, " << nombre << std::endl;
    return 0;
}
`

// Scaffold returns main.cpp and compile_flags.txt; the name does not change them.
func (Scaffold) Scaffold(string) (map[string]string, string) {
	return map[string]string{
		"main.cpp":          mainTemplate,
		"compile_flags.txt": "-std=c++17\n-Wall\n-Wextra\n",
	}, "main.cpp"
}
