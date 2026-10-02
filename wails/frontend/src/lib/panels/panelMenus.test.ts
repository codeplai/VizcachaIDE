import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge, mockControls, type Unsubscribe } from '../bridge'
import { setupI18n } from '../i18n'
import { clearOutput, connectStores, consoleEntries, outputLines, runLines } from '../stores'
import ConsolePanel from './ConsolePanel.svelte'
import OutputPanel from './OutputPanel.svelte'
import ProblemsPanel from './ProblemsPanel.svelte'

let stop: Unsubscribe | undefined

beforeAll(() => {
  setupI18n('en')
})

beforeEach(async () => {
  stop = await connectStores(bridge)
})

afterEach(() => {
  cleanup()
  stop?.()
  vi.restoreAllMocks()
})

const openMenu = async (target: Element): Promise<string[]> => {
  await fireEvent.contextMenu(target)
  await waitFor(() => expect(screen.getAllByRole('menuitem').length).toBeGreaterThan(0))
  return screen
    .getAllByRole('menuitem')
    .map((item) => item.textContent?.replace(/\s+/g, ' ').trim() ?? '')
}

describe('right-click menus of the text panels', () => {
  it('Output offers copy, copy all, paste into the input, select all and clear', async () => {
    render(OutputPanel)
    const items = await openMenu(screen.getByRole('log'))
    expect(items).toEqual([
      'Copy Ctrl+C',
      'Copy all',
      "Paste into the program's input",
      'Select all',
      'Clear'
    ])
  })

  it('Copy all puts the whole output in the clipboard', async () => {
    const write = vi.spyOn(bridge.system, 'writeClipboard').mockResolvedValue()
    await mockControls?.play('write')
    render(OutputPanel)
    await openMenu(screen.getByRole('log'))
    await fireEvent.click(screen.getByRole('menuitem', { name: 'Copy all' }))
    await waitFor(() => expect(write).toHaveBeenCalled())
    expect(write.mock.calls[0]?.[0]).toContain('Hola, Go')
  })

  it('Clear empties the Output panel', () => {
    runLines.set([{ kind: 'stdout', text: 'hola' }])
    clearOutput()
    expect(get(runLines)).toEqual([])
    expect(get(outputLines)).toEqual([])
  })

  it('Problems offers copy, copy all and select all', async () => {
    render(ProblemsPanel)
    const items = await openMenu(document.querySelector('.list') as Element)
    expect(items).toEqual(['Copy Ctrl+C', 'Copy all', 'Select all'])
  })

  it('Console: paste goes to the input line, clear screen keeps the session', async () => {
    vi.spyOn(bridge.system, 'readClipboard').mockResolvedValue('2 + 3')
    consoleEntries.set([{ id: 1, code: 'x := 1', result: '', output: '', error: '' }])
    render(ConsolePanel)
    const area = document.querySelector('.console .scroll') as Element
    await openMenu(area)
    await fireEvent.click(screen.getByRole('menuitem', { name: /^Paste/ }))
    await waitFor(() =>
      expect((screen.getByRole('textbox') as HTMLTextAreaElement).value).toBe('2 + 3')
    )
    await openMenu(area)
    await fireEvent.click(screen.getByRole('menuitem', { name: /^Clear screen/ }))
    expect(get(consoleEntries)).toEqual([])
  })
})
