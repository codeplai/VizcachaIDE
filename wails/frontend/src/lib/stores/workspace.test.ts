import { get } from 'svelte/store'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import type { FileNode } from '../domain'
import {
  activePath,
  buffers,
  clampFontSize,
  completeFirstRun,
  confirmCloseChanges,
  connectStores,
  dirty,
  fileTree,
  formatOnSave,
  isUntitled,
  missingToolIn,
  notice,
  openFile,
  openTabs,
  pendingConfirm,
  requestCloseTab,
  resolveProjectPath,
  runActiveFile,
  saveFile,
  settings,
  splitOutputLinks,
  toolOrigin
} from '.'

const tree: FileNode = {
  name: 'proj',
  path: 'C:\\proj',
  isDir: true,
  children: [{ name: 'main.go', path: 'C:\\proj\\main.go', isDir: false, children: [] }]
}

describe('output links', () => {
  it('finds file:line:column places in a Go error', () => {
    const parts = splitOutputLinks('./main.go:11:5: declared and not used: resultado')
    expect(parts[0]).toEqual({
      text: './main.go:11:5',
      location: { file: './main.go', line: 11, column: 5 }
    })
    expect(parts[1]?.text).toBe(': declared and not used: resultado')
  })

  it('finds places in a panic trace and in Windows paths, and ignores plain text', () => {
    expect(splitOutputLinks('\tC:\\proj\\main.go:12 +0x1d')[1]?.location).toEqual({
      file: 'C:\\proj\\main.go',
      line: 12,
      column: 1
    })
    expect(splitOutputLinks('Hola, Go')).toEqual([{ text: 'Hola, Go' }])
  })

  it('maps what Go prints to a file of the open project', () => {
    expect(resolveProjectPath(tree, './main.go')).toBe('C:\\proj\\main.go')
    expect(resolveProjectPath(tree, 'other.go')).toBe('other.go')
    expect(resolveProjectPath(null, 'main.go')).toBe('main.go')
  })
})

describe('settings helpers', () => {
  it('keeps the font size in range and tells where a tool path comes from', () => {
    expect(clampFontSize(3)).toBe(10)
    expect(clampFontSize(99)).toBe(28)
    expect(clampFontSize(Number.NaN)).toBe(10)
    expect(toolOrigin('')).toBe('automatic')
    expect(toolOrigin('C:\\go\\bin\\go.exe')).toBe('custom')
  })

  it('recognises a missing tool in a backend error', () => {
    expect(missingToolIn('exec: "go": executable file not found in %PATH%')).toBe('go')
    expect(missingToolIn('go: not found')).toBe('go')
    expect(missingToolIn('dlv not found')).toBe('delve')
    expect(missingToolIn('something else')).toBeNull()
  })
})

describe('workspace with the mock bridge', () => {
  let stop: (() => void) | undefined
  const mock = createMockBridge()
  const { bridge } = mock

  beforeEach(async () => {
    buffers.set({})
    dirty.set({})
    openTabs.set([])
    activePath.set(null)
    fileTree.set(null)
    notice.set(null)
    stop = await connectStores(bridge)
  })
  afterEach(() => stop?.())

  it('asks before closing a file with unsaved changes', async () => {
    await openFile(bridge, 'hola-go/main.go')
    dirty.set({ 'hola-go/main.go': true })
    const closing = requestCloseTab(bridge, 'hola-go/main.go')
    expect(get(pendingConfirm)?.messageKey).toBe('confirm.closeChanges')
    expect(get(pendingConfirm)?.values).toEqual({ file: 'main.go' })
    get(pendingConfirm)?.answer('cancel')
    await closing
    expect(get(openTabs)).toEqual(['hola-go/main.go'])
  })

  it("saves and closes when the user chooses Save, and closes without saving on Don't save", async () => {
    await openFile(bridge, 'hola-go/main.go')
    const save = vi.spyOn(bridge.files, 'saveFile')
    dirty.set({ 'hola-go/main.go': true })
    const closing = requestCloseTab(bridge, 'hola-go/main.go')
    get(pendingConfirm)?.answer('save')
    await closing
    expect(save).toHaveBeenCalledOnce()
    expect(get(openTabs)).toEqual([])
    save.mockRestore()
  })

  it('answers the escape key with the safe choice', async () => {
    const answer = confirmCloseChanges('a.go')
    expect(get(pendingConfirm)?.dismissId).toBe('cancel')
    get(pendingConfirm)?.answer('cancel')
    expect(await answer).toBe('cancel')
    expect(get(pendingConfirm)).toBeNull()
  })

  it('formats when saving if asked to, and shows a retry message when saving fails', async () => {
    await openFile(bridge, 'hola-go/main.go')
    formatOnSave.set(true)
    const format = vi.spyOn(bridge.run, 'format').mockResolvedValue('formatted')
    expect(await saveFile(bridge, 'hola-go/main.go')).toBe(true)
    expect(get(buffers)['hola-go/main.go']).toBe('formatted')
    format.mockRestore()
    const failing = vi.spyOn(bridge.files, 'saveFile').mockRejectedValue(new Error('disk full'))
    expect(await saveFile(bridge, 'hola-go/main.go')).toBe(false)
    expect(get(notice)?.messageKey).toBe('errors.saveFailed')
    expect(get(notice)?.values).toMatchObject({ file: 'main.go', reason: 'disk full' })
    failing.mockRestore()
  })

  it('opens the example or a blank file and runs it from memory', async () => {
    await bridge.settings.save({ ...(await bridge.settings.get()), firstRun: true })
    expect(get(settings)?.firstRun).toBe(true)
    await completeFirstRun(bridge, 'hello')
    expect(get(settings)?.firstRun).toBe(false)
    const path = get(activePath) ?? ''
    expect(isUntitled(path)).toBe(true)
    expect(get(buffers)[path]).toContain('Hola, Go')
    const runUntitled = vi.spyOn(bridge.run, 'runUntitled')
    await runActiveFile(bridge)
    expect(runUntitled).toHaveBeenCalledWith(get(buffers)[path], [])
  })

  it('shows "Go not installed" with its two buttons when running fails', async () => {
    await completeFirstRun(bridge, 'blank')
    vi.spyOn(bridge.run, 'runUntitled').mockRejectedValue(new Error('go: not found'))
    await runActiveFile(bridge)
    expect(get(notice)?.messageKey).toBe('errors.goNotFound')
    expect(get(notice)?.actions.map((action) => action.labelKey)).toEqual([
      'errors.goNotFoundInstall',
      'errors.goNotFoundChoose'
    ])
  })
})
