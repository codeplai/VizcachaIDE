import { describe, expect, it } from 'vitest'
import { TERMINATED_BY_USER } from '../events'
import { runFailed } from './assistant'

describe('runFailed', () => {
  it('explains a run that ended with an error', () => {
    expect(runFailed(1, false)).toBe(true)
  })

  it('does not explain a run that ended well', () => {
    expect(runFailed(0, false)).toBe(false)
  })

  it('does not explain a run the user stopped, whatever its exit code', () => {
    // Python stopped with Ctrl+C ends with KeyboardInterrupt and STATUS_CONTROL_C_EXIT on Windows.
    expect(runFailed(-1073741510, true)).toBe(false)
    expect(runFailed(TERMINATED_BY_USER, false)).toBe(false)
  })
})
