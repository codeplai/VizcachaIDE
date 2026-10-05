import { get } from 'svelte/store'
import { afterEach, describe, expect, it } from 'vitest'
import { languageProfiles } from '../bridge/languageProfiles'
import type { FileNode } from '../domain'
import { profiles } from './codeLanguages'
import { activePath, fileTree } from './files'
import { folderCodeLanguage, workingCodeLanguage } from './workingLanguage'

const file = (path: string): FileNode => ({
  name: path.split('/').at(-1) ?? path,
  path,
  isDir: false,
  children: []
})
const folder = (path: string, children: FileNode[]): FileNode => ({
  name: path.split('/').at(-1) ?? path,
  path,
  isDir: true,
  children
})

afterEach(() => {
  activePath.set(null)
  fileTree.set(null)
})

describe('the language a new file is written in', () => {
  it('is the marker of the folder: Cargo.toml, go.mod, compile_flags.txt', () => {
    expect(
      folderCodeLanguage(
        folder('/p', [file('/p/Cargo.toml'), folder('/p/src', [file('/p/src/main.rs')])]),
        languageProfiles
      )
    ).toBe('rust')
    expect(
      folderCodeLanguage(folder('/p', [file('/p/go.mod'), file('/p/main.go')]), languageProfiles)
    ).toBe('go')
    expect(
      folderCodeLanguage(
        folder('/p', [file('/p/compile_flags.txt'), file('/p/main.cpp')]),
        languageProfiles
      )
    ).toBe('cpp')
  })

  it('is the language most sources have when there is no marker', () => {
    const tree = folder('/p', [
      file('/p/main.py'),
      file('/p/calculo.py'),
      file('/p/notas.go'),
      file('/p/LEEME.txt')
    ])
    expect(folderCodeLanguage(tree, languageProfiles)).toBe('python')
    expect(folderCodeLanguage(folder('/vacia', []), languageProfiles)).toBeNull()
  })

  it('follows the open file first, then the folder, then the default', () => {
    profiles.set(languageProfiles)
    fileTree.set(folder('/p', [file('/p/Cargo.toml')]))
    expect(get(workingCodeLanguage)).toBe('rust')
    activePath.set('/otro/app.py')
    expect(get(workingCodeLanguage)).toBe('python')
    activePath.set(null)
    fileTree.set(null)
    expect(get(workingCodeLanguage)).toBe('go')
  })
})
