package rust

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// CargoProject is the Cargo crate around a file (docs/PLAN_RUST.md sections 4.3 and 3.2 point
// 8). In a workspace the crate is one member: cargo runs it with "-p <Name>" from the workspace,
// whose target/ is shared.
type CargoProject struct {
	Root     string // folder of the crate's Cargo.toml (the workspace's, for a virtual manifest)
	Manifest string // that Cargo.toml
	Name     string // [package] name ("" for a virtual manifest)
	Edition  string // [package] edition ("2015" when absent, as cargo does)
	// Workspace is the folder of the workspace's Cargo.toml when the crate is a member, or the
	// crate itself when it is not in a workspace.
	Workspace string
	// Virtual is true when the file is under a workspace root with no crate of its own:
	// Members lists the crates the user can choose to run.
	Virtual    bool
	Members    []string // folders of the members ([workspace] members, globs expanded)
	Binaries   []Binary // what "cargo run --bin" accepts; empty for a library
	DefaultRun string   // [package] default-run
}

// Binary is one program of a crate.
type Binary struct {
	Name string
	Path string // its main source file
}

// Context is the domain view of the project for RunConfiguration.Project.
func (p CargoProject) Context() *domain.ProjectContext {
	return &domain.ProjectContext{Root: p.Root, Kind: domain.ProjectCargo, Name: p.Name}
}

// BinaryFor is the program a file belongs to: src/bin/<name>.rs or src/bin/<name>/ is <name>;
// anything else is default-run, the binary named like the package, or the only one. ok is false
// for a library.
func (p CargoProject) BinaryFor(file string) (Binary, bool) {
	if relative, err := filepath.Rel(filepath.Join(p.Root, "src", "bin"), file); err == nil && !strings.HasPrefix(relative, "..") {
		first := strings.Split(filepath.ToSlash(relative), "/")[0]
		for _, binary := range p.Binaries {
			if binary.Name == strings.TrimSuffix(first, ".rs") {
				return binary, true
			}
		}
	}
	for _, wanted := range []string{p.DefaultRun, p.Name} {
		for _, binary := range p.Binaries {
			if wanted != "" && binary.Name == wanted {
				return binary, true
			}
		}
	}
	if len(p.Binaries) > 0 {
		return p.Binaries[0], true
	}
	return Binary{}, false
}

// manifest is the part of Cargo.toml the IDE reads.
type manifest struct {
	Package *struct {
		Name       string `toml:"name"`
		Edition    string `toml:"edition"`
		DefaultRun string `toml:"default-run"`
	} `toml:"package"`
	Workspace *struct {
		Members []string `toml:"members"`
	} `toml:"workspace"`
	Bin []struct {
		Name string `toml:"name"`
		Path string `toml:"path"`
	} `toml:"bin"`
}

// FindProject looks for Cargo.toml from the folder of path upwards. ok is false for a loose .rs
// file (no Cargo.toml above it). An unreadable Cargo.toml is an error.
func FindProject(path string) (project CargoProject, ok bool, err error) {
	for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
		file := filepath.Join(dir, "Cargo.toml")
		if _, statErr := os.Stat(file); statErr == nil {
			found, readErr := readProject(dir)
			return found, readErr == nil, readErr
		}
		if filepath.Dir(dir) == dir {
			return CargoProject{}, false, nil
		}
	}
}

func readProject(dir string) (CargoProject, error) {
	parsed, err := readManifest(dir)
	if err != nil {
		return CargoProject{}, err
	}
	project := CargoProject{Root: dir, Manifest: filepath.Join(dir, "Cargo.toml"), Workspace: dir}
	if parsed.Workspace != nil {
		project.Members = expandMembers(dir, parsed.Workspace.Members)
	}
	if parsed.Package == nil {
		project.Virtual = parsed.Workspace != nil
		return project, nil
	}
	project.Name, project.Edition, project.DefaultRun = parsed.Package.Name, parsed.Package.Edition, parsed.Package.DefaultRun
	if project.Edition == "" {
		project.Edition = "2015"
	}
	project.Binaries = binariesOf(dir, project.Name, parsed)
	if root, found := workspaceAbove(dir); found {
		project.Workspace = root
	}
	return project, nil
}

func readManifest(dir string) (manifest, error) {
	var parsed manifest
	_, err := toml.DecodeFile(filepath.Join(dir, "Cargo.toml"), &parsed)
	return parsed, err
}

// workspaceAbove finds the workspace that lists dir among its members.
func workspaceAbove(dir string) (string, bool) {
	for parent := filepath.Dir(dir); parent != filepath.Dir(parent); parent = filepath.Dir(parent) {
		parsed, err := readManifest(parent)
		if err != nil || parsed.Workspace == nil {
			continue
		}
		for _, member := range expandMembers(parent, parsed.Workspace.Members) {
			if sameFolder(member, dir) {
				return parent, true
			}
		}
	}
	return "", false
}

// expandMembers resolves the member globs ("crates/*") to the folders that have a Cargo.toml.
func expandMembers(root string, members []string) []string {
	folders := []string{}
	for _, member := range members {
		matches, _ := filepath.Glob(filepath.Join(root, filepath.FromSlash(member)))
		for _, folder := range matches {
			if _, err := os.Stat(filepath.Join(folder, "Cargo.toml")); err == nil {
				folders = append(folders, folder)
			}
		}
	}
	return folders
}

// binariesOf lists the programs of a crate: [[bin]] entries, src/main.rs (named like the
// package), src/bin/<name>.rs and src/bin/<name>/main.rs, as cargo discovers them.
func binariesOf(dir, packageName string, parsed manifest) []Binary {
	binaries := []Binary{}
	seen := map[string]bool{}
	add := func(name, path string) {
		if name != "" && !seen[name] {
			seen[name] = true
			binaries = append(binaries, Binary{Name: name, Path: path})
		}
	}
	for _, bin := range parsed.Bin {
		path := filepath.Join(dir, filepath.FromSlash(bin.Path))
		if bin.Path == "" {
			path = filepath.Join(dir, "src", "bin", bin.Name+".rs")
		}
		add(bin.Name, path)
	}
	if main := filepath.Join(dir, "src", "main.rs"); isFile(main) {
		add(packageName, main)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "src", "bin"))
	for _, entry := range entries {
		full := filepath.Join(dir, "src", "bin", entry.Name())
		switch {
		case !entry.IsDir() && strings.HasSuffix(entry.Name(), ".rs"):
			add(strings.TrimSuffix(entry.Name(), ".rs"), full)
		case entry.IsDir() && isFile(filepath.Join(full, "main.rs")):
			add(entry.Name(), filepath.Join(full, "main.rs"))
		}
	}
	return binaries
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func sameFolder(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}
