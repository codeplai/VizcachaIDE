import { get } from 'svelte/store'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import {
  activeCodeLanguage,
  activePath,
  capabilities,
  codeLanguageOf,
  connectCodeLanguages,
  connectSettings,
  outputTab,
  pickTool,
  profiles,
  settings,
  toolStatus,
  tools
} from '.'

describe('code languages from the backend', () => {
  let stop: (() => void) | undefined
  const { bridge } = createMockBridge()

  beforeEach(async () => {
    activePath.set(null)
    stop = await connectSettings(bridge)
    const offCodeLanguages = await connectCodeLanguages(bridge)
    const offSettings = stop
    stop = () => {
      offSettings()
      offCodeLanguages()
    }
  })
  afterEach(() => stop?.())

  it('loads the profiles and finds a language by lower-case extension', () => {
    expect(get(profiles).map((profile) => profile.id)).toEqual(['go', 'python', 'cpp', 'rust'])
    expect(codeLanguageOf('C:/work/main.GO')).toBe('go')
    expect(codeLanguageOf('C:/work/tool.py')).toBe('python')
    expect(codeLanguageOf('C:/work/Main.CPP')).toBe('cpp')
    expect(codeLanguageOf('notes.txt')).toBeNull()
    expect(codeLanguageOf('Makefile')).toBeNull()
  })

  it('finds the language of an untitled name the same way', () => {
    expect(codeLanguageOf('untitled/main.py')).toBe('python')
    expect(codeLanguageOf('untitled/main2.cpp')).toBe('cpp')
  })

  it('follows the open file, and falls back to the default for new files', async () => {
    expect(get(activeCodeLanguage)).toBe('go')
    activePath.set('src/app.py')
    expect(get(activeCodeLanguage)).toBe('python')
    activePath.set('readme.txt')
    expect(get(activeCodeLanguage)).toBe('go')
    await bridge.settings.save({ ...(await bridge.settings.get()), defaultCodeLanguage: 'cpp' })
    expect(get(activeCodeLanguage)).toBe('cpp')
    await bridge.settings.save({ ...(await bridge.settings.get()), defaultCodeLanguage: 'go' })
  })

  it('derives the capabilities of the active language', () => {
    activePath.set('main.go')
    expect(get(capabilities)).toMatchObject({
      console: true,
      build: true,
      threadsLabel: 'debug.goroutines'
    })
    activePath.set('main.py')
    expect(get(capabilities)).toMatchObject({
      console: true,
      build: false,
      threadsLabel: 'debug.threads'
    })
    activePath.set('main.cpp')
    expect(get(capabilities)).toMatchObject({ console: false, build: true, packageActions: [] })
  })

  it('leaves the Console tab when the file changes to a language without console', () => {
    activePath.set('main.go')
    outputTab.set('console')
    activePath.set('main.cpp')
    expect(get(outputTab)).toBe('output')
    outputTab.set('output')
  })

  it('reports the status of each tool and reloads it after choosing a path', async () => {
    expect(toolStatus('dlv')?.version).toBe('1.27.2')
    expect(toolStatus('nope')).toBeUndefined()
    await pickTool(bridge, 'go')
    expect(toolStatus('go', get(tools))?.source).toBe('configured')
    await bridge.settings.save({
      ...(get(settings) ?? (await bridge.settings.get())),
      toolPaths: {}
    })
    await vi.waitFor(() => expect(toolStatus('go', get(tools))?.source).toBe('bundled'))
  })
})
