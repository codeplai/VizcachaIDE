import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { languageProfiles } from '../bridge/languageProfiles'
import type { LanguageProfile, RunConfiguration } from '../domain'
import { setupI18n } from '../i18n'
import {
  activePath,
  buffers,
  lastRunConfiguration,
  missingToolIn,
  notice,
  openDialog,
  profiles,
  resetRun,
  runActiveFile,
  runLines,
  sendProgramInput
} from '.'

const { bridge } = createMockBridge()

const clangd = {
  id: 'clangd',
  role: 'languageServer' as const,
  labelKey: 'settings.toolClangd',
  missingKey: 'errors.clangdNotFound',
  installUrl: 'https://clangd.llvm.org/installation',
  installCommand: 'sudo apt install clangd',
  providedBy: ''
}

const withClangd = (): LanguageProfile[] =>
  languageProfiles.map((profile) =>
    profile.id === 'cpp' ? { ...profile, tools: [clangd] } : profile
  )

const configuration = (echo: boolean): RunConfiguration => ({
  codeLanguage: 'python',
  target: 'main.py',
  workingDir: '.',
  mode: 'file',
  programArgs: [],
  project: null,
  echo
})

beforeAll(() => setupI18n('en'))

beforeEach(() => {
  resetRun()
  notice.set(null)
  buffers.set({})
  profiles.set(withClangd())
  activePath.set('C:/work/main.go')
})

afterEach(() => {
  vi.restoreAllMocks()
  activePath.set(null)
  lastRunConfiguration.set(null)
})

describe('missingToolIn', () => {
  it('reads the id the backend puts in the error', () => {
    expect(missingToolIn('tool "dlv": tool not found')).toBe('dlv')
    expect(missingToolIn('debug: tool "clangd": tool not found')).toBe('clangd')
  })

  it('still recognises the free text of a 2.0 backend', () => {
    expect(missingToolIn('exec: "go": executable file not found in %PATH%')).toBe('go')
    expect(missingToolIn('dlv not found')).toBe('dlv')
    expect(missingToolIn('boom')).toBeNull()
  })
})

describe('a missing tool while running', () => {
  it('shows the notice of the ToolSpec with install, copy and choose actions', async () => {
    vi.spyOn(bridge.run, 'run').mockRejectedValue(new Error('tool "clangd": tool not found'))
    const open = vi.spyOn(bridge.system, 'openUrl').mockImplementation(() => {})
    const write = vi.fn(async () => {})
    vi.stubGlobal('navigator', { clipboard: { writeText: write } })
    await runActiveFile(bridge)
    const shown = get(notice)
    expect(shown?.messageKey).toBe('errors.clangdNotFound')
    expect(shown?.detail).toBe('sudo apt install clangd')
    expect(shown?.actions.map((action) => action.labelKey)).toEqual([
      'errors.toolInstall',
      'errors.toolCopyCommand',
      'errors.goNotFoundChoose'
    ])
    shown?.actions[0]?.run()
    expect(open).toHaveBeenCalledWith(clangd.installUrl)
    shown?.actions[1]?.run()
    expect(write).toHaveBeenCalledWith('sudo apt install clangd')
    shown?.actions[2]?.run()
    expect(get(openDialog)).toBe('settings')
    vi.unstubAllGlobals()
    openDialog.set(null)
  })

  it('keeps the Go messages of 2.0 when the text names no tool', async () => {
    vi.spyOn(bridge.run, 'run').mockRejectedValue(new Error('exec: "go": executable not found'))
    await runActiveFile(bridge)
    expect(get(notice)?.messageKey).toBe('errors.goNotFound')
  })

  it('rethrows an error that is not about tools', async () => {
    vi.spyOn(bridge.run, 'run').mockRejectedValue(new Error('disk on fire'))
    await expect(runActiveFile(bridge)).rejects.toThrow('disk on fire')
  })
})

describe('a language without an adapter yet', () => {
  it('says the action is not available for C++', async () => {
    activePath.set('C:/work/app.cpp')
    await runActiveFile(bridge)
    const shown = get(notice)
    expect(shown?.messageKey).toBe('errors.unsupportedAction')
    expect(shown?.values['codeLanguage']).toBe('C++')
  })
})

describe('typed input', () => {
  it('is shown in Output when the program does not echo it', async () => {
    lastRunConfiguration.set(configuration(false))
    await sendProgramInput(bridge, 'Ada')
    expect(get(runLines).at(-1)).toEqual({ kind: 'stdout', text: 'Ada' })
  })

  it('is not repeated when the run echoes it (a PTY)', async () => {
    lastRunConfiguration.set(configuration(true))
    const write = vi.spyOn(bridge.run, 'writeInput')
    await sendProgramInput(bridge, 'Ada')
    expect(write).toHaveBeenCalledWith('Ada')
    expect(get(runLines)).toEqual([])
  })
})
