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
