import { get } from 'svelte/store'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { locale } from '../i18n'
import {
  clearSelection,
  collapsedFolders,
  connectStores,
  copyName,
  fileTree,
  openFolder,
  selectEntry,
  selectedPaths,
  selectionAfterClick,
  topLevelPaths
} from '.'

const DIR = 'hola-go'
const MAIN = `${DIR}/main.go`
const CALC = `${DIR}/calculadora.go`
const SRC = `${DIR}/src`
const plain = { toggle: false, range: false }

beforeEach(async () => {
  fileTree.set(null)
  collapsedFolders.set(new Set())
  await locale.set('en')
  await connectStores(createMockBridge().bridge)
  await openFolder(createMockBridge().bridge)
  clearSelection()
})

afterEach(() => clearSelection())

describe('selection rules', () => {
  const visible = ['r', 'a', 'b', 'c', 'd']

  it('a plain click selects one row', () => {
    expect(selectionAfterClick(['a', 'b'], 'a', 'c', visible, plain).paths).toEqual(['c'])
  })

  it('Ctrl+click toggles a row', () => {
    const toggle = { toggle: true, range: false }
    expect(selectionAfterClick(['a'], 'a', 'c', visible, toggle).paths).toEqual(['a', 'c'])
    expect(selectionAfterClick(['a', 'c'], 'c', 'a', visible, toggle).paths).toEqual(['c'])
  })

  it('Shift+click selects the visible rows between the start and the click, either way', () => {
    const range = { toggle: false, range: true }
    expect(selectionAfterClick(['b'], 'b', 'd', visible, range).paths).toEqual(['b', 'c', 'd'])
    expect(selectionAfterClick(['d'], 'd', 'b', visible, range).paths).toEqual(['b', 'c', 'd'])
  })

  it('works on the real tree and keeps the focused row', () => {
    selectEntry(MAIN, plain)
    selectEntry(CALC, { toggle: true, range: false })
    expect(get(selectedPaths)).toEqual([MAIN, CALC])
    clearSelection()
    expect(get(selectedPaths)).toEqual([])
  })

  it('leaves out the root and what is inside another selected folder', () => {
    expect(topLevelPaths([DIR, SRC, `${SRC}/a.go`, MAIN], DIR)).toEqual([SRC, MAIN])
  })
})

describe('names of copies', () => {
  it('numbers the copies and keeps the extension (English)', () => {
    expect(copyName('a.txt', false, ['a.txt'], true)).toBe('a (copy).txt')
    expect(copyName('a.txt', false, ['a.txt', 'a (copy).txt'], true)).toBe('a (copy 2).txt')
    expect(copyName('src', true, ['src'], true)).toBe('src (copy)')
    expect(copyName('.env', false, ['.env'], true)).toBe('.env (copy)')
  })

  it('follows the UI language (Spanish)', async () => {
    await locale.set('es')
    expect(copyName('canción.go', false, ['canción.go', 'canción (copia).go'], true)).toBe(
      'canción (copia 2).go'
    )
  })

  it('keeps the plain name in another folder unless it is taken, ignoring letter case', () => {
    expect(copyName('a.txt', false, ['b.txt'], false)).toBe('a.txt')
    expect(copyName('a.txt', false, ['A.TXT'], false)).toBe('a (copy).txt')
  })
})
