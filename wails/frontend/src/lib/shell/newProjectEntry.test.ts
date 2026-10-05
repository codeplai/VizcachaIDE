import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, describe, expect, it } from 'vitest'
import { bridge } from '../bridge'
import { setupI18n } from '../i18n'
import { openDialog } from '../stores'
import FileMenu from './FileMenu.svelte'
import { registerShortcuts } from './shortcuts'

beforeAll(() => setupI18n('en'))

afterEach(() => {
  cleanup()
  openDialog.set(null)
})

describe('New project entry points', () => {
  it('File menu lists "New project…" and it opens the dialog', async () => {
    render(FileMenu)
    await fireEvent.keyDown(screen.getByRole('button', { name: 'File' }), { key: 'Enter' })
    const entry = await screen.findByRole('menuitem', { name: /^New project…/ })
    expect(entry.textContent).toContain('Ctrl+Shift+N')
    await fireEvent.click(entry)
    await waitFor(() => expect(get(openDialog)).toBe('newProject'))
  })

  it('Ctrl+Shift+N opens the dialog', async () => {
    const stop = registerShortcuts(bridge)
    await fireEvent.keyDown(window, { key: 'N', ctrlKey: true, shiftKey: true })
    await waitFor(() => expect(get(openDialog)).toBe('newProject'))
    stop()
  })
})
