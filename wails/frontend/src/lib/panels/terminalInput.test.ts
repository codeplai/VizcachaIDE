import { describe, expect, it } from 'vitest'
import { orderedInput } from './terminalInput'

describe('typing in the terminal', () => {
  it('reaches the shell in order even when the calls answer out of order', async () => {
    const received: string[] = []
    const answers: (() => void)[] = []
    const send = (data: string): Promise<void> =>
      new Promise((resolve) => {
        received.push(data)
        answers.push(resolve)
      })
    const type = orderedInput(send)
    for (const key of 'clang++') type(key)
    expect(received).toEqual(['c']) // one write in flight
    answers.shift()?.()
    await Promise.resolve()
    await Promise.resolve()
    expect(received).toEqual(['c', 'lang++']) // the rest together, in order
    answers.shift()?.()
    await Promise.resolve()
    type(' --version\r')
    await Promise.resolve()
    expect(received.join('')).toBe('clang++ --version\r')
  })

  it('keeps going after a write fails', async () => {
    const received: string[] = []
    const type = orderedInput(async (data) => {
      received.push(data)
      if (data === 'a') throw new Error('session gone')
    })
    type('a')
    await Promise.resolve()
    type('b')
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(received).toEqual(['a', 'b'])
  })
})
