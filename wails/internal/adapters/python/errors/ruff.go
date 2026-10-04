package errors

import (
	"encoding/json"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// ruffPosition is a row and column of the JSON of `ruff check --output-format json`.
type ruffPosition struct {
	Row    int `json:"row"`
	Column int `json:"column"`
}

// ruffEntry is one finding. Code is null for syntax errors.
type ruffEntry struct {
	Code        *string      `json:"code"`
	Message     string       `json:"message"`
	Filename    string       `json:"filename"`
	Location    ruffPosition `json:"location"`
	EndLocation ruffPosition `json:"end_location"`
}

// syntaxCode is what a syntax error of ruff (which has no code) is reported as.
const syntaxCode = "E999"

// ruffDiagnostics turns the JSON array of ruff into one diagnostic per finding. Findings are
// warnings (F401, F841, F821...) except syntax errors, which are errors. Unreadable JSON
// gives no diagnostics.
func (p parser) ruffDiagnostics(rawOutput string) []domain.Diagnostic {
	var entries []ruffEntry
	if err := json.Unmarshal([]byte(rawOutput), &entries); err != nil {
		return nil
	}
	diagnostics := make([]domain.Diagnostic, 0, len(entries))
	for _, entry := range entries {
		diagnostics = append(diagnostics, p.ruffDiagnostic(entry))
	}
	return diagnostics
}

func (p parser) ruffDiagnostic(entry ruffEntry) domain.Diagnostic {
	code, severity := syntaxCode, domain.SeverityError
	if entry.Code != nil {
		code, severity = *entry.Code, domain.SeverityWarning
	}
	file := errorcatalog.ResolvePath(entry.Filename, p.workingDir)
	message := strings.TrimSpace(entry.Message)
	diagnostic := domain.Diagnostic{
		Location: &domain.SourceLocation{File: file, Line: entry.Location.Row, Column: max(entry.Location.Column, 1)},
		Severity: severity,
		Message:  message,
		RawText:  message,
		Source:   sourceRuff,
		Code:     code,
	}
	if entry.EndLocation.Row > 0 {
		diagnostic.End = &domain.SourceLocation{File: file, Line: entry.EndLocation.Row, Column: max(entry.EndLocation.Column, 1)}
	}
	return diagnostic
}
