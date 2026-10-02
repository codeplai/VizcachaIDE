import { get } from 'svelte/store'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/svelte'
import { bridge, mockControls } from '../bridge'
import { SAMPLE_MAIN } from '../bridge/mockData'
import type { ExplainedDiagnostic } from '../domain'
import { connectStores, resetRun, runDiagnostics } from '.'

let stop: (() => void) | undefined

const warning: ExplainedDiagnostic = {
  diagnostic: {
    location: { file: SAMPLE_MAIN, line: 9, column: 2 },
    end: null,
    severity: 'warning',
    message: 'unreachable code',
    rawText: './main.go:9:2: unreachable code',
    source: 'vet',
    code: 'V-UNREACHABLE'
  },
  explanation: null
}

beforeEach(async () => {
  resetRun()
  runDiagnostics.set([])
  stop = await connectStores(bridge)
})

afterEach(() => {
  stop?.()
  vi.restoreAllMocks()
})

describe('go vet after a run', () => {
  it('feeds the vet output to the explainer and shows the warnings', async () => {
    const vet = vi.spyOn(bridge.run, 'vet').mockResolvedValue('./main.go:9:2: unreachable code\n')
    const explain = vi.spyOn(bridge.assistant, 'explain').mockResolvedValue([warning])
    await mockControls?.play('write')
    await waitFor(() => expect(get(runDiagnostics)).toHaveLength(1))
    expect(vet).toHaveBeenCalledTimes(1)
    expect(explain).toHaveBeenCalledWith('./main.go:9:2: unreachable code\n', expect.any(String))
    expect(get(runDiagnostics)[0]?.code).toBe('V-UNREACHABLE')
  })

  it('does nothing when vet finds nothing', async () => {
    const vet = vi.spyOn(bridge.run, 'vet').mockResolvedValue('')
    const explain = vi.spyOn(bridge.assistant, 'explain')
    await mockControls?.play('write')
    await waitFor(() => expect(vet).toHaveBeenCalled())
    expect(explain).not.toHaveBeenCalled()
    expect(get(runDiagnostics)).toHaveLength(0)
  })

  it('is not run when the build failed', async () => {
    const vet = vi.spyOn(bridge.run, 'vet').mockResolvedValue('')
    await mockControls?.play('error')
    await waitFor(() => expect(get(runDiagnostics).length).toBeGreaterThan(0))
    expect(vet).not.toHaveBeenCalled()
  })
})
