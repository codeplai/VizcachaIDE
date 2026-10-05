package cpp

// Scaffold is the "Hello + keyboard input" project of C++: main.cpp, a compile_flags.txt so
// clangd uses the same flags as the compiler, and a .clang-format with the style the IDE formats
// with (otherwise clang-format outside the IDE uses 2 spaces and reformats the code). It
// implements app.ProjectScaffold and touches no disk.
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

// formatStyle is the IDE's default C++ style (clangformat.defaultStyle) written down.
const formatStyle = "BasedOnStyle: LLVM\nIndentWidth: 4\n"

// Scaffold returns main.cpp, compile_flags.txt and .clang-format; the name does not change them.
func (Scaffold) Scaffold(string) (map[string]string, string) {
	return map[string]string{
		"main.cpp":          mainTemplate,
		"compile_flags.txt": "-std=c++17\n-Wall\n-Wextra\n",
		".clang-format":     formatStyle,
	}, "main.cpp"
}
