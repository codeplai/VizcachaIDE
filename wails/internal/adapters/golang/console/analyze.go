package console

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"

	"github.com/traefik/yaegi/stdlib"
)

// isValueExpression reports whether the snippet is a single expression whose value
// should be shown. Statements, declarations and imports show nothing, and neither do
// the fmt.Print* calls (their count and error are noise for a learner).
func isValueExpression(code string) bool {
	expr, err := parser.ParseExpr(code)
	if err != nil {
		return false
	}
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return true
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return true
	}
	pkg, ok := selector.X.(*ast.Ident)
	return !ok || pkg.Name != "fmt" || !strings.HasPrefix(selector.Sel.Name, "Print")
}

// packagePaths maps a package name ("rand") to its import path, preferring the
// shortest path when several packages share the name ("math/rand").
var packagePaths = buildPackagePaths()

func buildPackagePaths() map[string]string {
	paths := map[string]string{}
	for key := range stdlib.Symbols {
		slash := strings.LastIndex(key, "/")
		if slash < 0 {
			continue
		}
		path, name := key[:slash], key[slash+1:]
		if current, taken := paths[name]; !taken || len(path) < len(current) {
			paths[name] = path
		}
	}
	return paths
}

// missingImports lists the standard packages the snippet uses (fmt.Println) without
// importing them, so a learner does not have to type the import line.
func missingImports(code string) []string {
	file := parseSnippet(code)
	if file == nil {
		return nil
	}
	found := map[string]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if pkg, ok := selector.X.(*ast.Ident); ok && pkg.Obj == nil {
			if path, known := packagePaths[pkg.Name]; known {
				found[path] = true
			}
		}
		return true
	})
	paths := make([]string, 0, len(found))
	for path := range found {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

// parseSnippet parses the code as statements, or as declarations when that fails.
func parseSnippet(code string) *ast.File {
	fset := token.NewFileSet()
	if file, err := parser.ParseFile(fset, "", "package p\nfunc _() {\n"+code+"\n}", 0); err == nil {
		return file
	}
	if file, err := parser.ParseFile(fset, "", "package p\n"+code, 0); err == nil {
		return file
	}
	return nil
}

// declaredImports lists the packages that the snippet imports itself.
func declaredImports(code string) []string {
	file := parseSnippet(code)
	if file == nil {
		return nil
	}
	paths := []string{}
	for _, spec := range file.Imports {
		paths = append(paths, strings.Trim(spec.Path.Value, `"`))
	}
	return paths
}
