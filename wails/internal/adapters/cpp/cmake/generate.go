package cmake

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Markers of the block VizcachaIDE edits when a library is added or removed. They are the same
// in every language: the library manager looks for them.
const (
	LibrariesBegin = "# VizcachaIDE libraries (vcpkg) · begin"
	LibrariesEnd   = "# VizcachaIDE libraries (vcpkg) · end"
)

const listsTemplate = `cmake_minimum_required(VERSION 3.25)
project(@TARGET@ LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

@COMMENT@
file(GLOB SOURCES CONFIGURE_DEPENDS *.cpp *.cc *.cxx)
add_executable(@TARGET@ ${SOURCES})
target_compile_options(@TARGET@ PRIVATE -Wall -Wextra)

` + LibrariesBegin + `
` + LibrariesEnd + `
`

const (
	sourcesCommentEN = "# Every .cpp in this folder is part of the program: a new file is enough."
	sourcesCommentES = "# Todo .cpp de esta carpeta es parte del programa: basta con crear un archivo nuevo."
)

const presetsFile = `{
  "version": 3,
  "configurePresets": [
    {
      "name": "debug",
      "displayName": "Debug",
      "generator": "Ninja",
      "binaryDir": "${sourceDir}/build",
      "cacheVariables": { "CMAKE_BUILD_TYPE": "Debug" },
      "toolchainFile": "$env{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake"
    }
  ],
  "buildPresets": [
    { "name": "debug", "configurePreset": "debug" }
  ]
}
`

// Names of the files of a project.
const (
	ManifestFile = "vcpkg.json"
	PresetsFile  = "CMakePresets.json"
	IgnoreFile   = ".gitignore"
)

// Files are the texts of a new project's files by name, for the folder name given. lang is the
// language of the comments in CMakeLists.txt: "es" for Spanish, English otherwise. The scaffold
// of a new project and Generate write the same texts.
func Files(folderName, lang string) map[string]string {
	target := TargetName(folderName)
	comment := sourcesCommentEN
	if lang == "es" {
		comment = sourcesCommentES
	}
	lists := strings.NewReplacer("@TARGET@", target, "@COMMENT@", comment).Replace(listsTemplate)
	return map[string]string{
		ListsFile:    lists,
		ManifestFile: fmt.Sprintf("{\n  \"name\": %q,\n  \"version\": %q,\n  \"dependencies\": []\n}\n", packageName(target), "0.1.0"),
		PresetsFile:  presetsFile,
		IgnoreFile:   "build/\n",
	}
}

// packageName is the target as a vcpkg name, which allows only lower case letters, digits and "-".
func packageName(target string) string { return strings.ReplaceAll(target, "_", "-") }

// Generate writes the files of a CMake project into folder, the ones that are missing: it never
// overwrites a file. It returns the project those files describe.
func Generate(folder, lang string) (Project, error) {
	files := Files(filepath.Base(folder), lang)
	for _, name := range []string{ListsFile, ManifestFile, PresetsFile, IgnoreFile} {
		if err := writeNew(filepath.Join(folder, name), files[name]); err != nil {
			return Project{}, err
		}
	}
	return Project{Root: folder, Target: targetOf(folder)}, nil
}

// writeNew creates the file with text unless it exists.
func writeNew(path, text string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("create %s: %w", filepath.Base(path), err)
	}
	if _, err := file.WriteString(text); err != nil {
		_ = file.Close()
		return fmt.Errorf("write %s: %w", filepath.Base(path), err)
	}
	return file.Close()
}
