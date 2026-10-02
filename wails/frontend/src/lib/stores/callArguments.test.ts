import { describe, expect, it } from 'vitest'
import type { Variable } from '../domain'
import { parameterNames, splitArguments } from './callArguments'

const variable = (name: string, value = '1'): Variable => ({
  name,
  typeName: 'int',
  value,
  reference: 0,
  changed: false,
  children: []
})

const SOURCE = `package main

func factorial(n int) int {
\treturn n
}

func suma(a, b int, nombre string) int { return a + b }

func (c *Cuenta) Depositar(monto float64) {}
`

describe('call arguments', () => {
  it('reads parameter names from the function declaration', () => {
    expect(parameterNames(SOURCE, 'main.factorial')).toEqual(['n'])
    expect(parameterNames(SOURCE, 'main.suma')).toEqual(['a', 'b', 'nombre'])
    expect(parameterNames(SOURCE, 'main.(*Cuenta).Depositar')).toEqual(['monto'])
    expect(parameterNames(SOURCE, 'main.main')).toEqual([])
  })

  it('moves parameters out of the locals and hides the ~r return slots', () => {
    const split = splitArguments(
      { arguments: [], locals: [variable('n', '5'), variable('~r0', '0'), variable('total')] },
      ['n']
    )
    expect(split.arguments.map((item) => item.name)).toEqual(['n'])
    expect(split.locals.map((item) => item.name)).toEqual(['total'])
  })

  it('keeps arguments the debugger already separated', () => {
    const split = splitArguments({ arguments: [variable('x')], locals: [variable('~r0')] }, ['x'])
    expect(split.arguments.map((item) => item.name)).toEqual(['x'])
    expect(split.locals).toEqual([])
  })
})
