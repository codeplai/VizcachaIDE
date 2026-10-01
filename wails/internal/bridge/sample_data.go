package bridge

import "github.com/codeplai/VizcachaIDE/wails/internal/domain"

// Sample data of the W0 stubs. It matches the approved prototype (docs/wails/prototype.html).
// The W1 tracks replace the stubs that use it with real adapters.

const (
	sampleDir  = "hola-go"
	sampleMain = "hola-go/main.go"
	sampleCalc = "hola-go/calculadora.go"
)

func sampleLocation(line, column int) domain.SourceLocation {
	return domain.SourceLocation{File: sampleMain, Line: line, Column: column}
}

func sampleDiagnostic() domain.Diagnostic {
	loc, end := sampleLocation(11, 5), sampleLocation(11, 14)
	return domain.Diagnostic{
		Location: &loc,
		End:      &end,
		Severity: domain.SeverityError,
		Message:  "declared and not used: resultado",
		RawText:  "./main.go:11:5: declared and not used: resultado",
		Source:   "go",
	}
}

func sampleExplanation(language string) *domain.ErrorExplanation {
	placeholders := map[string]string{"name": "resultado", "line": "11"}
	if language == domain.LanguageES {
		return &domain.ErrorExplanation{
			ExplanationID: "E-UNUSED-VAR",
			Title:         "La variable «resultado» nunca se usa",
			Body:          "Creaste «resultado» en la línea 11, pero no lees su valor en ninguna parte. Go no compila código con variables sin usar, porque casi siempre es un error o un resto de código viejo.",
			FixHint:       "Imprímela con fmt.Println(resultado), bórrala o cámbiala por _ si no la necesitas.",
			Placeholders:  placeholders,
		}
	}
	return &domain.ErrorExplanation{
		ExplanationID: "E-UNUSED-VAR",
		Title:         "The variable “resultado” is never used",
		Body:          "You created “resultado” on line 11, but you never read its value. Go refuses to compile unused variables because they are almost always a mistake or leftover code.",
		FixHint:       "Print it with fmt.Println(resultado), delete it, or replace it with _ if you don't need it.",
		Placeholders:  placeholders,
	}
}

func sampleDebugState() domain.DebugState {
	top, caller := sampleLocation(6, 1), sampleLocation(11, 1)
	goroutine := 1
	return domain.DebugState{
		Reason: domain.StopBreakpoint,
		Frames: []domain.StackFrame{
			{FrameID: 1, Function: "sumar", Location: &top},
			{FrameID: 2, Function: "main", Location: &caller},
		},
		Variables: []domain.Variable{
			{Name: "a", TypeName: "int", Value: "5"},
			{Name: "b", TypeName: "int", Value: "7"},
			{Name: "total", TypeName: "int", Value: "12", Changed: true},
		},
		Goroutines:       []domain.Goroutine{{GoroutineID: 1, Name: "main.main", Location: &caller}},
		CurrentGoroutine: &goroutine,
	}
}

func sampleSymbols() []domain.DocumentSymbol {
	return []domain.DocumentSymbol{
		{Name: "sumar", Kind: domain.SymbolFunction, Location: sampleLocation(5, 6), Detail: "func(a, b int) int"},
		{Name: "main", Kind: domain.SymbolFunction, Location: sampleLocation(10, 6), Detail: "func()"},
	}
}
