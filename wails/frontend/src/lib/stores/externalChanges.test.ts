import { get } from 'svelte/store'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Bridge } from '../bridge'
import { createEmitter } from '../bridge/emitter'
import { createMockBridge } from '../bridge/mock'
import { Events } from '../events'
import { activePath, buffers, dirty, openTabs } from './files'
import { pendingConfirm } from './confirm'
import { connectExternalChanges } from './externalChanges'

const PATH = 'C:\\proj\\main.go'

const setup = () => {
  const emitter = createEmitter()
  const disk = { text: 'on disk' }
  const watchFiles = vi.fn(async () => {})
  const base = createMockBridge().bridge
  const bridge: Bridge = {
    ...base,
    files: { ...base.files, readFile: async () => disk.text, watchFiles },
    on: emitter.on as Bridge['on']
  }
  const changed = async (): Promise<void> => {
    emitter.emit(Events.fileChanged, { path: PATH })
    await vi.waitFor(() => expect(true).toBe(true))
    await new Promise((resolve) => setTimeout(resolve, 0))
  }
  return { bridge, disk, watchFiles, changed }
}

describe('external changes', () => {
  let stop: (() => void) | undefined

  beforeEach(() => {
    buffers.set({ [PATH]: 'mine' })
    dirty.set({ [PATH]: false })
    openTabs.set([PATH])
    activePath.set(PATH)
    pendingConfirm.set(null)
  })
  afterEach(() => stop?.())

  it('keeps the backend watching the open tabs, without untitled files', () => {
    const { bridge, watchFiles } = setup()
    stop = connectExternalChanges(bridge)
    expect(watchFiles).toHaveBeenLastCalledWith([PATH])
    openTabs.set([PATH, 'untitled/main.go'])
    expect(watchFiles).toHaveBeenCalledTimes(1)
  })

  it('reloads a clean tab silently', async () => {
    const { bridge, changed } = setup()
    stop = connectExternalChanges(bridge)
    await changed()
    expect(get(buffers)[PATH]).toBe('on disk')
    expect(get(pendingConfirm)).toBeNull()
  })

  it('does nothing when the disk text is already in the buffer', async () => {
    const { bridge, disk, changed } = setup()
    disk.text = 'mine'
    dirty.set({ [PATH]: true })
    stop = connectExternalChanges(bridge)
    await changed()
    expect(get(pendingConfirm)).toBeNull()
    expect(get(dirty)[PATH]).toBe(true)
  })

  it('asks before replacing unsaved changes, and reloads on request', async () => {
    const { bridge, changed } = setup()
    dirty.set({ [PATH]: true })
    stop = connectExternalChanges(bridge)
    await changed()
    expect(get(pendingConfirm)?.messageKey).toBe('confirm.reloadFile')
    expect(get(pendingConfirm)?.values).toEqual({ file: 'main.go' })
    get(pendingConfirm)?.answer('reload')
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(get(buffers)[PATH]).toBe('on disk')
    expect(get(dirty)[PATH]).toBe(false)
  })

  it('keeps the unsaved text when the user chooses keep mine', async () => {
    const { bridge, changed } = setup()
    dirty.set({ [PATH]: true })
    stop = connectExternalChanges(bridge)
    await changed()
    get(pendingConfirm)?.answer('keep')
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(get(buffers)[PATH]).toBe('mine')
    expect(get(dirty)[PATH]).toBe(true)
  })
})
