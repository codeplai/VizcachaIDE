import { get } from 'svelte/store'
import { beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { SAMPLE_CALC, SAMPLE_MAIN } from '../bridge/mockData'
import {
  activePath,
  buffers,
  dirty,
  editBuffer,
  fileTree,
  initialSearch,
  openFile,
  openTabs,
  patchSearch,
  pendingConfirm,
  replaceEverywhere,
  replaceInFile,
  replaceMatch,
  runSearch,
  searchState
} from '.'

const setUp = async () => {
  const { bridge } = createMockBridge()
  fileTree.set(await bridge.files.openFolder())
  return bridge
}

beforeEach(() => {
  openTabs.set([])
  activePath.set(null)
  buffers.set({})
  dirty.set({})
  pendingConfirm.set(null)
  searchState.set(initialSearch())
})

describe('search', () => {
  it('finds matches on disk and in the unsaved text of open files', async () => {
    const bridge = await setUp()
    await openFile(bridge, SAMPLE_MAIN)
    editBuffer(SAMPLE_MAIN, 'unsaved restar here\n')
    patchSearch({ query: 'restar' })
    await runSearch(bridge)
    const state = get(searchState)
    expect(state.status).toBe('done')
    expect(state.files.map((file) => file.path).sort()).toEqual([SAMPLE_CALC, SAMPLE_MAIN].sort())
  })

  it('reports an invalid regular expression without asking the backend', async () => {
    const bridge = await setUp()
    patchSearch({ query: '(', options: { caseSensitive: false, wholeWord: false, regex: true } })
    await runSearch(bridge)
    expect(get(searchState).error).toBe('search.invalidPattern')
  })
})

describe('replace', () => {
  it('replaces one match of an open file in its buffer and marks it unsaved', async () => {
    const bridge = await setUp()
    await openFile(bridge, SAMPLE_MAIN)
    patchSearch({ query: 'sumar', replacement: 'add' })
    await runSearch(bridge)
    // calculadora.go calls sumar too (doble): take the match of main.go.
    const mainFile = () => get(searchState).files.find((file) => file.path === SAMPLE_MAIN)
    const first = mainFile()?.matches[0]
    if (!first) throw new Error('no match')
    await replaceMatch(bridge, SAMPLE_MAIN, first)
    const text = get(buffers)[SAMPLE_MAIN] ?? ''
    expect(text).toContain('func add(a, b int)')
    expect(text).toContain('sumar(5, 7)')
    expect(get(dirty)[SAMPLE_MAIN]).toBe(true)
    expect(get(searchState).replaced).toEqual({ files: 1, matches: 1 })
    expect(mainFile()?.matches).toHaveLength(2)
  })

  it('replaces a file that is not open on disk, leaving no unsaved mark', async () => {
    const bridge = await setUp()
    patchSearch({ query: 'restar', replacement: 'subtract' })
    await runSearch(bridge)
    const file = get(searchState).files[0]
    if (!file) throw new Error('no file')
    await replaceInFile(bridge, file)
    expect(await bridge.files.readFile(SAMPLE_CALC)).toContain('func subtract(')
    expect(SAMPLE_CALC in get(buffers)).toBe(false)
    expect(get(searchState).files).toHaveLength(0)
  })

  it('Replace all asks first, naming matches and files, and does nothing when cancelled', async () => {
    const bridge = await setUp()
    await openFile(bridge, SAMPLE_MAIN)
    patchSearch({ query: 'sumar', replacement: 'add' })
    await runSearch(bridge)
    const pending = replaceEverywhere(bridge)
    const request = get(pendingConfirm)
    expect(request?.messageKey).toBe('confirm.replaceInFiles')
    expect(request?.values).toEqual({ count: 4, files: 2 })
    request?.answer('cancel')
    expect(await pending).toBe(false)
    expect(get(buffers)[SAMPLE_MAIN]).toContain('sumar')
    expect(get(dirty)[SAMPLE_MAIN]).toBeUndefined()
  })

  it('Replace all changes open files in buffers and the others on disk', async () => {
    const bridge = await setUp()
    await openFile(bridge, SAMPLE_MAIN)
    patchSearch({
      query: 'int',
      replacement: 'long',
      options: { ...get(searchState).options, wholeWord: true }
    })
    await runSearch(bridge)
    const pending = replaceEverywhere(bridge)
    get(pendingConfirm)?.answer('replace')
    expect(await pending).toBe(true)
    expect(get(buffers)[SAMPLE_MAIN]).toContain('func sumar(a, b long) long')
    expect(get(dirty)[SAMPLE_MAIN]).toBe(true)
    expect(await bridge.files.readFile(SAMPLE_CALC)).toContain('func restar(a, b long) long')
    expect(get(searchState).replaced?.files).toBe(2)
  })
})
