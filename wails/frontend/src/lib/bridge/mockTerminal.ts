// Terminal part of the mock bridge (see mock.ts): a shell that does not exist. It echoes what is
// typed, answers "mock terminal" to every line and ends with `exit`, so `npm run dev` shows the tab.
import type { Emit } from './mockScenarios'
import type { TerminalApi } from './types'

const PROMPT = '$ '
const ENTER = '\r'
const BACKSPACE = '\x7f'

export const mockTerminal = (emit: Emit): TerminalApi => {
  let counter = 0
  const lines = new Map<string, string>()
  const say = (id: string, data: string): void => emit('terminal:output', { id, data })

  const typed = (id: string, character: string): void => {
    const line = lines.get(id) ?? ''
    if (character === BACKSPACE) {
      if (!line) return
      lines.set(id, line.slice(0, -1))
      say(id, '\b \b')
    } else if (character === ENTER) {
      lines.set(id, '')
      if (line.trim() === 'exit') return emit('terminal:exit', { id, exitCode: 0 })
      say(id, `\r\n${line.trim() ? 'mock terminal\r\n' : ''}${PROMPT}`)
    } else {
      lines.set(id, line + character)
      say(id, character)
    }
  }

  return {
    start: async () => {
      const id = `mock-${++counter}`
      lines.set(id, '')
      setTimeout(() => say(id, `Mock terminal (no real shell in the browser).\r\n${PROMPT}`), 0)
      return id
    },
    write: async (id, data) => {
      if (lines.has(id)) [...data].forEach((character) => typed(id, character))
    },
    resize: async () => {},
    close: async (id) => {
      lines.delete(id)
    }
  }
}
