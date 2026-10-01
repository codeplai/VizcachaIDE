import { get } from 'svelte/store'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Bridge } from '../bridge'
import { PREPARING_DELAY_MS, connectRun, outputLines, resetRun } from '.'

describe('first-time message', () => {
  const handlers: Record<string, (payload: unknown) => void> = {}
  const fakeBridge = {
    on: (name: string, handler: (payload: unknown) => void) => {
      handlers[name] = handler
      return () => {}
    }
  } as unknown as Bridge

  beforeEach(() => {
    vi.useFakeTimers()
    localStorage.clear()
    resetRun()
  })
  afterEach(() => vi.useRealTimers())

  const hasNotice = (): boolean => get(outputLines).some((line) => line.key === 'run.firstBuild')

  it('appears when Go is silent for a while on the first build, and goes away with the output', () => {
    const off = connectRun(fakeBridge)
    handlers['run:started']?.({ target: 'main.go' })
    expect(hasNotice()).toBe(false)
    vi.advanceTimersByTime(PREPARING_DELAY_MS)
    expect(hasNotice()).toBe(true)
    handlers['run:output']?.({ stream: 'stdout', text: 'hi' })
    expect(hasNotice()).toBe(false)
    off()
  })

  it('never appears again once a run has finished', () => {
    const off = connectRun(fakeBridge)
    handlers['run:started']?.({ target: 'main.go' })
    handlers['run:finished']?.({ exitCode: 0, durationMs: 10 })
    handlers['run:started']?.({ target: 'main.go' })
    vi.advanceTimersByTime(PREPARING_DELAY_MS * 2)
    expect(hasNotice()).toBe(false)
    off()
  })
})
