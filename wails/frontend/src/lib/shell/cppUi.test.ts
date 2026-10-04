import { language as languageFacet } from '@codemirror/language'
import { EditorState } from '@codemirror/state'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge } from '../bridge'
import { cppProfile, languageProfiles } from '../bridge/languageProfiles'
import type { ToolStatus } from '../domain'
import { languageExtensionsFor } from '../editor/languageSupport'
import { newFileTemplates } from '../editor/templates'
import { setupI18n } from '../i18n'
import {
  activePath,
  connectSettings,
  debugInputEnabled,
  debuggedPath,
  lastRunConfiguration,
  outputLines,
  profiles,
  resetRun,
  runResult,
  tools
} from '../stores'
import FirstRunCppHint from './FirstRunCppHint.svelte'
import FirstRunTools from './FirstRunTools.svelte'
import MoreMenu from './MoreMenu.svelte'

beforeAll(() => setupI18n('en'))

let stopSettings: (() => void) | undefined

beforeEach(async () => {
  stopSettings = await connectSettings(bridge)
  profiles.set(languageProfiles)
  await bridge.settings.save({
    ...(await bridge.settings.get()),
    enabledCodeLanguages: ['cpp'],
    defaultCodeLanguage: 'cpp'
  })
})

afterEach(() => {
  stopSettings?.()
  cleanup()
  vi.restoreAllMocks()
  activePath.set(null)
  debuggedPath.set(null)
  lastRunConfiguration.set(null)
  resetRun()
})

const openMore = async (): Promise<HTMLElement[]> => {
  await fireEvent.keyDown(screen.getByRole('button', { name: /More/ }), { key: 'Enter' })
  return screen.findAllByRole('menuitem')
}

describe('Build entry of the More menu', () => {
  it('is there for a saved C++ file and builds it through the bridge', async () => {
    const build = vi.spyOn(bridge.run, 'build')
    activePath.set('hola-cpp/main.cpp')
    render(MoreMenu)
    const items = await openMore()
    const entry = items.find((item) => /build/i.test(item.textContent ?? ''))
    expect(entry).toBeTruthy()
    await fireEvent.click(entry as HTMLElement)
    await waitFor(() => expect(build).toHaveBeenCalledWith('hola-cpp/main.cpp', []))
  })

  it('is there for Go, which can build too', async () => {
    activePath.set('hola-go/main.go')
    render(MoreMenu)
    const items = await openMore()
    expect(items.some((item) => /build/i.test(item.textContent ?? ''))).toBe(true)
  })

  it('is absent without the build capability (Python) and for an untitled file', async () => {
    activePath.set('hola-py/main.py')
    const view = render(MoreMenu)
    const hasBuild = (items: HTMLElement[]) => items.some((i) => /build/i.test(i.textContent ?? ''))
    expect(hasBuild(await openMore())).toBe(false)
    view.unmount()
    activePath.set('untitled/main.cpp')
    render(MoreMenu)
    expect(hasBuild(await openMore())).toBe(false)
  })
})

describe('First start hints for C++', () => {
  const missing: ToolStatus = {
    id: 'cxx',
    codeLanguage: 'cpp',
    role: 'compiler',
    version: '',
    source: 'missing',
    path: ''
  }
  const spec = cppProfile.tools[0]!

  it('shows the Command Line Tools command with a copy button on macOS', async () => {
    const copy = vi.spyOn(bridge.system, 'writeClipboard').mockResolvedValue()
    render(FirstRunCppHint, { spec, platform: 'macos' })
    expect(screen.getByText('xcode-select --install')).toBeTruthy()
    expect(screen.getByText(/Command Line Tools/)).toBeTruthy()
    await fireEvent.click(screen.getByRole('button'))
    await waitFor(() => expect(copy).toHaveBeenCalledWith('xcode-select --install'))
  })

  it('shows the apt command with a copy button on Linux', async () => {
    const copy = vi.spyOn(bridge.system, 'writeClipboard').mockResolvedValue()
    render(FirstRunCppHint, { spec, platform: 'linux' })
    await fireEvent.click(screen.getByRole('button'))
    await waitFor(() =>
      expect(copy).toHaveBeenCalledWith('sudo apt install g++ lldb clangd clang-format')
    )
  })

  it('says the full installer includes it and links WinLibs on Windows', async () => {
    const open = vi.spyOn(bridge.system, 'openUrl').mockImplementation(() => {})
    render(FirstRunCppHint, { spec, platform: 'windows' })
    expect(screen.getByText(/full installer already includes a C\+\+ compiler/)).toBeTruthy()
    await fireEvent.click(screen.getByRole('button'))
    expect(open).toHaveBeenCalledWith('https://winlibs.com/')
  })

  it('appears in the tools step only when the compiler is missing', async () => {
    tools.set([{ ...missing, version: '15.2.0', source: 'path' }])
    const view = render(FirstRunTools)
    expect(await screen.findByText(/C\+\+ · /)).toBeTruthy()
    expect(view.container.querySelector('.hint')).toBeNull()
    view.unmount()
    tools.set([missing])
    const again = render(FirstRunTools)
    await waitFor(() => expect(again.container.querySelector('.hint')).not.toBeNull())
  })
})

describe('C++ in the editor and the panels', () => {
  it('has the greeting template of the plan and a blank one', () => {
    const { hello, blank, extension } = newFileTemplates.cpp
    expect(hello).toContain('#include <iostream>')
    expect(hello).toContain('using namespace std;')
    expect(hello).toContain('cout << "Hola, C++" << endl;')
    expect(blank).toContain('int main()')
    expect(extension).toBe('.cpp')
  })

  it.each(['main.cpp', 'util.h', 'util.hpp'])('highlights %s with the C++ language', (name) => {
    const state = EditorState.create({
      doc: '#include <iostream>\nint main() { return 0; }\n',
      extensions: languageExtensionsFor(name, cppProfile)
    })
    expect(state.facet(languageFacet)?.name).toBe('cpp')
    expect(state.tabSize).toBe(4)
  })

  it('lets the debuggee read the keyboard', () => {
    debuggedPath.set('hola-cpp/main.cpp')
    expect(get(debugInputEnabled)).toBe(true)
    debuggedPath.set('hola-go/main.go')
    expect(get(debugInputEnabled)).toBe(false)
  })

  it('ends a crashed C++ run with the crash verdict, and a Go exit code with the code', () => {
    const cppConfig = { ...cppRun('cpp'), codeLanguage: 'cpp' as const }
    lastRunConfiguration.set(cppConfig)
    runResult.set({ exitCode: 139, durationMs: 10 })
    expect(get(outputLines).at(-1)?.key).toBe('run.crashed')
    runResult.set({ exitCode: 0xc0000005, durationMs: 10 })
    expect(get(outputLines).at(-1)?.key).toBe('run.crashed')
    runResult.set({ exitCode: 1, durationMs: 10 })
    expect(get(outputLines).at(-1)?.key).toBe('run.exitCode')
    lastRunConfiguration.set({ ...cppConfig, codeLanguage: 'go' })
    runResult.set({ exitCode: 139, durationMs: 10 })
    expect(get(outputLines).at(-1)?.key).toBe('run.exitCode')
  })
})

const cppRun = (codeLanguage: 'cpp') => ({
  codeLanguage,
  target: 'main.cpp',
  workingDir: '.',
  mode: 'file' as const,
  programArgs: [],
  project: null,
  echo: true
})
