package bridge

import "github.com/codeplai/VizcachaIDE/wails/internal/domain"

// Sample data of the W0 stubs. It matches the approved prototype (docs/wails/prototype.html).
// The W1 tracks replace the stubs that use it with real adapters.

const (
	sampleDir  = "hola-go"
	sampleMain = "hola-go/main.go"
	sampleCalc = "hola-go/calculadora.go"
)

const sampleMainSource = `package main

import "fmt"

func sumar(a, b int) int {
    total := a + b
    return total
}

func main() {
    resultado := sumar(5, 7)
    fmt.Println("Hola, Go")
}
`

const sampleCalcSource = `package main

func restar(a, b int) int {
    return a - b
}
`

func sampleTree() domain.FileNode {
	return domain.FileNode{
		Name: sampleDir, Path: sampleDir, IsDir: true,
		Children: []domain.FileNode{
			{Name: "go.mod", Path: sampleDir + "/go.mod"},
			{Name: "main.go", Path: sampleMain},
			{Name: "calculadora.go", Path: sampleCalc},
		},
	}
}

func sampleLocation(line, column int) domain.SourceLocation {
	return domain.SourceLocation{File: sampleMain, Line: line, Column: column}
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
