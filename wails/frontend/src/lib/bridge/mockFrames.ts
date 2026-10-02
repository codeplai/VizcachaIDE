// The frames below the top one of the debug demo, and what each frame holds.
import type { FrameVariables, SourceLocation, StackFrame, Variable } from '../domain'

const at = (line: number): SourceLocation => ({ file: 'hola-go/main.go', line, column: 1 })

/** factorial(3) paused three calls deep: the recursion the Calls view draws as boxes. */
export const SAMPLE_CALLERS: StackFrame[] = [
  { frameId: 2, function: 'factorial', location: at(16) },
  { frameId: 3, function: 'factorial', location: at(18) },
  { frameId: 4, function: 'factorial', location: at(18) },
  { frameId: 5, function: 'main', location: at(11) },
  { frameId: 6, function: 'runtime.main', location: null }
]

const intVariable = (name: string, value: string): Variable => ({
  name,
  typeName: 'int',
  value,
  reference: 0,
  changed: false,
  children: []
})

/** Arguments and locals of each frame of the sample stack. */
export const sampleFrameVariables = (frameId: number): FrameVariables => {
  const empty: FrameVariables = { arguments: [], locals: [] }
  const byFrame: Record<number, FrameVariables> = {
    1: {
      arguments: [intVariable('a', '1'), intVariable('b', '0')],
      locals: [intVariable('total', '1')]
    },
    2: { arguments: [intVariable('n', '1')], locals: [] },
    3: { arguments: [intVariable('n', '2')], locals: [] },
    4: { arguments: [intVariable('n', '3')], locals: [] },
    5: { arguments: [], locals: [intVariable('resultado', '0')] }
  }
  return byFrame[frameId] ?? empty
}
