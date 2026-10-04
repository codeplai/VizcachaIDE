// What the backend now offers: program arguments for Run and Debug, "Choose…" a tool, and the
// language it resolved for "auto".
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge as appBridge, type Bridge } from '../bridge'
import { createMockBridge } from '../bridge/mock'
import { SAMPLE_MAIN } from '../bridge/mockData'
import { setupI18n } from '../i18n'
import SettingsDialog from '../shell/SettingsDialog.svelte'
import {
  activePath,
  breakpoints,
  buffers,
  debugActive,
  connectStores,
  interfaceLanguage,
  notice,
  openDialog,
  openFile,
  openTabs,
  programArguments,
  runActiveFile,
  settingsTab,
  startDebugging,
  toolStatus,
  tools,
  updateSettings
} from '.'

let stop: (() => void) | undefined

const setUp = async (): Promise<Bridge> => {
  const { bridge } = createMockBridge()
  stop = await connectStores(bridge)
  await openFile(bridge, SAMPLE_MAIN)
  return bridge
}

beforeAll(() => setupI18n('en'))
beforeEach(() => {
  openTabs.set([])
  activePath.set(null)
  buffers.set({})
  notice.set(null)
  programArguments.set('')
})
afterEach(() => {
  cleanup()
  stop?.()
})

describe('program arguments', () => {
  it('Run and Debug both receive what the user typed', async () => {
    const bridge = await setUp()
    const run = vi.spyOn(bridge.run, 'run')
    const start = vi.spyOn(bridge.debug, 'start')
    programArguments.set('--name "Ada Lovelace"')
    await runActiveFile(bridge)
    expect(run).toHaveBeenCalledWith(SAMPLE_MAIN, ['--name', '"Ada', 'Lovelace"'])
    await startDebugging(bridge)
    expect(start.mock.calls[0]?.[2]).toBe('--name "Ada Lovelace"')
  })

  it('Debug sends the breakpoints of the other files of the same language too', async () => {
    const bridge = await setUp()
    const start = vi.spyOn(bridge.debug, 'start')
    const helper = SAMPLE_MAIN.replace(/main\.go$/, 'helper.go')
    breakpoints.set({ [helper]: [3], [SAMPLE_MAIN]: [5], 'C:/otro/app.py': [2] })
    debugActive.set(false)
    await startDebugging(bridge)
    expect(start).toHaveBeenCalled()
    const files = (start.mock.calls[0]?.[1] ?? []).map((point) => point.location.file)
    expect(files).toContain(helper)
    expect(files).toContain(SAMPLE_MAIN)
    expect(files).not.toContain('C:/otro/app.py')
    breakpoints.set({})
  })

  it('says so and does not run when the arguments cannot be split', async () => {
    const bridge = await setUp()
    vi.spyOn(bridge.run, 'splitArguments').mockRejectedValue(new Error('unclosed quote'))
    const run = vi.spyOn(bridge.run, 'run')
    programArguments.set('"open')
    await runActiveFile(bridge)
    expect(run).not.toHaveBeenCalled()
    expect(get(notice)?.messageKey).toBe('errors.invalidArgs')
  })
})

describe('tools and language from the backend', () => {
  it('"Choose…" saves the picked path and shows where the tool comes from', async () => {
    const bridge = appBridge
    stop = await connectStores(bridge)
    openDialog.set('settings')
    settingsTab.set('tools')
    render(SettingsDialog)
    const buttons = await screen.findAllByRole('button', { name: 'Choose…' })
    await fireEvent.click(buttons[0] as HTMLElement)
    await waitFor(() => expect(toolStatus('go', get(tools))?.source).toBe('configured'))
    expect((await bridge.settings.get()).toolPaths['go']).toBe('C:\\tools\\go.exe')
    expect(await screen.findByText(/Location you chose/)).toBeTruthy()
  })

  it('the language for "auto" is the one the backend resolved', async () => {
    const bridge = await setUp()
    expect(get(interfaceLanguage)).toBe('en')
    vi.spyOn(bridge.settings, 'resolvedLanguage').mockResolvedValue('es')
    await updateSettings(bridge, { language: 'en' })
    await updateSettings(bridge, { language: 'auto' })
    await waitFor(() => expect(get(interfaceLanguage)).toBe('es'))
    await updateSettings(bridge, { language: 'en' })
    expect(get(interfaceLanguage)).toBe('en')
  })
})
