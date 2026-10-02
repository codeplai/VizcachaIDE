import { get } from 'svelte/store'
import { beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import {
  consoleCursor,
  consoleEntries,
  consoleHistory,
  evalConsole,
  resetConsole,
  stepConsoleHistory
} from './console'

const { bridge } = createMockBridge()

beforeEach(async () => {
  await resetConsole(bridge)
  consoleHistory.set([])
})

describe('console store', () => {
  it('adds the result of an expression to the scrollback', async () => {
    await evalConsole(bridge, '2 + 3')
    expect(get(consoleEntries)).toMatchObject([{ code: '2 + 3', result: '5', error: '' }])
  })

  it('ignores blank snippets', async () => {
    await evalConsole(bridge, '   \n')
    expect(get(consoleEntries)).toEqual([])
    expect(get(consoleHistory)).toEqual([])
  })

  it('keeps variables between snippets and forgets them on reset', async () => {
    await evalConsole(bridge, 'x := 10')
    await evalConsole(bridge, 'x * 2')
    expect(get(consoleEntries)[1]?.result).toBe('20')
    await resetConsole(bridge)
    expect(get(consoleEntries)).toEqual([])
    await evalConsole(bridge, 'x')
    expect(get(consoleEntries)[0]?.error).not.toBe('')
  })

  it('walks the history with up and down', async () => {
    await evalConsole(bridge, '1')
    await evalConsole(bridge, '2')
    await evalConsole(bridge, '2')
    expect(get(consoleHistory)).toEqual(['1', '2'])
    expect(stepConsoleHistory(-1)).toBe('2')
    expect(stepConsoleHistory(-1)).toBe('1')
    expect(stepConsoleHistory(-1)).toBe('1')
    expect(stepConsoleHistory(1)).toBe('2')
    expect(stepConsoleHistory(1)).toBe('')
    expect(get(consoleCursor)).toBeNull()
    expect(stepConsoleHistory(1)).toBeNull()
  })

  it('shows a failing backend as an error entry', async () => {
    const broken = {
      ...bridge,
      console: { eval: () => Promise.reject(new Error('boom')), reset: async () => {} }
    }
    await evalConsole(broken, '1')
    expect(get(consoleEntries)[0]?.error).toContain('boom')
  })
})
