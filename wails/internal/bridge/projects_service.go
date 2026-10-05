package bridge

import (
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ProjectsService creates the starting project of a language (File > New project). Languages
// without a scaffold answer app.ErrProjectNoScaffold.
type ProjectsService struct {
	supportRouter
}

// NewProjectsService creates the service.
func NewProjectsService(registry *app.LanguageRegistry) *ProjectsService {
	return &ProjectsService{supportRouter{registry: registry}}
}

// Create makes the folder <location>/<name>, writes the language's template in it and returns
// the folder and the file to open. Its errors start with the i18n key of their text
// (project.errorNameEmpty...).
func (s *ProjectsService) Create(codeLanguage domain.CodeLanguage, location, name string) (domain.NewProject, error) {
	support, err := s.supportOf(codeLanguage)
	if err != nil {
		return domain.NewProject{}, fmt.Errorf("new project: %w", err)
	}
	project, err := app.CreateProject(support.Scaffold, location, name)
	if err != nil {
		return domain.NewProject{}, fmt.Errorf("%w", err)
	}
	return project, nil
}
