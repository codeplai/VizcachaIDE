import { cleanup, fireEvent, render, screen } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import type { Diagnostic } from '../domain'
import { setupI18n } from '../i18n'
import {
  activePath,
  assistantOpen,
  diagnosticsByFile,
  selectSidebarView,
  sidebarOpen,
  sidebarView,
  startDebugging
} from '../stores'
import PanelToggles from './PanelToggles.svelte'
import { registerShortcuts } from './shortcuts'

const problem: Diagnostic = {
  location: { file: 'main.go', line: 3, column: 2 },
  severity: 'error',
  message: 'undefined: x',
  rawText: 'main.go:3:2: undefined: x',
  source: 'compiler',
  code: '',
  end: null
}

beforeAll(() => setupI18n('en'))

beforeEach(() => {
  sidebarOpen.set(true)
  assistantOpen.set(true)
  sidebarView.set('files')
})

afterEach(() => {
  cleanup()
  diagnosticsByFile.set({})
  activePath.set(null)
})

describe('hiding the side panel and the Assistant', () => {
  it('a click on the active rail view hides the side panel; another view shows it', () => {
    selectSidebarView('files')
    expect(get(sidebarOpen)).toBe(false)
    selectSidebarView('outline')
    expect(get(sidebarOpen)).toBe(true)
    expect(get(sidebarView)).toBe('outline')
  })

  it('the title bar buttons toggle each panel', async () => {
    render(PanelToggles)
    await fireEvent.click(screen.getByRole('button', { name: /side panel/ }))
    expect(get(sidebarOpen)).toBe(false)
    await fireEvent.click(screen.getByRole('button', { name: /Assistant/ }))
    expect(get(assistantOpen)).toBe(false)
  })

  it('shows the number of problems on the Assistant button only while it is hidden', async () => {
    activePath.set('main.go')
    diagnosticsByFile.set({ 'main.go': [problem] })
    render(PanelToggles)
    expect(document.querySelector('.badge')).toBeNull()
    assistantOpen.set(false)
    expect((await screen.findByText('1')).classList.contains('badge')).toBe(true)
  })

  it('Ctrl+B toggles the side panel and Ctrl+Alt+B the Assistant', () => {
    const { bridge } = createMockBridge()
    const off = registerShortcuts(bridge)
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'b', ctrlKey: true }))
    expect(get(sidebarOpen)).toBe(false)
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'b', ctrlKey: true, altKey: true }))
    expect(get(assistantOpen)).toBe(false)
    off()
  })

  it('starting to debug opens the Assistant again (variables and call stack live there)', async () => {
    const { bridge } = createMockBridge()
    assistantOpen.set(false)
    activePath.set('hola-go/main.go')
    await startDebugging(bridge)
    expect(get(assistantOpen)).toBe(true)
  })
})
