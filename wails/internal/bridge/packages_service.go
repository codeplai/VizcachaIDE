package bridge

import (
	"context"
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// PackagesService runs the package commands of a language (go mod init, go get, go mod tidy,
// pip install...). They run in the shared process slot and emit the run events. A language
// without a package manager, or without the verb, answers app.ErrUnsupported.
type PackagesService struct {
	supportRouter
}

// NewPackagesService creates the service.
func NewPackagesService(registry *app.LanguageRegistry) *PackagesService {
	return &PackagesService{supportRouter{registry: registry}}
}

// Init creates the project of a language in dir (go mod init <name>).
func (s *PackagesService) Init(codeLanguage domain.CodeLanguage, dir, name string) error {
	return s.run(codeLanguage, "init", func(p app.PackageManager) error {
		return p.Init(context.Background(), dir, name)
	})
}

// Add installs a package (go get <pkg>).
func (s *PackagesService) Add(codeLanguage domain.CodeLanguage, dir, pkg string) error {
	return s.run(codeLanguage, "add", func(p app.PackageManager) error {
		return p.Add(context.Background(), dir, pkg)
	})
}

// Remove uninstalls a package.
func (s *PackagesService) Remove(codeLanguage domain.CodeLanguage, dir, pkg string) error {
	return s.run(codeLanguage, "remove", func(p app.PackageManager) error {
		return p.Remove(context.Background(), dir, pkg)
	})
}

// Tidy cleans up the dependencies (go mod tidy).
func (s *PackagesService) Tidy(codeLanguage domain.CodeLanguage, dir string) error {
	return s.run(codeLanguage, "tidy", func(p app.PackageManager) error {
		return p.Tidy(context.Background(), dir)
	})
}

// List shows the installed packages.
func (s *PackagesService) List(codeLanguage domain.CodeLanguage, dir string) error {
	return s.run(codeLanguage, "list", func(p app.PackageManager) error {
		return p.List(context.Background(), dir)
	})
}

func (s *PackagesService) run(codeLanguage domain.CodeLanguage, verb string, call func(app.PackageManager) error) error {
	support, err := s.supportOf(codeLanguage)
	if err != nil {
		return fmt.Errorf("packages %s: %w", verb, err)
	}
	if support.Packages == nil {
		return fmt.Errorf("packages %s %s: %w", codeLanguage, verb, app.ErrUnsupported)
	}
	if err := call(support.Packages); err != nil {
		return fmt.Errorf("packages %s %s: %w", codeLanguage, verb, err)
	}
	return nil
}
