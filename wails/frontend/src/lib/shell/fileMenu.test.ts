import { cleanup, fireEvent, render, screen } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge } from '../bridge'
import { setupI18n } from '../i18n'
import { activePath, buffers, dirty, openTabs, pendingConfirm } from '../stores'
import FileButtons from './FileButtons.svelte'
import FileMenu from './FileMenu.svelte'
import { registerShortcuts } from './shortcuts'

beforeAll(() => setupI18n('en'))

beforeEach(() => {
  buffers.set({})
  dirty.set({})
  openTabs.set([])
  activePath.set(null)
  pendingConfirm.set(null)
})

afterEach(() => cleanup())

describe('File menu', () => {
  it('lists the file actions with their shortcuts', async () => {
    render(FileMenu)
    await fireEvent.keyDown(screen.getByRole('button', { name: 'File' }), { key: 'Enter' })
    for (const name of [
      /^New file/,
      /^Open file…/,
      /^Open folder…/,
      /^Save\s*(Ctrl\+S)?$/,
      /^Save as…/,
      /^Save all/,
      /^Close\s*(Ctrl\+W)?$/,
      /^Close all/
    ]) {
      expect(await screen.findByRole('menuitem', { name })).toBeTruthy()
    }
    expect(screen.getByRole('menuitem', { name: /^New file/ }).textContent).toContain('Ctrl+N')
    expect(screen.getByText('Open recent')).toBeTruthy()
  })

  it('New file opens an untitled tab', async () => {
    render(FileMenu)
    await fireEvent.keyDown(screen.getByRole('button', { name: 'File' }), { key: 'Enter' })
    await fireEvent.click(await screen.findByRole('menuitem', { name: /^New file/ }))
    await vi.waitFor(() => expect(get(activePath)).toBe('untitled/main.go'))
  })
})

describe('file buttons', () => {
  it('Save is disabled without changes and enabled for a new file', async () => {
    render(FileButtons)
    const save = screen.getByRole('button', { name: 'Save' }) as HTMLButtonElement
    expect(save.disabled).toBe(true)
    await fireEvent.click(screen.getByRole('button', { name: 'New file' }))
    await vi.waitFor(() => expect(save.disabled).toBe(false))
    expect(screen.getByRole('button', { name: 'Open file…' }).title).toContain('Ctrl+O')
  })
})

describe('file shortcuts', () => {
  it('Ctrl+N, Ctrl+O, Ctrl+Shift+S and Ctrl+W are handled and prevented', async () => {
    const off = registerShortcuts(bridge)
    const open = vi.spyOn(bridge.files, 'openFileDialog').mockResolvedValue('')
    const saveDialog = vi.spyOn(bridge.files, 'saveFileDialog').mockResolvedValue('')
    const press = (init: KeyboardEventInit): boolean =>
      window.dispatchEvent(new KeyboardEvent('keydown', { cancelable: true, ...init }))

    expect(press({ key: 'n', ctrlKey: true })).toBe(false)
    await vi.waitFor(() => expect(get(activePath)).toBe('untitled/main.go'))
    expect(press({ key: 'o', ctrlKey: true })).toBe(false)
    expect(open).toHaveBeenCalled()
    expect(press({ key: 'S', ctrlKey: true, shiftKey: true })).toBe(false)
    await vi.waitFor(() => expect(saveDialog).toHaveBeenCalled())
    expect(press({ key: 'w', metaKey: true })).toBe(false)
    await vi.waitFor(() => expect(get(openTabs)).toEqual([]))
    off()
  })
})
