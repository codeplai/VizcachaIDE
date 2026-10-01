// Title bar, keyboard input and debugger gaps found by the parity QA (docs/wails/QA_WAILS.md).
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge as appBridge, type Bridge } from '../bridge'
import { SAMPLE_MAIN } from '../bridge/mockData'
import { setupI18n } from '../i18n'
import OutputPanel from '../panels/OutputPanel.svelte'
import TitleBar from '../shell/TitleBar.svelte'
import {
  activePath,
  buffers,
  connectStores,
  cursor,
  debugActive,
  diagnosticsByFile,
  explained,
  openFile,
  openTabs,
  outputLines,
  programArguments,
  runDiagnostics,
  runLines,
  running,
  runToCursor,
  sendProgramInput
} from '.'

let stop: (() => void) | undefined

const setUp = async (): Promise<Bridge> => {
  const bridge = appBridge
  stop = await connectStores(bridge)
  await openFile(bridge, SAMPLE_MAIN)
  return bridge
}

beforeAll(() => setupI18n('en'))
beforeEach(() => {
  openTabs.set([])
  activePath.set(null)
  buffers.set({})
  programArguments.set('')
  runLines.set([])
  running.set(false)
  diagnosticsByFile.set({})
  explained.set([])
  runDiagnostics.set([])
})
afterEach(() => {
  cleanup()
  stop?.()
})

describe('program arguments in the title bar', () => {
  it('is a visible field next to Run that feeds the program arguments', async () => {
    const bridge = await setUp()
    render(TitleBar)
    const field = screen.getByLabelText('Program arguments') as HTMLInputElement
    await fireEvent.input(field, { target: { value: '--name Ada' } })
    expect(get(programArguments)).toBe('--name Ada')
    const run = vi.spyOn(bridge.run, 'run')
    await fireEvent.keyDown(field, { key: 'Enter' })
    await waitFor(() => expect(run).toHaveBeenCalledWith(SAMPLE_MAIN, ['--name', 'Ada']))
  })

  it('is locked while the program runs', async () => {
    await setUp()
    running.set(true)
    render(TitleBar)
    expect((screen.getByLabelText('Program arguments') as HTMLInputElement).disabled).toBe(true)
  })
})

describe('keyboard input for the running program', () => {
  it('shows the input only while running and sends the typed line', async () => {
    const bridge = await setUp()
    const write = vi.spyOn(bridge.run, 'writeInput')
    render(OutputPanel)
    const label = 'Type here and press Enter to answer your program'
    expect(screen.queryByLabelText(label)).toBeNull()
    running.set(true)
    const field = (await screen.findByLabelText(label)) as HTMLInputElement
    await fireEvent.input(field, { target: { value: 'Ada' } })
    await fireEvent.keyDown(field, { key: 'Enter' })
    expect(write).toHaveBeenCalledWith('Ada')
    expect(get(runLines).at(-1)).toEqual({ kind: 'stdout', text: 'Ada' })
    await waitFor(() => expect(field.value).toBe(''))
  })

  it('echoes the line so the Output reads like a terminal session', async () => {
    const bridge = await setUp()
    await sendProgramInput(bridge, 'hola')
    expect(get(outputLines).map((line) => line.text)).toContain('hola')
  })
})

describe('run to cursor', () => {
  it('asks the debugger to run to the line with the cursor, only while debugging', async () => {
    const bridge = await setUp()
    const runTo = vi.spyOn(bridge.debug, 'runTo')
    cursor.set({ line: 9, column: 3 })
    await runToCursor(bridge)
    expect(runTo).not.toHaveBeenCalled()
    debugActive.set(true)
    await runToCursor(bridge)
    expect(runTo).toHaveBeenCalledWith({ file: SAMPLE_MAIN, line: 9, column: 3 })
    debugActive.set(false)
  })
})
