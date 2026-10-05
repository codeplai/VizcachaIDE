import { describe, expect, it } from 'vitest'
import { terminalShortcutLabel } from './terminalShortcut'

const layout = (entries: Record<string, string>): Map<string, string> =>
  new Map(Object.entries(entries))

describe('terminalShortcutLabel', () => {
  it('names the US key when the browser does not say the layout', () => {
    expect(terminalShortcutLabel(null, false)).toBe('Ctrl+`')
  })

  it('names the key left of 1 of a US keyboard, with no Ñ', () => {
    expect(terminalShortcutLabel(layout({ Backquote: '`', Semicolon: ';' }), false)).toBe('Ctrl+`')
  })

  it('adds Ctrl+Ñ on a Spanish keyboard', () => {
    expect(terminalShortcutLabel(layout({ Backquote: 'º', Semicolon: 'ñ' }), false)).toBe(
      'Ctrl+º / Ctrl+Ñ'
    )
  })

  it('names the | key of a Latin American keyboard', () => {
    expect(terminalShortcutLabel(layout({ Backquote: '|', Semicolon: 'ñ' }), false)).toBe(
      'Ctrl+| / Ctrl+Ñ'
    )
  })

  it('uses Cmd on macOS', () => {
    expect(terminalShortcutLabel(layout({ Backquote: '`' }), true)).toBe('Cmd+`')
  })
})
