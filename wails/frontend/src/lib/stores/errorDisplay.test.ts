// Errors in the code: a setting hides them while writing, a build or run always shows them.
import { get } from 'svelte/store'
import { afterEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { SAMPLE_MAIN, defaultSettings } from '../bridge/mockData'
import { connectRun, errorsRevealed, resetRun, settings, showErrorsInCode } from '.'

afterEach(() => {
  settings.set(null)
  errorsRevealed.set(false)
  resetRun()
})

describe('errors in the code', () => {
  it('are shown by default, also before the settings load', () => {
    expect(get(showErrorsInCode)).toBe(true)
    settings.set(defaultSettings())
    expect(get(showErrorsInCode)).toBe(true)
  })

  it('can be hidden with the setting', () => {
    settings.set({ ...defaultSettings(), inlineErrors: false })
    expect(get(showErrorsInCode)).toBe(false)
  })

  it('come back when a run starts, and go again when the user edits', async () => {
    const { bridge } = createMockBridge()
    const stop = connectRun(bridge)
    settings.set({ ...defaultSettings(), inlineErrors: false })
    expect(get(showErrorsInCode)).toBe(false)
    await bridge.run.run(SAMPLE_MAIN, [])
    expect(get(showErrorsInCode)).toBe(true)
    errorsRevealed.set(false) // what EditorPane does when the user types
    expect(get(showErrorsInCode)).toBe(false)
    stop()
  })
})
