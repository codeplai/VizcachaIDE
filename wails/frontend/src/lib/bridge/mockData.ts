// Demo data of the mock bridge. It matches the approved prototype and the Go stubs
// (internal/bridge/sample_data.go).
import type {
  CodeLanguage,
  DebugState,
  Diagnostic,
  DocumentSymbol,
  ErrorExplanation,
  FileNode,
  Settings,
  SourceLocation,
  Variable
} from '../domain'
import { SAMPLE_CALLERS } from './mockFrames'

export const SAMPLE_DIR = 'hola-go'
export const SAMPLE_MAIN = 'hola-go/main.go'
export const SAMPLE_CALC = 'hola-go/calculadora.go'

export const SAMPLE_SOURCES: Record<string, string> = {
  [SAMPLE_MAIN]: `package main

import "fmt"

func sumar(a, b int) int {
    total := a + b
    return total
}

func main() {
    resultado := sumar(5, 7)
    fmt.Println("Hola, Go")
}

func factorial(n int) int {
    if n <= 1 {
        return sumar(1, 0)
    }
    return n * factorial(n-1)
}
`,
  [SAMPLE_CALC]: `package main

func restar(a, b int) int {
    return a - b
}
`
}

export const sampleLocation = (line: number, column: number): SourceLocation => ({
  file: SAMPLE_MAIN,
  line,
  column
})

export const sampleTree = (): FileNode => ({
  name: SAMPLE_DIR,
  path: SAMPLE_DIR,
  isDir: true,
  children: [
    { name: 'go.mod', path: `${SAMPLE_DIR}/go.mod`, isDir: false, children: [] },
    { name: 'main.go', path: SAMPLE_MAIN, isDir: false, children: [] },
    { name: 'calculadora.go', path: SAMPLE_CALC, isDir: false, children: [] }
  ]
})

export const sampleDiagnostic = (): Diagnostic => ({
  location: sampleLocation(11, 5),
  end: sampleLocation(11, 14),
  severity: 'error',
  message: 'declared and not used: resultado',
  rawText: './main.go:11:5: declared and not used: resultado',
  source: 'compiler',
  code: ''
})

const EXPLANATIONS: Record<'en' | 'es', ErrorExplanation> = {
  en: {
    explanationId: 'E-UNUSED-VAR',
    title: 'The variable “resultado” is never used',
    body: 'You created “resultado” on line 11, but you never read its value. Go refuses to compile unused variables because they are almost always a mistake or leftover code.',
    fixHint:
      "Print it with fmt.Println(resultado), delete it, or replace it with _ if you don't need it.",
    placeholders: { name: 'resultado', line: '11' }
  },
  es: {
    explanationId: 'E-UNUSED-VAR',
    title: 'La variable «resultado» nunca se usa',
    body: 'Creaste «resultado» en la línea 11, pero no lees su valor en ninguna parte. Go no compila código con variables sin usar, porque casi siempre es un error o un resto de código viejo.',
    fixHint: 'Imprímela con fmt.Println(resultado), bórrala o cámbiala por _ si no la necesitas.',
    placeholders: { name: 'resultado', line: '11' }
  }
}

export const sampleExplanation = (language: 'en' | 'es'): ErrorExplanation => EXPLANATIONS[language]

/** A struct whose fields load on demand (`debug:variables` answers `requestVariables`). */
export const SAMPLE_STRUCT_REFERENCE = 11

const sampleStructVariable: Variable = {
  name: 'punto',
  typeName: 'main.Punto',
  value: '{X: 3, Y: 4}',
  reference: SAMPLE_STRUCT_REFERENCE,
  changed: false,
  children: []
}

export const sampleStructFields = (): Variable[] => [
  { name: 'X', typeName: 'int', value: '3', reference: 0, changed: false, children: [] },
  { name: 'Y', typeName: 'int', value: '4', reference: 0, changed: true, children: [] }
]

export const sampleDebugState = (line = 6): DebugState => ({
  reason: line === 6 ? 'breakpoint' : 'step',
  frames: [{ frameId: 1, function: 'sumar', location: sampleLocation(line, 1) }, ...SAMPLE_CALLERS],
  variables: [
    { name: 'a', typeName: 'int', value: '5', reference: 0, changed: false, children: [] },
    { name: 'b', typeName: 'int', value: '7', reference: 0, changed: false, children: [] },
    {
      name: 'total',
      typeName: 'int',
      value: '12',
      reference: 0,
      changed: line === 6,
      children: []
    },
    ...(line > 6 ? [sampleStructVariable] : [])
  ],
  threads: [
    { threadId: 1, name: 'main.main', location: sampleLocation(11, 1) },
    { threadId: 2, name: 'runtime.gopark', location: null }
  ],
  currentThread: 1,
  description: ''
})

export const sampleSymbols = (): DocumentSymbol[] => [
  {
    name: 'sumar',
    kind: 'function',
    location: sampleLocation(5, 6),
    range: null,
    detail: 'func(a, b int) int',
    children: []
  },
  {
    name: 'main',
    kind: 'function',
    location: sampleLocation(10, 6),
    range: null,
    detail: 'func()',
    children: []
  }
]

export const defaultSettings = (defaultCodeLanguage: CodeLanguage = 'go'): Settings => ({
  language: 'auto',
  theme: 'system',
  fontSize: 14,
  toolPaths: {},
  defaultCodeLanguage,
  enabledCodeLanguages: [],
  firstRun: false,
  lastFolder: '',
  formatOnSave: true,
  recentFiles: [],
  checkUpdates: true,
  lastUpdateCheck: '',
  inlayHints: true
})
