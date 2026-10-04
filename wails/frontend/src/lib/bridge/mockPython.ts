// Demo data of the Python sample project (`?language=python` in the browser): the same three
// states as the Go one (run, error, debug), with a factorial(n) to look at while debugging.
import type {
  DebugState,
  Diagnostic,
  ErrorExplanation,
  FileNode,
  FrameVariables,
  SourceLocation,
  StackFrame,
  Variable
} from '../domain'

export const PYTHON_DIR = 'hola-py'
export const PYTHON_MAIN = 'hola-py/main.py'
export const PYTHON_HELPER = 'hola-py/calculadora.py'
/** The line the dev bar puts a breakpoint on: the base case of factorial. */
export const PYTHON_BREAKPOINT_LINE = 8
export const PYTHON_LAST_LINE = 13

export const PYTHON_SOURCES: Record<string, string> = {
  [PYTHON_MAIN]: `def sumar(a, b):
    total = a + b
    return total


def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)


resultado = factorial(3)
print("Hola, Python")
`,
  [PYTHON_HELPER]: `def restar(a, b):
    return a - b
`
}

export const pythonTree = (): FileNode => ({
  name: PYTHON_DIR,
  path: PYTHON_DIR,
  isDir: true,
  children: [
    { name: 'main.py', path: PYTHON_MAIN, isDir: false, children: [] },
    { name: 'calculadora.py', path: PYTHON_HELPER, isDir: false, children: [] }
  ]
})

const at = (line: number): SourceLocation => ({ file: PYTHON_MAIN, line, column: 1 })

/** The NameError of the error scenario, as ruff/pylsp and the traceback report it. */
export const pythonDiagnostic = (): Diagnostic => ({
  location: { file: PYTHON_MAIN, line: 13, column: 7 },
  end: { file: PYTHON_MAIN, line: 13, column: 16 },
  severity: 'error',
  message: "name 'resultad' is not defined",
  rawText: `  File "${PYTHON_MAIN}", line 13, in <module>\nNameError: name 'resultad' is not defined`,
  source: 'pyflakes',
  code: 'F821'
})

const EXPLANATIONS: Record<'en' | 'es', ErrorExplanation> = {
  en: {
    explanationId: 'E-PY-NAME-ERROR',
    title: 'Python doesn’t know the name “resultad”',
    body: 'On line 13 you used “resultad”, but no variable or function has that name yet. Python only knows the names you created before using them, and it is picky about every letter.',
    fixHint:
      'Check the spelling: you may have meant “resultado”. Or create the variable above that line.',
    placeholders: { name: 'resultad', line: '13' }
  },
  es: {
    explanationId: 'E-PY-NAME-ERROR',
    title: 'Python no conoce el nombre «resultad»',
    body: 'En la línea 13 usaste «resultad», pero ninguna variable o función se llama así todavía. Python solo conoce los nombres que creaste antes de usarlos, y es exigente con cada letra.',
    fixHint:
      'Revisa cómo está escrito: quizá querías «resultado». O crea la variable antes de esa línea.',
    placeholders: { name: 'resultad', line: '13' }
  }
}

export const pythonExplanation = (language: 'en' | 'es'): ErrorExplanation => EXPLANATIONS[language]

const intVariable = (name: string, value: string, changed = false): Variable => ({
  name,
  typeName: 'int',
  value,
  reference: 0,
  changed,
  children: []
})

/** factorial(3) paused at its base case: three calls deep plus the module. */
const CALLERS: StackFrame[] = [
  { frameId: 2, function: 'factorial', location: at(9) },
  { frameId: 3, function: 'factorial', location: at(9) },
  { frameId: 4, function: '<module>', location: at(12) }
]

export const pythonDebugState = (line = PYTHON_BREAKPOINT_LINE): DebugState => ({
  reason: line === PYTHON_BREAKPOINT_LINE ? 'breakpoint' : 'step',
  frames: [{ frameId: 1, function: 'factorial', location: at(line) }, ...CALLERS],
  variables: [intVariable('n', '1', line === PYTHON_BREAKPOINT_LINE)],
  threads: [{ threadId: 1, name: 'MainThread', location: at(line) }],
  currentThread: 1,
  description: ''
})

/** Arguments and locals of each frame of the sample stack (Python has no separate arguments). */
export const pythonFrameVariables = (frameId: number): FrameVariables => {
  const byFrame: Record<number, FrameVariables> = {
    1: { arguments: [intVariable('n', '1')], locals: [] },
    2: { arguments: [intVariable('n', '2')], locals: [] },
    3: { arguments: [intVariable('n', '3')], locals: [] },
    4: { arguments: [], locals: [] }
  }
  return byFrame[frameId] ?? { arguments: [], locals: [] }
}
