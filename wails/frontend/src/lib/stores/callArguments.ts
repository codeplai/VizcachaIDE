// Delve's DAP server sends a function's arguments mixed with its locals (one "Locals"
// scope) plus return-value slots named "~r0". These helpers rebuild "factorial(n=5)"
// from the function's own source and hide the "~r" slots a beginner would not recognize.
import type { FrameVariables, Variable } from '../domain'

const escape = (text: string): string => text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

/** "main.(*Cuenta).Depositar" → "Depositar"; "main.factorial" → "factorial". */
const bareName = (functionName: string): string => functionName.split('.').pop() ?? functionName

/** Names of the parameters in `func name(a, b int, s string)`, read from the source text. */
export const parameterNames = (source: string, functionName: string): string[] => {
  const name = escape(bareName(functionName))
  const match = new RegExp(`func\\s*(?:\\([^)]*\\)\\s*)?${name}\\s*\\(([^)]*)\\)`).exec(source)
  if (!match?.[1]?.trim()) return []
  return match[1]
    .split(',')
    .map((part) => part.trim().split(/\s+/)[0] ?? '')
    .filter((part) => /^[\p{L}_][\p{L}\p{N}_]*$/u.test(part))
}

/** True for the debugger's return-value slots ("~r0", "~r1"). */
export const isHiddenVariable = (variable: Variable): boolean => variable.name.startsWith('~')

/** Moves the parameters out of the locals when the debugger did not separate them. */
export const splitArguments = (details: FrameVariables, parameters: string[]): FrameVariables => {
  const locals = details.locals.filter((variable) => !isHiddenVariable(variable))
  if (details.arguments.length > 0 || parameters.length === 0) return { ...details, locals }
  return {
    arguments: locals.filter((variable) => parameters.includes(variable.name)),
    locals: locals.filter((variable) => !parameters.includes(variable.name))
  }
}
