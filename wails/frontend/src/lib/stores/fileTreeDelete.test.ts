import { get } from 'svelte/store'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import type { Bridge } from '../bridge'
import {
  activePath,
  buffers,
  collapsedFolders,
  commitEdit,
  connectStores,
  deleteEntry,
  dirty,
  editBuffer,
  fileTree,
  findNode,
  notice,
  openFile,
  openFolder,
  openTabs,
  pendingConfirm,
  startCreate,
  treeEdit
} from '.'

const DIR = 'hola-go'
const MAIN = `${DIR}/main.go`
const CALC = `${DIR}/calculadora.go`

let bridge: Bridge

const namesInRoot = (): string[] =>
  (get(fileTree)?.children.map((child) => child.name) ?? []).sort()

const listeners: (() => void)[] = []

/** Answers the next confirmation with the given button (and remembers the question). */
const answerWith = (id: string): { asked: () => string | undefined } => {
  let asked: string | undefined
  let answered = false
  listeners.push(
    pendingConfirm.subscribe((request) => {
      if (!request || answered) return
      answered = true
      asked = request.messageKey
      queueMicrotask(() => request.answer(id))
    })
  )
  return { asked: () => asked }
}

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
  listeners.splice(0).forEach((off) => off())
})

describe('deleting', () => {
  it('asks first and does nothing when you cancel', async () => {
    const answer = answerWith('cancel')
    await deleteEntry(bridge, CALC)
    expect(answer.asked()).toBe('confirm.delete')
    expect(namesInRoot()).toContain('calculadora.go')
  })

  it('moves the file to the Recycle Bin and closes its tab', async () => {
    await openFile(bridge, CALC)
    answerWith('delete')
    await deleteEntry(bridge, CALC)
    expect(namesInRoot()).not.toContain('calculadora.go')
    expect(get(openTabs)).toEqual([])
    expect(get(buffers)).toEqual({})
  })

  it('warns when the file has unsaved changes', async () => {
    await openFile(bridge, CALC)
    editBuffer(CALC, 'x')
    const answer = answerWith('cancel')
    await deleteEntry(bridge, CALC)
    expect(answer.asked()).toBe('confirm.deleteUnsaved')
    expect(get(openTabs)).toEqual([CALC])
  })

  it('closes every tab of a deleted folder and says how many files it has', async () => {
    startCreate('folder', DIR)
    await commitEdit(bridge, 'src')
    await bridge.files.createFile(`${DIR}/src/a.go`, 'package a')
    await openFile(bridge, `${DIR}/src/a.go`)
    await openFile(bridge, MAIN)
    const answer = answerWith('delete')
    await deleteEntry(bridge, `${DIR}/src`)
    expect(answer.asked()).toBe('confirm.deleteFolder')
    expect(get(openTabs)).toEqual([MAIN])
    expect(findNode(get(fileTree), `${DIR}/src`)).toBeNull()
  })

  it('never deletes the root folder', async () => {
    const answer = answerWith('delete')
    await deleteEntry(bridge, DIR)
    expect(answer.asked()).toBeUndefined()
  })

  it('shows a notice and keeps the file when the Recycle Bin fails', async () => {
    bridge.files.moveToTrash = async () => {
      throw new Error('bin unavailable')
    }
    answerWith('delete')
    await deleteEntry(bridge, CALC)
    expect(get(notice)?.messageKey).toBe('tree.deleteFailed')
    expect(namesInRoot()).toContain('calculadora.go')
  })
})
