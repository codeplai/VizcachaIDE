package errors

import (
	"encoding/json"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// cargoEnvelope is a line of `cargo --message-format=json`; rustc's own diagnostic has no reason
// (and its "message" is a string, so it is kept raw).
type cargoEnvelope struct {
	Reason  string          `json:"reason"`
	Message json.RawMessage `json:"message"`
}

type rustcSpan struct {
	FileName    string `json:"file_name"`
	LineStart   int    `json:"line_start"`
	LineEnd     int    `json:"line_end"`
	ColumnStart int    `json:"column_start"`
	ColumnEnd   int    `json:"column_end"`
	IsPrimary   bool   `json:"is_primary"`
}

type rustcDiagnostic struct {
	MessageType string `json:"$message_type"`
	Message     string `json:"message"`
	Code        *struct {
		Code string `json:"code"`
	} `json:"code"`
	Level    string      `json:"level"`
	Spans    []rustcSpan `json:"spans"`
	Rendered *string     `json:"rendered"`
}

// jsonDiagnostic reads one JSON line: a rustc diagnostic, or the compiler-message of cargo.
// Other lines (artifacts, build-finished, the summaries at the end) give nil.
func (r *reader) jsonDiagnostic(line string) *domain.Diagnostic {
	var envelope cargoEnvelope
	if json.Unmarshal([]byte(line), &envelope) != nil {
		return nil
	}
	payload := []byte(line)
	switch envelope.Reason {
	case "":
	case "compiler-message":
		payload = envelope.Message
	default:
		return nil
	}
	var raw rustcDiagnostic
	if json.Unmarshal(payload, &raw) != nil || (raw.MessageType != "" && raw.MessageType != "diagnostic") {
		return nil
	}
	severity, ok := severityOf(raw.Level)
	if !ok || (len(raw.Spans) == 0 && summaryMessage.MatchString(raw.Message)) {
		return nil
	}
	code := ""
	if raw.Code != nil {
		code = raw.Code.Code
	}
	rawText := raw.Message
	if raw.Rendered != nil {
		rawText = strings.TrimRight(*raw.Rendered, "\r\n")
	}
	diagnostic := &domain.Diagnostic{
		Severity: severity, Message: raw.Message, RawText: rawText,
		Source: sourceOf(code), Code: r.codeOrID(code, raw.Message),
	}
	if span, found := primarySpan(raw.Spans); found {
		diagnostic.Location = spanLocation(span.FileName, span.LineStart, span.ColumnStart, r.workingDir)
		diagnostic.End = spanLocation(span.FileName, span.LineEnd, span.ColumnEnd, r.workingDir)
	}
	return diagnostic
}

func primarySpan(spans []rustcSpan) (rustcSpan, bool) {
	for _, span := range spans {
		if span.IsPrimary {
			return span, true
		}
	}
	if len(spans) > 0 {
		return spans[0], true
	}
	return rustcSpan{}, false
}

func spanLocation(file string, line, column int, workingDir string) *domain.SourceLocation {
	return &domain.SourceLocation{File: errorcatalog.ResolvePath(file, workingDir), Line: line, Column: column}
}
