import { get } from 'svelte/store'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import type { Bridge } from '../bridge'
import {
  activePath,
  buffers,
  cancelEdit,
  collapsedFolders,
  commitEdit,
  connectStores,
  dirty,
  editBuffer,
  fileTree,
  findNode,
  notice,
  openFile,
  openFolder,
  openTabs,
  rememberRecentFile,
  settings,
  startCreate,
  startRename,
  treeEdit
} from '.'

const DIR = 'hola-go'
const MAIN = `${DIR}/main.go`
const CALC = `${DIR}/calculadora.go`

let bridge: Bridge

const namesInRoot = (): string[] =>
  (get(fileTree)?.children.map((child) => child.name) ?? []).sort()

beforeEach(async () => {
  bridge = createMockBridge().bridge
  fileTree.set(null)
  openTabs.set([])
  activePath.set(null)
  buffers.set({})
  dirty.set({})
  notice.set(null)
  treeEdit.set(null)
  collapsedFolders.set(new Set())
  await connectStores(bridge)
  await openFolder(bridge)
})

afterEach(() => {
  treeEdit.set(null)
})

describe('creating files and folders', () => {
  it('creates a file, adds .go when there is no extension, and opens it with the template', async () => {
    startCreate('file', DIR)
    await commitEdit(bridge, 'hola')
    expect(namesInRoot()).toContain('hola.go')
    expect(get(treeEdit)).toBeNull()
    expect(get(activePath)).toBe(`${DIR}/hola.go`)
    expect(get(buffers)[`${DIR}/hola.go`]).toBe('package main\n\nfunc main() {\n}\n')
  })

  it('keeps another extension and does not open or fill it', async () => {
    startCreate('file', DIR)
    await commitEdit(bridge, 'notes.txt')
    expect(namesInRoot()).toContain('notes.txt')
    expect(get(openTabs)).toEqual([])
  })

  it('creates a folder', async () => {
    startCreate('folder', DIR)
    await commitEdit(bridge, 'utils')
    const folder = findNode(get(fileTree), `${DIR}/utils`)
    expect(folder?.isDir).toBe(true)
  })

  it('shows an inline error for an empty name and does nothing', async () => {
    startCreate('file', DIR)
    await commitEdit(bridge, '   ')
    expect(get(treeEdit)?.error?.key).toBe('tree.nameEmpty')
    expect(namesInRoot()).toEqual(['calculadora.go', 'go.mod', 'main.go'])
  })

  it.each([
    'a/b.go',
    'a\\b.go',
    'a:b.go',
    'a*.go',
    'a?.go',
    'a"b.go',
    'a<b.go',
    'a>b.go',
    'a|b.go'
  ])('rejects %s', async (name) => {
    startCreate('file', DIR)
    await commitEdit(bridge, name)
    expect(get(treeEdit)?.error?.key).toBe('tree.nameInvalid')
  })

  it('rejects a duplicate name (ignoring letter case) and keeps the box open', async () => {
    startCreate('file', DIR)
    await commitEdit(bridge, 'MAIN.go')
    expect(get(treeEdit)?.error).toEqual({ key: 'tree.nameExists', values: { name: 'MAIN.go' } })
  })

  it('checks the duplicate after adding .go', async () => {
    startCreate('file', DIR)
    await commitEdit(bridge, 'main')
    expect(get(treeEdit)?.error?.key).toBe('tree.nameExists')
  })

  it('cancel closes the box and changes nothing', async () => {
    startCreate('file', DIR)
    cancelEdit()
    expect(get(treeEdit)).toBeNull()
    expect(namesInRoot()).toEqual(['calculadora.go', 'go.mod', 'main.go'])
  })

  it('expands the folder where the new entry will appear', () => {
    collapsedFolders.set(new Set([DIR]))
    startCreate('file', DIR)
    expect(get(collapsedFolders).has(DIR)).toBe(false)
  })
})

describe('renaming', () => {
  it('renames a file in the tree', async () => {
    startRename(CALC)
    await commitEdit(bridge, 'matematicas.go')
    expect(namesInRoot()).toContain('matematicas.go')
    expect(namesInRoot()).not.toContain('calculadora.go')
  })

  it('moves the open tab, buffer, dirty mark, active file and recent files', async () => {
    await openFile(bridge, CALC)
    await rememberRecentFile(bridge, CALC)
    editBuffer(CALC, 'package main // changed')
    startRename(CALC)
    await commitEdit(bridge, 'matematicas.go')
    const renamed = `${DIR}/matematicas.go`
    expect(get(openTabs)).toEqual([renamed])
    expect(get(activePath)).toBe(renamed)
    expect(get(buffers)).toEqual({ [renamed]: 'package main // changed' })
    expect(get(dirty)).toEqual({ [renamed]: true })
    expect(get(settings)?.recentFiles).toEqual([renamed])
  })

  it('keeps the name when it did not change and does not complain', async () => {
    startRename(CALC)
    await commitEdit(bridge, 'calculadora.go')
    expect(get(treeEdit)).toBeNull()
    expect(get(notice)).toBeNull()
  })

  it('does not rename onto an existing file', async () => {
    startRename(CALC)
    await commitEdit(bridge, 'main.go')
    expect(get(treeEdit)?.error?.key).toBe('tree.nameExists')
  })

  it('allows changing only the letter case', async () => {
    startRename(CALC)
    await commitEdit(bridge, 'Calculadora.go')
    expect(namesInRoot()).toContain('Calculadora.go')
  })

  it('cancel does not rename', async () => {
    startRename(CALC)
    cancelEdit()
    expect(namesInRoot()).toContain('calculadora.go')
  })

  it('updates every open tab under a renamed folder', async () => {
    startCreate('folder', DIR)
    await commitEdit(bridge, 'src')
    await bridge.files.createFile(`${DIR}/src/a.go`, 'package a')
    await bridge.files.createFile(`${DIR}/src/b.go`, 'package b')
    await openFile(bridge, `${DIR}/src/a.go`)
    await openFile(bridge, `${DIR}/src/b.go`)
    await openFile(bridge, MAIN)
    startRename(`${DIR}/src`)
    await commitEdit(bridge, 'lib')
    expect(get(openTabs)).toEqual([`${DIR}/lib/a.go`, `${DIR}/lib/b.go`, MAIN])
    expect(Object.keys(get(buffers)).sort()).toEqual(
      [`${DIR}/lib/a.go`, `${DIR}/lib/b.go`, MAIN].sort()
    )
    expect(await bridge.files.readFile(`${DIR}/lib/a.go`)).toBe('package a')
  })

  it('cannot rename the root folder', () => {
    startRename(DIR)
    expect(get(treeEdit)).toBeNull()
  })
})
