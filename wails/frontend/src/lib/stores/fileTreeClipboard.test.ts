import { get } from 'svelte/store'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import type { Bridge } from '../bridge'
import { locale } from '../i18n'
import {
  activePath,
  breakpoints,
  buffers,
  canDrop,
  clearSelection,
  selectEntry,
  selectedPaths,
  collapsedFolders,
  connectStores,
  copyEntries,
  cutEntries,
  dirty,
  dropEntries,
  duplicateEntries,
  fileTree,
  findNode,
  moveEntries,
  moveProblem,
  notice,
  openFile,
  openFolder,
  openTabs,
  panelClipboard,
  pasteEntries,
  startDrag,
  treeEdit
} from '.'

const DIR = 'hola-go'
const MAIN = `${DIR}/main.go`
const CALC = `${DIR}/calculadora.go`
const SRC = `${DIR}/src`

let bridge: Bridge

const names = (path = DIR): string[] =>
  (findNode(get(fileTree), path)?.children.map((child) => child.name) ?? []).sort()

const plain = { toggle: false, range: false }

beforeEach(async () => {
  bridge = createMockBridge().bridge
  fileTree.set(null)
  openTabs.set([])
  activePath.set(null)
  buffers.set({})
  dirty.set({})
  breakpoints.set({})
  notice.set(null)
  treeEdit.set(null)
  panelClipboard.set(null)
  collapsedFolders.set(new Set())
  await locale.set('en')
  await connectStores(bridge)
  await openFolder(bridge)
  await bridge.files.createFolder(SRC)
  fileTree.set(await bridge.files.listTree(DIR))
  clearSelection()
})

afterEach(() => clearSelection())

describe('copy, paste and duplicate', () => {
  it('copies into the selected folder and never overwrites', async () => {
    selectEntry(CALC, plain)
    copyEntries(CALC)
    selectEntry(SRC, plain)
    await pasteEntries(bridge)
    expect(names(SRC)).toEqual(['calculadora.go'])
    await pasteEntries(bridge)
    expect(names(SRC)).toEqual(['calculadora (copy).go', 'calculadora.go'])
    expect(names()).toContain('calculadora.go')
  })

  it('pastes next to the selected file when a file is selected', async () => {
    copyEntries(CALC)
    selectEntry(MAIN, plain)
    await pasteEntries(bridge)
    expect(names()).toContain('calculadora (copy).go')
  })

  it('duplicates in place, and every selected entry', async () => {
    selectEntry(MAIN, plain)
    selectEntry(CALC, { toggle: true, range: false })
    await duplicateEntries(bridge, CALC)
    expect(names()).toEqual(expect.arrayContaining(['main (copy).go', 'calculadora (copy).go']))
    expect(get(selectedPaths)).toHaveLength(2)
  })

  it('refuses to copy a folder into itself', async () => {
    copyEntries(SRC)
    await pasteEntries(bridge, SRC)
    expect(get(notice)?.messageKey).toBe('tree.moveIntoItself')
    expect(names(SRC)).toEqual([])
  })

  it('shows a translated error when the copy fails', async () => {
    bridge.files.copy = async () => {
      throw new Error('access denied')
    }
    copyEntries(CALC)
    await pasteEntries(bridge, SRC)
    expect(get(notice)?.messageKey).toBe('tree.copyEntryFailed')
    expect(get(notice)?.detail).toBe('access denied')
  })
})

describe('moving', () => {
  it('cut and paste moves the file, empties the clipboard, and the open tab follows', async () => {
    await openFile(bridge, CALC)
    breakpoints.set({ [CALC]: [3] })
    cutEntries(CALC)
    await pasteEntries(bridge, SRC)
    expect(names()).not.toContain('calculadora.go')
    expect(names(SRC)).toEqual(['calculadora.go'])
    expect(get(openTabs)).toEqual([`${SRC}/calculadora.go`])
    expect(get(activePath)).toBe(`${SRC}/calculadora.go`)
    expect(Object.keys(get(buffers))).toEqual([`${SRC}/calculadora.go`])
    expect(get(breakpoints)).toEqual({ [`${SRC}/calculadora.go`]: [3] })
    expect(get(panelClipboard)).toBeNull()
  })

  it('a moved folder takes its open files along', async () => {
    await bridge.files.createFolder(`${DIR}/other`)
    await bridge.files.createFile(`${SRC}/a.go`, 'package a')
    fileTree.set(await bridge.files.listTree(DIR))
    await openFile(bridge, `${SRC}/a.go`)
    await moveEntries(bridge, [SRC], `${DIR}/other`)
    expect(get(openTabs)).toEqual([`${DIR}/other/src/a.go`])
  })

  it('does not overwrite: says the name is taken and moves nothing', async () => {
    await bridge.files.createFile(`${SRC}/main.go`, 'x')
    fileTree.set(await bridge.files.listTree(DIR))
    await moveEntries(bridge, [MAIN], SRC)
    expect(get(notice)?.messageKey).toBe('tree.moveExists')
    expect(names()).toContain('main.go')
  })

  it('validates the drop target', () => {
    const tree = get(fileTree)
    expect(moveProblem(tree, SRC, SRC)).toBe('itself')
    expect(moveProblem(tree, SRC, `${SRC}/deeper`)).toBe('itself')
    expect(moveProblem(tree, MAIN, DIR)).toBe('same')
    expect(moveProblem(tree, MAIN, SRC)).toBeNull()
    expect(canDrop([MAIN, CALC], SRC)).toBe(true)
    expect(canDrop([SRC], SRC)).toBe(false)
    expect(canDrop([MAIN], MAIN)).toBe(false)
    expect(canDrop([], SRC)).toBe(false)
  })

  it('drag and drop moves the dragged entries onto a folder', async () => {
    startDrag(MAIN, [])
    await dropEntries(bridge, SRC)
    expect(names(SRC)).toEqual(['main.go'])
  })

  it('dropping a folder on itself is refused with a message', async () => {
    startDrag(SRC, [])
    await dropEntries(bridge, SRC)
    expect(get(notice)?.messageKey).toBe('tree.moveIntoItself')
    expect(names()).toContain('src')
  })
})
