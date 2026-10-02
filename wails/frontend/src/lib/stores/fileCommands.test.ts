import { get } from 'svelte/store'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import type { Bridge } from '../bridge'
import {
  activeNeedsSave,
  activePath,
  buffers,
  closeActive,
  closeAll,
  connectStores,
  dirty,
  editBuffer,
  newFile,
  openFile,
  openFileFromDialog,
  openTabs,
  pendingConfirm,
  saveActiveAs,
  saveActiveFile,
  saveAll,
  settings
} from '.'

let bridge: Bridge

const reset = (): void => {
  buffers.set({})
  dirty.set({})
  openTabs.set([])
  activePath.set(null)
  pendingConfirm.set(null)
}

beforeEach(async () => {
  bridge = createMockBridge().bridge
  reset()
  await connectStores(bridge)
})

const answerNext = async (id: string): Promise<void> => {
  await vi.waitFor(() => expect(get(pendingConfirm)).not.toBeNull())
  get(pendingConfirm)?.answer(id)
}

describe('new file and open file', () => {
  it('new file opens an untitled tab with the blank program', async () => {
    await newFile(bridge)
    expect(get(activePath)).toBe('untitled/main.go')
    expect(get(buffers)['untitled/main.go']).toContain('func main()')
    expect(get(activeNeedsSave)).toBe(true)
  })

  it('open file opens the chosen file', async () => {
    vi.spyOn(bridge.files, 'openFileDialog').mockResolvedValue('hola-go/calculadora.go')
    await openFileFromDialog(bridge)
    expect(get(openTabs)).toEqual(['hola-go/calculadora.go'])
    expect(get(activePath)).toBe('hola-go/calculadora.go')
  })

  it('open file does nothing when the dialog is cancelled', async () => {
    vi.spyOn(bridge.files, 'openFileDialog').mockResolvedValue('')
    await openFileFromDialog(bridge)
    expect(get(openTabs)).toEqual([])
  })
})

describe('saving a new file', () => {
  const startUntitled = async (): Promise<void> => {
    await newFile(bridge)
    editBuffer('untitled/main.go', 'package main\n\nfunc main() { println(1) }\n')
  }

  it('Save as turns the untitled tab into the real file', async () => {
    await startUntitled()
    const save = vi.spyOn(bridge.files, 'saveFile')
    const closeDoc = vi.spyOn(bridge.language, 'closeDocument')
    const openDoc = vi.spyOn(bridge.language, 'openDocument')
    vi.spyOn(bridge.files, 'saveFileDialog').mockResolvedValue('C:\\p\\hola.go')
    await saveActiveFile(bridge)
    expect(save).toHaveBeenCalledWith('C:\\p\\hola.go', expect.stringContaining('println(1)'))
    expect(get(openTabs)).toEqual(['C:\\p\\hola.go'])
    expect(get(activePath)).toBe('C:\\p\\hola.go')
    expect(Object.keys(get(buffers))).toEqual(['C:\\p\\hola.go'])
    expect(get(dirty)['C:\\p\\hola.go']).toBe(false)
    expect(closeDoc).toHaveBeenCalledWith('untitled/main.go')
    expect(openDoc).toHaveBeenCalledWith('C:\\p\\hola.go', expect.any(String))
    expect(get(settings)?.recentFiles?.[0]).toBe('C:\\p\\hola.go')
    expect(get(activeNeedsSave)).toBe(false)
  })

  it('cancelling the dialog keeps the untitled tab', async () => {
    await startUntitled()
    vi.spyOn(bridge.files, 'saveFileDialog').mockResolvedValue('')
    await saveActiveFile(bridge)
    expect(get(openTabs)).toEqual(['untitled/main.go'])
    expect(get(dirty)['untitled/main.go']).toBe(true)
  })

  it('Save as on a real file moves the tab to the new file', async () => {
    await openFile(bridge, 'hola-go/main.go')
    vi.spyOn(bridge.files, 'saveFileDialog').mockResolvedValue('C:\\p\\copia.go')
    await saveActiveAs(bridge)
    expect(get(openTabs)).toEqual(['C:\\p\\copia.go'])
  })

  it('closing an untitled tab with changes and choosing Save opens Save as', async () => {
    await startUntitled()
    vi.spyOn(bridge.files, 'saveFileDialog').mockResolvedValue('C:\\p\\hola.go')
    const closing = closeActive(bridge)
    await answerNext('save')
    await closing
    expect(get(openTabs)).toEqual([])
    expect(get(buffers)).toEqual({})
  })
})

describe('save all and close all', () => {
  it('save all saves every tab with changes', async () => {
    await openFile(bridge, 'hola-go/main.go')
    await openFile(bridge, 'hola-go/calculadora.go')
    editBuffer('hola-go/main.go', 'a')
    editBuffer('hola-go/calculadora.go', 'b')
    const save = vi.spyOn(bridge.files, 'saveFile')
    await saveAll(bridge)
    expect(save).toHaveBeenCalledTimes(2)
    expect(get(dirty)).toEqual({ 'hola-go/main.go': false, 'hola-go/calculadora.go': false })
  })

  it('close all asks for each file with changes and Cancel stops the loop', async () => {
    await openFile(bridge, 'hola-go/main.go')
    await openFile(bridge, 'hola-go/calculadora.go')
    editBuffer('hola-go/calculadora.go', 'b')
    const closing = closeAll(bridge)
    await answerNext('cancel')
    await closing
    expect(get(openTabs)).toEqual(['hola-go/calculadora.go'])
  })

  it('close all closes everything when nothing is pending', async () => {
    await openFile(bridge, 'hola-go/main.go')
    await newFile(bridge)
    await closeAll(bridge)
    expect(get(openTabs)).toEqual([])
  })
})
