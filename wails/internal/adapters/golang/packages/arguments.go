package packages

import "github.com/codeplai/VizcachaIDE/wails/internal/app"

// modInitArguments returns the arguments of "go mod init <modulePath>".
func modInitArguments(modulePath string) ([]string, error) {
	name, err := app.SingleWordArgument(modulePath, "module path")
	if err != nil {
		return nil, err
	}
	return []string{"mod", "init", name}, nil
}

// getArguments returns the arguments of "go get <pkg>".
func getArguments(pkg string) ([]string, error) {
	name, err := app.SingleWordArgument(pkg, "package")
	if err != nil {
		return nil, err
	}
	return []string{"get", name}, nil
}

// modTidyArguments returns the arguments of "go mod tidy".
func modTidyArguments() []string { return []string{"mod", "tidy"} }
