// Demo data of the C++ sample project (`?language=cpp` in the browser): the states of the other
// languages plus a crash, with a factorial(int n) to look at while debugging.
import type {
  DebugState,
  Diagnostic,
  ErrorExplanation,
  ExplainedDiagnostic,
  FileNode,
  FrameVariables,
  SourceLocation,
  StackFrame,
  Variable
} from '../domain'

export const CPP_DIR = 'hola-cpp'
export const CPP_MAIN = 'hola-cpp/main.cpp'
export const CPP_HELPER = 'hola-cpp/calculadora.h'
/** The line the dev bar puts a breakpoint on: the base case of factorial. */
export const CPP_BREAKPOINT_LINE = 11
export const CPP_LAST_LINE = 19
/** What the compiler stage prints before the program runs (the backend's own line). */
export const CPP_COMPILING_NOTICE = 'Compiling…'

export const CPP_SOURCES: Record<string, string> = {
  [CPP_MAIN]: `#include <iostream>
using namespace std;

int sumar(int a, int b) {
    int total = a + b;
    return total;
}

int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

int main() {
    int resultado = factorial(3);
    cout << "Hola, C++" << endl;
    return 0;
}
`,
  [CPP_HELPER]: `int restar(int a, int b) {
    return a - b;
}
`
}

export const cppTree = (): FileNode => ({
  name: CPP_DIR,
  path: CPP_DIR,
  isDir: true,
  children: [
    { name: 'calculadora.h', path: CPP_HELPER, isDir: false, children: [] },
    { name: 'main.cpp', path: CPP_MAIN, isDir: false, children: [] }
  ]
})

const at = (line: number): SourceLocation => ({ file: CPP_MAIN, line, column: 1 })

/** The undeclared name of the error scenario, as g++ reports it. */
export const cppDiagnostic = (): Diagnostic => ({
  location: { file: CPP_MAIN, line: 18, column: 13 },
  end: { file: CPP_MAIN, line: 18, column: 21 },
  severity: 'error',
  message: "'resultad' was not declared in this scope; did you mean 'resultado'?",
  rawText: `${CPP_MAIN}: In function 'int main()':\n${CPP_MAIN}:18:13: error: 'resultad' was not declared in this scope; did you mean 'resultado'?\n   18 |     cout << resultad << endl;\n      |             ^~~~~~~~\n      |             resultado`,
  source: 'g++',
  code: 'CPP-UNDECLARED'
})

/** The line the runner prints when the program dies on an invalid memory access. */
export const cppCrashDiagnostic = (): Diagnostic => ({
  location: null,
  end: null,
  severity: 'error',
  message: 'Segmentation fault',
  rawText: 'Segmentation fault',
  source: 'runtime',
  code: 'CPP-SEGFAULT'
})

type Texts = Record<'en' | 'es', ErrorExplanation>

const UNDECLARED: Texts = {
  en: {
    explanationId: 'CPP-UNDECLARED',
    title: 'C++ doesn’t know the name “resultad”',
    body: 'On line 18 you used “resultad”, but nothing with that name was declared before. C++ needs every variable and function to be declared above the place where you use it, and it checks every letter.',
    fixHint:
      'Check the spelling: you may have meant “resultado”. Or declare it above that line, for example int resultad = 0;',
    placeholders: { name: 'resultad', line: '18' }
  },
  es: {
    explanationId: 'CPP-UNDECLARED',
    title: 'C++ no conoce el nombre «resultad»',
    body: 'En la línea 18 usaste «resultad», pero nada con ese nombre fue declarado antes. C++ necesita que cada variable y función esté declarada arriba del lugar donde la usas, y revisa cada letra.',
    fixHint:
      'Revisa cómo está escrito: quizá querías «resultado». O decláralo antes de esa línea, por ejemplo int resultad = 0;',
    placeholders: { name: 'resultad', line: '18' }
  }
}

const SEGFAULT: Texts = {
  en: {
    explanationId: 'CPP-SEGFAULT',
    title: 'The program touched memory it doesn’t own',
    body: 'The operating system stopped your program (“Segmentation fault”) because it read or wrote a place in memory that was not yours. It usually happens with a pointer that points nowhere, or an array position past its end.',
    fixHint:
      'Debug it with F6: the program stops on the line that crashes. Check the pointers and the array positions there.',
    placeholders: {}
  },
  es: {
    explanationId: 'CPP-SEGFAULT',
    title: 'El programa tocó memoria que no es suya',
    body: 'El sistema operativo detuvo tu programa («Segmentation fault») porque leyó o escribió un lugar de la memoria que no era tuyo. Suele pasar con un puntero que no apunta a nada, o con una posición de un arreglo más allá de su final.',
    fixHint:
      'Depúralo con F6: el programa se detiene en la línea que se cae. Revisa ahí los punteros y las posiciones del arreglo.',
    placeholders: {}
  }
}

export const cppExplanation = (language: 'en' | 'es'): ErrorExplanation => UNDECLARED[language]

export const cppCrashExplained = (language: 'en' | 'es'): ExplainedDiagnostic[] => [
  { diagnostic: cppCrashDiagnostic(), explanation: SEGFAULT[language] }
]

const intVariable = (name: string, value: string, changed = false): Variable => ({
  name,
  typeName: 'int',
  value,
  reference: 0,
  changed,
  children: []
})

/** factorial(3) paused at its base case: three calls deep plus main. */
const CALLERS: StackFrame[] = [
  { frameId: 2, function: 'factorial(int)', location: at(13) },
  { frameId: 3, function: 'factorial(int)', location: at(13) },
  { frameId: 4, function: 'main', location: at(17) }
]

export const cppDebugState = (line = CPP_BREAKPOINT_LINE): DebugState => ({
  reason: line === CPP_BREAKPOINT_LINE ? 'breakpoint' : 'step',
  frames: [{ frameId: 1, function: 'factorial(int)', location: at(line) }, ...CALLERS],
  variables: [intVariable('n', '1', line === CPP_BREAKPOINT_LINE)],
  threads: [{ threadId: 1, name: 'main', location: at(line) }],
  currentThread: 1,
  description: ''
})

/** Arguments and locals of each frame of the sample stack. */
export const cppFrameVariables = (frameId: number): FrameVariables => {
  const byFrame: Record<number, FrameVariables> = {
    1: { arguments: [intVariable('n', '1')], locals: [] },
    2: { arguments: [intVariable('n', '2')], locals: [] },
    3: { arguments: [intVariable('n', '3')], locals: [] },
    4: { arguments: [], locals: [intVariable('resultado', '21845')] }
  }
  return byFrame[frameId] ?? { arguments: [], locals: [] }
}
