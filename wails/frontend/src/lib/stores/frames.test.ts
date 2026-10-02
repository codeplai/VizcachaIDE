import { describe, expect, it } from 'vitest'
import type { StackFrame } from '../domain'
import { argumentsText, callBoxes, shortFunctionName } from './frames'

const frame = (frameId: number, name: string, line = 10): StackFrame => ({
  frameId,
  function: name,
  location: { file: 'main.go', line, column: 1 }
})

// Innermost first, as the debugger sends it.
const recursion = (depth: number): StackFrame[] => [
  ...Array.from({ length: depth }, (_, index) => frame(index + 1, 'main.factorial', 5)),
  frame(depth + 1, 'main.main', 11),
  frame(depth + 2, 'runtime.main')
]

describe('callBoxes', () => {
  it('orders the calls from the outermost to the innermost and hides runtime frames', () => {
    const { boxes, hidden } = callBoxes(recursion(3))

    expect(boxes.map((box) => box.name)).toEqual(['main', 'factorial', 'factorial', 'factorial'])
    expect(boxes.map((box) => box.depth)).toEqual([0, 1, 2, 3])
    expect(hidden).toBe(0)
  })

  it('marks only the innermost call as current', () => {
    const { boxes } = callBoxes(recursion(3))

    expect(boxes.filter((box) => box.current).map((box) => box.frameId)).toEqual([1])
    expect(boxes.at(-1)?.current).toBe(true)
  })

  it('keeps the innermost calls and counts the ones left out', () => {
    const { boxes, hidden } = callBoxes(recursion(12), 8)

    expect(boxes).toHaveLength(8)
    expect(hidden).toBe(5) // 12 factorial + main = 13 user calls
    expect(boxes.at(-1)?.frameId).toBe(1)
    expect(boxes[0]?.name).toBe('factorial')
  })

  it('has nothing to draw without a stack', () => {
    expect(callBoxes([])).toEqual({ boxes: [], hidden: 0 })
  })
})

describe('names and arguments', () => {
  it('drops the main package prefix only', () => {
    expect(shortFunctionName('main.factorial')).toBe('factorial')
    expect(shortFunctionName('main.(*Cuenta).Depositar')).toBe('(*Cuenta).Depositar')
    expect(shortFunctionName('otro.Hacer')).toBe('otro.Hacer')
  })

  it('writes the arguments as name=value', () => {
    const variable = (name: string, value: string) => ({
      name,
      typeName: 'int',
      value,
      reference: 0,
      changed: false,
      children: []
    })

    expect(argumentsText({ arguments: [variable('a', '5'), variable('b', '7')], locals: [] })).toBe(
      'a=5, b=7'
    )
    expect(argumentsText(undefined)).toBe('')
  })
})
