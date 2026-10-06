import { get } from 'svelte/store'
import { beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { languageProfiles } from '../bridge/languageProfiles'
import {
  activePath,
  buffers,
  codeLanguageOf,
  isUntitled,
  newFile,
  openTabs,
  openUntitled,
  profiles,
  settings
} from '.'

const { bridge } = createMockBridge()

beforeEach(() => {
  buffers.set({})
  openTabs.set([])
  activePath.set(null)
  profiles.set(languageProfiles)
})

describe('untitled files', () => {
  it('names a new file with the extension of its language, and the lookup finds it', async () => {
    expect(await openUntitled(bridge, 'go')).toBe('untitled/main.go')
    expect(await openUntitled(bridge, 'python')).toBe('untitled/main.py')
    expect(await openUntitled(bridge, 'cpp')).toBe('untitled/main.cpp')
    for (const path of get(openTabs)) expect(codeLanguageOf(path)).not.toBeNull()
    expect(codeLanguageOf('untitled/main.py')).toBe('python')
  })

  it('numbers a second file of the same language', async () => {
    await openUntitled(bridge, 'python')
    expect(await openUntitled(bridge, 'python')).toBe('untitled/main2.py')
  })

  it('starts each language with its template', async () => {
    const python = await openUntitled(bridge, 'python', 'hello')
    expect(get(buffers)[python]).toBe('print("Hola, Python")\n')
    const cpp = await openUntitled(bridge, 'cpp', 'hello')
    expect(get(buffers)[cpp]).toContain('#include <iostream>')
    expect(get(buffers)[cpp]).toContain('using namespace std;')
    expect(get(buffers)[cpp]).toContain('cout << "Hola, C++" << endl;')
    const go = await openUntitled(bridge, 'go')
    expect(get(buffers)[go]).toContain('package main')
  })

  it('uses the absolute path the backend gives, one folder per file, and still knows it is unsaved', async () => {
    const real = {
      ...bridge,
      language: {
        ...bridge.language,
        untitledFile: async (name: string) =>
          `C:\\Temp\\VizcachaIDE\\untitled\\42\\${name.replace(/\.\w+$/, '')}\\${name}`
      }
    }
    const first = await openUntitled(real, 'go')
    expect(first).toBe('C:\\Temp\\VizcachaIDE\\untitled\\42\\main\\main.go')
    expect(await openUntitled(real, 'go')).toBe(
      'C:\\Temp\\VizcachaIDE\\untitled\\42\\main2\\main2.go'
    )
    expect(isUntitled(first)).toBe(true)
    expect(isUntitled('C:\\proyectos\\main.go')).toBe(false)
    expect(codeLanguageOf(first)).toBe('go')
  })

  it('New opens a file of the default language and "New file of" the one chosen', async () => {
    settings.set({ ...(await bridge.settings.get()), defaultCodeLanguage: 'python' })
    await newFile(bridge)
    expect(get(activePath)).toBe('untitled/main.py')
    await newFile(bridge, 'cpp')
    expect(get(activePath)).toBe('untitled/main.cpp')
  })
})
