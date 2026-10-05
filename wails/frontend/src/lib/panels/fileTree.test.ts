import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
import { bridge } from '../bridge'
import { setupI18n } from '../i18n'
import {
  activePath,
  buffers,
  collapsedFolders,
  fileTree,
  openTabs,
  pendingConfirm,
  selectedNode,
  treeEdit
} from '../stores'
import FilesPanel from './FilesPanel.svelte'

beforeAll(() => setupI18n('en'))

beforeEach(async () => {
  openTabs.set([])
  activePath.set(null)
  buffers.set({})
  selectedNode.set(null)
  treeEdit.set(null)
  collapsedFolders.set(new Set())
  fileTree.set(await bridge.files.openFolder())
})

afterEach(() => cleanup())

const nameBox = (): HTMLInputElement => screen.getByRole('textbox', { name: 'Name' })

describe('Files panel header', () => {
  it('has New file, New folder and Refresh buttons when a folder is open', () => {
    render(FilesPanel)
    for (const name of ['New file', 'New folder', 'Refresh']) {
      expect(screen.getByRole('button', { name }).getAttribute('title')).toBe(name)
    }
  })

  it('has no such buttons without a folder', () => {
    fileTree.set(null)
    render(FilesPanel)
    expect(screen.queryByRole('button', { name: 'New file' })).toBeNull()
  })

  it('creates a file with the name typed in the tree, then opens it in a tab', async () => {
    render(FilesPanel)
    await fireEvent.click(screen.getByRole('button', { name: 'New file' }))
    const box = nameBox()
    await fireEvent.input(box, { target: { value: 'saludo' } })
    await fireEvent.keyDown(box, { key: 'Enter' })
    await waitFor(() => expect(get(activePath)).toBe('hola-go/saludo.go'))
    expect(await screen.findByRole('button', { name: 'saludo.go' })).toBeTruthy()
    expect(screen.queryByRole('textbox')).toBeNull()
  })

  it('shows the duplicate error under the box and keeps it open', async () => {
    render(FilesPanel)
    await fireEvent.click(screen.getByRole('button', { name: 'New folder' }))
    const box = nameBox()
    await fireEvent.input(box, { target: { value: 'main.go' } })
    await fireEvent.keyDown(box, { key: 'Enter' })
    expect((await screen.findByRole('alert')).textContent).toBe(
      'A folder named main.go already exists in this folder. Choose another name.'
    )
    expect(box.getAttribute('aria-invalid')).toBe('true')
    await fireEvent.input(box, { target: { value: 'ok' } })
    expect(screen.queryByRole('alert')).toBeNull()
  })

  it('Esc cancels and creates nothing', async () => {
    render(FilesPanel)
    await fireEvent.click(screen.getByRole('button', { name: 'New file' }))
    await fireEvent.keyDown(nameBox(), { key: 'Escape' })
    expect(screen.queryByRole('textbox')).toBeNull()
    expect(get(fileTree)?.children.map((child) => child.name)).not.toContain('.go')
  })
})

describe('Files panel rows', () => {
  it('F2 turns the row into a name box that renames on Enter', async () => {
    render(FilesPanel)
    await fireEvent.keyDown(screen.getByRole('button', { name: 'calculadora.go' }), { key: 'F2' })
    const box = nameBox()
    expect(box.value).toBe('calculadora.go')
    await fireEvent.input(box, { target: { value: 'restas.go' } })
    await fireEvent.keyDown(box, { key: 'Enter' })
    expect(await screen.findByRole('button', { name: 'restas.go' })).toBeTruthy()
    expect(screen.queryByRole('button', { name: 'calculadora.go' })).toBeNull()
  })

  it('Delete asks to move the file to the Recycle Bin', async () => {
    render(FilesPanel)
    await fireEvent.keyDown(screen.getByRole('button', { name: 'go.mod' }), { key: 'Delete' })
    await waitFor(() => expect(get(pendingConfirm)?.messageKey).toBe('confirm.delete'))
    get(pendingConfirm)?.answer('cancel')
    expect(screen.getByRole('button', { name: 'go.mod' })).toBeTruthy()
  })

  it('the root folder ignores F2 and Delete', async () => {
    render(FilesPanel)
    const root = screen.getByRole('button', { name: /hola-go/ })
    await fireEvent.keyDown(root, { key: 'F2' })
    await fireEvent.keyDown(root, { key: 'Delete' })
    expect(screen.queryByRole('textbox')).toBeNull()
    expect(get(pendingConfirm)).toBeNull()
  })

  it('right click opens the menu with the file actions', async () => {
    render(FilesPanel)
    await fireEvent.contextMenu(screen.getByRole('button', { name: 'main.go' }))
    expect(await screen.findByRole('menuitem', { name: 'New file here' })).toBeTruthy()
    for (const name of [/^Rename/, /^Delete/, 'Show in Explorer', 'Copy path', 'New folder']) {
      expect(screen.getByRole('menuitem', { name })).toBeTruthy()
    }
    expect(screen.queryByRole('menuitem', { name: 'Close folder' })).toBeNull()
  })

  it('right click on the open folder offers to close it', async () => {
    const { container } = render(FilesPanel)
    const root = container.querySelector('button.file.dir')
    if (!root) throw new Error('no root folder row')
    await fireEvent.contextMenu(root)
    expect(await screen.findByRole('menuitem', { name: 'Close folder' })).toBeTruthy()
  })
})
