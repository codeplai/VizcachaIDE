import { get } from 'svelte/store'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { SAMPLE_CALC, SAMPLE_MAIN } from '../bridge/mockData'
import type { Bridge } from '../bridge'
import {
  activePath,
  applyFileEdits,
  buffers,
  connectStores,
  dirty,
  notice,
  openFile,
  openTabs,
  prepareRename,
  registerEditorBridge,
  renameSymbol
} from '.'

let bridge: Bridge

beforeEach(async () => {
  bridge = createMockBridge().bridge
  buffers.set({})
  dirty.set({})
  openTabs.set([])
  activePath.set(null)
  notice.set(null)
  await connectStores(bridge)
})

// "func sumar(a, b int) int {" is line 5 of main.go; sumar is called on lines 11 and 17 too, and
// by calculadora.go.
const atSumar = { file: SAMPLE_MAIN, line: 5, column: 7 }

describe('rename', () => {
  it('edits the open files in their buffers and the closed ones on disk, and says how much', async () => {
    await openFile(bridge, SAMPLE_MAIN) // calculadora.go stays closed
    expect(await renameSymbol(bridge, atSumar, 'sumarDos')).toBe(true)
    const main = get(buffers)[SAMPLE_MAIN] ?? ''
    expect(main).toContain('func sumarDos(a, b int) int {')
    expect(main).toContain('resultado := sumarDos(5, 7)')
    expect(main).not.toContain('sumar(')
    expect(get(dirty)[SAMPLE_MAIN]).toBe(true)
    expect(await bridge.files.readFile(SAMPLE_CALC)).toContain('return sumarDos(a, a)')
    expect(get(dirty)[SAMPLE_CALC]).toBeUndefined() // not open: nothing looks unsaved
    expect(get(notice)).toMatchObject({
      messageKey: 'refactor.renamed',
      values: { places: 4, files: 2 },
      tone: 'info'
    })
  })

  it('edits every open file in its buffer and keeps the language server in step', async () => {
    await openFile(bridge, SAMPLE_CALC)
    await openFile(bridge, SAMPLE_MAIN) // main.go is the active tab
    const change = vi.spyOn(bridge.language, 'changeDocument')
    await renameSymbol(bridge, atSumar, 'total')
    expect(get(buffers)[SAMPLE_CALC]).toContain('return total(a, a)')
    expect(get(dirty)).toMatchObject({ [SAMPLE_MAIN]: true, [SAMPLE_CALC]: true })
    expect(change).toHaveBeenCalledWith(SAMPLE_CALC, expect.stringContaining('total(a, a)'), 0)
  })

  it('lets the editor apply the changes of a file it holds, as one undoable edit', async () => {
    await openFile(bridge, SAMPLE_MAIN)
    const applyChanges = vi.fn(() => true)
    const unregister = registerEditorBridge({
      applyChanges,
      renameSymbol: () => {},
      findReferences: () => {}
    })
    await renameSymbol(bridge, atSumar, 'total')
    unregister()
    expect(applyChanges).toHaveBeenCalledTimes(1)
    const [path, changes] = applyChanges.mock.calls[0] as unknown as [string, unknown[]]
    expect(path).toBe(SAMPLE_MAIN)
    expect(changes).toHaveLength(3)
    // The active file is the editor's to update (it reports the edit back as a change).
    expect(get(buffers)[SAMPLE_MAIN]).toContain('sumar(')
  })

  it('finds the open file even when the server spells its path differently', async () => {
    const tab = 'D:\\proyecto\\main.go'
    buffers.set({ [tab]: 'package main\n' })
    const file = 'd:/proyecto/main.go' // a URI gives the drive letter in lower case
    const range = { start: { file, line: 1, column: 1 }, end: { file, line: 1, column: 8 } }
    await applyFileEdits(bridge, [{ file, edits: [{ range, newText: 'PACKAGE' }] }])
    expect(get(buffers)).toEqual({ [tab]: 'PACKAGE main\n' })
  })

  it('says why a keyword cannot be renamed and changes nothing', async () => {
    await openFile(bridge, SAMPLE_MAIN)
    const keyword = { file: SAMPLE_MAIN, line: 5, column: 2 } // "func"
    expect(await prepareRename(bridge, keyword)).toBeNull()
    expect(get(notice)?.messageKey).toBe('refactor.notRenameable')
    expect(get(buffers)[SAMPLE_MAIN]).toContain('func sumar(')
  })

  it('gives the allowed target (name and range) to the editor', async () => {
    await openFile(bridge, SAMPLE_MAIN)
    const target = await prepareRename(bridge, atSumar)
    expect(target?.placeholder).toBe('sumar')
    expect(target?.range?.start).toEqual({ file: SAMPLE_MAIN, line: 5, column: 6 })
    expect(get(notice)).toBeNull()
  })

  it('shows a translated message when the server cannot rename (Python without rope)', async () => {
    vi.spyOn(bridge.language, 'rename').mockResolvedValue({ refusal: 'unsupported', files: [] })
    expect(await renameSymbol(bridge, { file: 'proyecto/main.py', line: 1, column: 5 }, 'x')).toBe(
      false
    )
    expect(get(notice)).toMatchObject({
      messageKey: 'refactor.unsupportedPython',
      detail: 'pip install rope'
    })
  })

  it('reports a server failure with its reason', async () => {
    vi.spyOn(bridge.language, 'rename').mockResolvedValue({
      refusal: 'failed',
      detail: 'invalid name',
      files: []
    })
    expect(await renameSymbol(bridge, atSumar, '1x')).toBe(false)
    expect(get(notice)).toMatchObject({
      messageKey: 'refactor.failed',
      values: { reason: 'invalid name' }
    })
  })

  it('reports edits that cannot be applied instead of throwing', async () => {
    await openFile(bridge, SAMPLE_MAIN)
    vi.spyOn(bridge.language, 'rename').mockResolvedValue({
      refusal: '',
      files: [
        {
          file: SAMPLE_MAIN,
          edits: [
            {
              range: {
                start: { file: SAMPLE_MAIN, line: 99, column: 1 },
                end: { file: SAMPLE_MAIN, line: 99, column: 2 }
              },
              newText: 'x'
            }
          ]
        }
      ]
    })
    expect(await renameSymbol(bridge, atSumar, 'x')).toBe(false)
    expect(get(notice)?.messageKey).toBe('refactor.failed')
  })
})
