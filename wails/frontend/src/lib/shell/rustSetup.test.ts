import { cleanup, fireEvent, render, waitFor } from '@testing-library/svelte'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge } from '../bridge'
import { languageProfiles, rustProfile } from '../bridge/languageProfiles'
import type { ToolStatus } from '../domain'
import { setupI18n } from '../i18n'
import {
  RUST_ADVICE_COMMANDS,
  activePath,
  connectSettings,
  fileTree,
  lastRunConfiguration,
  memberChoice,
  openDialog,
  packageTask,
  profiles,
  resetRun,
  rustAdvice,
  rustupInstallCommand,
  tools
} from '../stores'
import FirstRunRustHint from './FirstRunRustHint.svelte'
import FirstRunTools from './FirstRunTools.svelte'
import SettingsTools from './SettingsTools.svelte'

beforeAll(() => setupI18n('en'))

let stopSettings: (() => void) | undefined

beforeEach(async () => {
  stopSettings = await connectSettings(bridge)
  profiles.set(languageProfiles)
  await bridge.settings.save({
    ...(await bridge.settings.get()),
    enabledCodeLanguages: ['rust'],
    defaultCodeLanguage: 'rust'
  })
})

afterEach(() => {
  stopSettings?.()
  cleanup()
  vi.restoreAllMocks()
  activePath.set(null)
  fileTree.set(null)
  openDialog.set(null)
  packageTask.set(null)
  memberChoice.set(null)
  rustAdvice.set([])
  lastRunConfiguration.set(null)
  resetRun()
})

const missing = (id: string, role: ToolStatus['role']): ToolStatus => ({
  id,
  codeLanguage: 'rust',
  role,
  version: '',
  source: 'missing',
  path: ''
})

const found = (status: ToolStatus): ToolStatus => ({ ...status, version: '1.99.0', source: 'path' })

describe('First start hint for Rust', () => {
  const spec = rustProfile.tools[0]!

  it('links rustup and gives the GNU command on Windows, copyable', async () => {
    const open = vi.spyOn(bridge.system, 'openUrl').mockImplementation(() => {})
    const copy = vi.spyOn(bridge.system, 'writeClipboard').mockResolvedValue()
    const view = render(FirstRunRustHint, { spec, platform: 'windows' })
    const command = view.container.querySelector('code')?.textContent
    expect(command).toBe(
      'rustup-init.exe --default-host x86_64-pc-windows-gnu --default-toolchain stable -y'
    )
    const [link, copyButton] = view.container.querySelectorAll('button')
    await fireEvent.click(link as HTMLElement)
    expect(open).toHaveBeenCalledWith('https://rustup.rs')
    await fireEvent.click(copyButton as HTMLElement)
    await waitFor(() => expect(copy).toHaveBeenCalledWith(command))
  })

  it.each(['macos', 'linux'] as const)('gives the curl command on %s', (platform) => {
    const view = render(FirstRunRustHint, { spec, platform })
    expect(view.container.querySelector('code')?.textContent).toBe(rustupInstallCommand(platform))
    expect(rustupInstallCommand(platform)).toContain('https://sh.rustup.rs')
  })

  it('appears in the tools step only when rustc is missing', async () => {
    tools.set([found(missing('rustc', 'compiler')), found(missing('lldb-dap', 'debugAdapter'))])
    const view = render(FirstRunTools)
    await waitFor(() => expect(view.container.querySelector('h3')).not.toBeNull())
    expect(view.container.querySelector('.hint')).toBeNull()
    view.unmount()
    tools.set([missing('rustc', 'compiler'), found(missing('lldb-dap', 'debugAdapter'))])
    const again = render(FirstRunTools)
    await waitFor(() => expect(again.container.querySelector('.hint')).not.toBeNull())
  })

  it('says debugging needs lldb-dap when only that tool is missing', async () => {
    tools.set([found(missing('rustc', 'compiler')), missing('lldb-dap', 'debugAdapter')])
    const view = render(FirstRunTools)
    await waitFor(() => expect(view.container.querySelectorAll('p')).toHaveLength(1))
    expect(view.container.querySelector('.hint')).toBeNull()
  })
})

describe('Toolchain advice in Settings → Tools', () => {
  it('shows nothing while the store is empty', () => {
    const view = render(SettingsTools)
    expect(view.container.querySelector('[data-advice]')).toBeNull()
  })

  it.each([
    ['errors.rustMsvcHost', 'rustup default stable-x86_64-pc-windows-gnu'],
    ['errors.rustNotStable', 'rustup update stable'],
    ['errors.rustTooOld', 'rustup update stable'],
    ['errors.rustupNoToolchain', 'rustup default stable']
  ])('shows %s with its command to copy', async (key, command) => {
    rustAdvice.set([key])
    const copy = vi.spyOn(bridge.system, 'writeClipboard').mockResolvedValue()
    const view = render(SettingsTools)
    const note = view.container.querySelector(`[data-advice="${key}"]`)
    expect(note).not.toBeNull()
    expect(note?.querySelector('code')?.textContent).toBe(command)
    await fireEvent.click(note?.querySelector('button') as HTMLElement)
    await waitFor(() => expect(copy).toHaveBeenCalledWith(command))
    expect(RUST_ADVICE_COMMANDS[key]).toBe(command)
  })

  it('ignores a key it does not know and shows advice only under the Rust group', () => {
    rustAdvice.set(['errors.somethingNew'])
    const view = render(SettingsTools)
    expect(view.container.querySelector('[data-advice]')).toBeNull()
  })

  it('is not shown when Rust is not one of the chosen languages', async () => {
    await bridge.settings.save({
      ...(await bridge.settings.get()),
      enabledCodeLanguages: ['go']
    })
    rustAdvice.set(['errors.rustTooOld'])
    const view = render(SettingsTools)
    expect(view.container.querySelector('[data-advice]')).toBeNull()
  })
})
