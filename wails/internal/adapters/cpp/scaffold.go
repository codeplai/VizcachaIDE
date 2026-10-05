package cpp

import "github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cmake"

// Scaffold is the "Hello + keyboard input" project of C++: a CMake project (CMakeLists.txt,
// CMakePresets.json, vcpkg.json and .gitignore, the same files cmake.Generate writes for an old
// folder), main.cpp, and a .clang-format with the style the IDE formats with (otherwise
// clang-format outside the IDE uses 2 spaces and reformats the code). clangd reads the flags from
// build/compile_commands.json, so there is no compile_flags.txt. It implements app.ProjectScaffold
// and touches no disk.
type Scaffold struct {
	// Language returns the language of the IDE, "es" or "en": the comments of CMakeLists.txt are
	// written in it. Nil means English.
	Language func() string
}

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

// Scaffold returns the CMake project of the folder name and main.cpp.
func (s Scaffold) Scaffold(name string) (map[string]string, string) {
	language := ""
	if s.Language != nil {
		language = s.Language()
	}
	files := cmake.Files(name, language)
	files["main.cpp"] = mainTemplate
	files[".clang-format"] = formatStyle
	return files, "main.cpp"
}
