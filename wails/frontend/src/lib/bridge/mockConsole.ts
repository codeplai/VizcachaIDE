import type { ConsoleResult } from '../domain'
import type { ConsoleApi } from './types'

const ARITHMETIC = /^[\d\s+\-*/%().]+$/
const PYTHON_PRINT = /^print\((["'])(.*)\1\)$/
const PYTHON_INPUT = /\binput\(/
/** errors.consoleNoInput in English: the real backend answers in the interface language. */
const NO_INPUT = "The console can't read the keyboard. Try it in a file with F5."
const DECLARATION = /^([A-Za-z_]\w*)\s*:?=\s*(.+)$/

const failure = (error: string): ConsoleResult => ({ result: '', output: '', error })

/** Evaluates plain arithmetic after replacing the variables the demo remembers. */
const arithmetic = (code: string, variables: Map<string, string>): string | null => {
  const replaced = code.replace(/[A-Za-z_]\w*/g, (name) => variables.get(name) ?? name)
  if (!ARITHMETIC.test(replaced)) return null
  try {
    return String(new Function(`return (${replaced})`)())
  } catch {
    return null
  }
}

/** What the Python console says that Go's does not: print(), input() and NameError. */
const pythonAnswer = (text: string): ConsoleResult => {
  if (PYTHON_INPUT.test(text)) return failure(NO_INPUT)
  const printed = PYTHON_PRINT.exec(text)
  if (printed) return { result: '', output: `${printed[2]}\n`, error: '' }
  return failure(`NameError: name '${text.split(/\W+/)[0]}' is not defined`)
}

/** A tiny stand-in for the Go and Python consoles: arithmetic, `x := 5` or `x = 5`, a canned answer. */
export const mockConsole = (): ConsoleApi => {
  const variables = new Map<string, string>()
  return {
    eval: async (codeLanguage, code) => {
      const text = code.trim()
      const declaration = DECLARATION.exec(text)
      if (declaration) {
        const [, name = '', expression = ''] = declaration
        const value = arithmetic(expression, variables)
        if (value === null) return failure(`cannot evaluate ${expression} in the demo`)
        variables.set(name, value)
        return { result: '', output: '', error: '' }
      }
      const value = arithmetic(text, variables)
      if (value !== null) return { result: value, output: '', error: '' }
      if (codeLanguage === 'python') return pythonAnswer(text)
      if (/^fmt\.Println\(/.test(text)) return { result: '', output: 'Hello!\n', error: '' }
      return failure(`undefined: ${text.split(/\W+/)[0]}`)
    },
    reset: async () => variables.clear()
  }
}
