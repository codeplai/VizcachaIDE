import { waitFor } from '@testing-library/svelte'
import { describe, expect, it, vi } from 'vitest'
import { terminalKeyHandler } from './terminalKeys'
import { terminalTheme, watchTheme } from './terminalTheme'

describe('terminal keys', () => {
  const clipboard = () => ({ hasSelection: vi.fn(() => false), copy: vi.fn(), paste: vi.fn() })
  const key = (init: KeyboardEventInit) =>
    new KeyboardEvent('keydown', { cancelable: true, ...init })

  it('sends Ctrl+C to the shell unless text is selected', () => {
    const board = clipboard()
    const handler = terminalKeyHandler(board)
    expect(handler(key({ key: 'c', ctrlKey: true }))).toBe(true)
    expect(board.copy).not.toHaveBeenCalled()
    board.hasSelection.mockReturnValue(true)
    expect(handler(key({ key: 'c', ctrlKey: true }))).toBe(false)
    expect(board.copy).toHaveBeenCalledOnce()
  })

  it('copies with Ctrl+Shift+C, pastes with Ctrl+Shift+V and lets Ctrl+` through', () => {
    const board = clipboard()
    const handler = terminalKeyHandler(board)
    expect(handler(key({ key: 'C', ctrlKey: true, shiftKey: true }))).toBe(false)
    expect(board.copy).toHaveBeenCalledOnce()
    expect(handler(key({ key: 'V', ctrlKey: true, shiftKey: true }))).toBe(false)
    expect(board.paste).toHaveBeenCalledOnce()
    expect(handler(key({ key: '`', ctrlKey: true }))).toBe(false)
    expect(handler(key({ key: 'a' }))).toBe(true)
  })
})

describe('terminal theme', () => {
  it('reads the app tokens and follows the theme switch', async () => {
    const root = document.documentElement
    root.style.setProperty('--win', '#ffffff')
    root.style.setProperty('--ansi-red', '#c2412d')
    expect(terminalTheme().background).toBe('#ffffff')
    expect(terminalTheme().red).toBe('#c2412d')
    const changed = vi.fn()
    const stop = watchTheme(changed)
    root.setAttribute('data-theme', 'dark')
    await waitFor(() => expect(changed).toHaveBeenCalled())
    stop()
    root.removeAttribute('data-theme')
  })
})
