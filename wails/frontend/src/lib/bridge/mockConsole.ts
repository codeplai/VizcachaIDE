import type { ConsoleResult } from '../domain'
import type { ConsoleApi } from './types'

const ARITHMETIC = /^[\d\s+\-*/%().]+$/
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

/** A tiny stand-in for the Go console: arithmetic, `x := 5` and a canned answer. */
export const mockConsole = (): ConsoleApi => {
  const variables = new Map<string, string>()
  return {
    eval: async (_codeLanguage, code) => {
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
      if (/^fmt\.Println\(/.test(text)) return { result: '', output: 'Hello!\n', error: '' }
      return failure(`undefined: ${text.split(/\W+/)[0]}`)
    },
    reset: async () => variables.clear()
  }
}
