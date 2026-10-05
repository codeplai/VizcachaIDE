import { get } from 'svelte/store'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { SAMPLE_CALC, SAMPLE_DIR, SAMPLE_MAIN } from '../bridge/mockData'
import type { FileNode } from '../domain'
import {
  activePath,
  buffers,
  closeFolder,
  connectStores,
  dirty,
  fileTree,
  openFile,
  openTabs,
  pendingConfirm,
  settings
} from '.'

const tree: FileNode = { name: SAMPLE_DIR, path: SAMPLE_DIR, isDir: true, children: [] }
let stop: (() => void) | undefined

beforeEach(() => {
  openTabs.set([])
  activePath.set(null)
  buffers.set({})
  dirty.set({})
})
afterEach(() => stop?.())

const setUp = async () => {
  const { bridge } = createMockBridge()
  stop = await connectStores(bridge)
  await openFile(bridge, SAMPLE_MAIN)
  await openFile(bridge, SAMPLE_CALC)
  openTabs.update((tabs) => [...tabs, 'untitled-1.py'])
  fileTree.set(tree)
  return bridge
}

describe('Close folder', () => {
  it('closes the tabs of the folder, keeps the others and forgets the folder', async () => {
    const bridge = await setUp()
    await closeFolder(bridge)
    expect(get(openTabs)).toEqual(['untitled-1.py'])
    expect(get(fileTree)).toBeNull()
    expect(get(settings)?.lastFolder).toBe('')
  })

  it('keeps the folder open when the learner cancels closing a file with changes', async () => {
    const bridge = await setUp()
    dirty.set({ [SAMPLE_MAIN]: true })
    const closing = closeFolder(bridge)
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(get(pendingConfirm)?.messageKey).toBe('confirm.closeChanges')
    get(pendingConfirm)?.answer('cancel')
    await closing
    expect(get(openTabs)).toContain(SAMPLE_MAIN)
    expect(get(fileTree)).toEqual(tree)
  })
})
