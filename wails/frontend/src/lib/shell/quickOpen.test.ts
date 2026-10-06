import { cleanup, fireEvent, render, screen } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge } from '../bridge'
import { SAMPLE_CALC } from '../bridge/mockData'
import { setupI18n } from '../i18n'
import {
  activePath,
  buffers,
  fileTree,
  openQuickOpen,
  openTabs,
  quickOpenQuery,
  quickOpenVisible,
  settings
} from '../stores'
import QuickOpen from './QuickOpen.svelte'
import { registerShortcuts } from './shortcuts'

beforeAll(() => {
  setupI18n('en')
  Element.prototype.scrollIntoView = () => {} // jsdom has no layout
})

beforeEach(async () => {
  buffers.set({})
  openTabs.set([])
  activePath.set(null)
  fileTree.set(await bridge.files.openFolder())
  quickOpenQuery.set('')
  quickOpenVisible.set(false)
})

afterEach(() => cleanup())

const open = async () => {
  await openQuickOpen()
  render(QuickOpen)
  return (await screen.findByRole('combobox')) as HTMLInputElement
}

describe('Quick open', () => {
  it('is a combobox over a listbox of the folder files with relative paths', async () => {
    const input = await open()
    const list = screen.getByRole('listbox')
    expect(input.getAttribute('aria-controls')).toBe(list.id)
    const options = screen.getAllByRole('option')
    expect(options.length).toBeGreaterThan(2)
    expect(options.some((option) => option.textContent?.includes('main.go'))).toBe(true)
    expect(document.activeElement).toBe(input)
  })

  it('narrows the list while typing and highlights the matched letters', async () => {
    const input = await open()
    await fireEvent.input(input, { target: { value: 'calc' } })
    const options = await screen.findAllByRole('option')
    expect(options).toHaveLength(1)
    expect(options[0]?.textContent).toContain('calculadora.go')
    expect(options[0]?.querySelectorAll('mark').length).toBeGreaterThan(0)
  })

  it('moves with the arrows, wraps around and follows the active option', async () => {
    const input = await open()
    const options = screen.getAllByRole('option')
    expect(options[0]?.getAttribute('aria-selected')).toBe('true')
    await fireEvent.keyDown(input, { key: 'ArrowDown' })
    expect(options[1]?.getAttribute('aria-selected')).toBe('true')
    expect(input.getAttribute('aria-activedescendant')).toBe(options[1]?.id)
    await fireEvent.keyDown(input, { key: 'ArrowUp' })
    await fireEvent.keyDown(input, { key: 'ArrowUp' })
    expect(options[options.length - 1]?.getAttribute('aria-selected')).toBe('true')
  })

  it('Enter opens the active file and closes the palette', async () => {
    const input = await open()
    await fireEvent.input(input, { target: { value: 'calc' } })
    await screen.findByRole('option')
    await fireEvent.keyDown(input, { key: 'Enter' })
    await vi.waitFor(() => expect(get(activePath)).toBe(SAMPLE_CALC))
    expect(get(quickOpenVisible)).toBe(false)
  })

  it('Escape closes it without opening anything', async () => {
    const input = await open()
    await fireEvent.keyDown(input, { key: 'Escape' })
    expect(get(quickOpenVisible)).toBe(false)
    expect(get(activePath)).toBeNull()
  })

  it('lists recently opened files first when nothing is typed', async () => {
    settings.set({ ...(await bridge.settings.get()), recentFiles: [SAMPLE_CALC] })
    await open()
    expect(screen.getAllByRole('option')[0]?.textContent).toContain('calculadora.go')
  })

  it('says so when nothing matches', async () => {
    const input = await open()
    await fireEvent.input(input, { target: { value: 'zzzzqq' } })
    expect(await screen.findByText('No file matches.')).toBeTruthy()
  })
})

describe('Ctrl+P', () => {
  it('opens the palette (Cmd+P too) and needs an open folder', async () => {
    const off = registerShortcuts(bridge)
    await fireEvent.keyDown(window, { key: 'p', ctrlKey: true })
    await vi.waitFor(() => expect(get(quickOpenVisible)).toBe(true))
    quickOpenVisible.set(false)
    await fireEvent.keyDown(window, { key: 'p', metaKey: true })
    await vi.waitFor(() => expect(get(quickOpenVisible)).toBe(true))
    quickOpenVisible.set(false)
    fileTree.set(null)
    await fireEvent.keyDown(window, { key: 'p', ctrlKey: true })
    expect(get(quickOpenVisible)).toBe(false)
    off()
  })
})
