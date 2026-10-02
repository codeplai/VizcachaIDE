import { cleanup, render, screen } from '@testing-library/svelte'
import { afterEach, beforeAll, describe, expect, it } from 'vitest'
import { setupI18n } from '../i18n'
import OutputPanel from '../panels/OutputPanel.svelte'
import { addChunk, parseAnsi, resetRun, runLines, styledSegments, type RawRunLine } from '.'

type Chunk = ['stdout' | 'stderr', string]

const feed = (...chunks: Chunk[]) => {
  let state: ReturnType<typeof addChunk> = { lines: [], tail: null }
  for (const [stream, text] of chunks) state = addChunk(state.lines, state.tail, stream, text)
  return state.lines.map((line) => line.text)
}

beforeAll(() => setupI18n('en'))
afterEach(() => {
  cleanup()
  resetRun()
})

describe('ANSI colors', () => {
  it('turns SGR codes into styled pieces', () => {
    const pieces = parseAnsi('ok \u001b[1;31mFAIL\u001b[0m done')
    expect(pieces.map((piece) => piece.text)).toEqual(['ok ', 'FAIL', ' done'])
    expect(pieces[1]).toMatchObject({ color: 'var(--ansi-red)', decorations: ['bold'] })
    expect(pieces[0]?.color).toBeUndefined()
  })

  it('understands the 256 palette and true color without leaking old colors', () => {
    const [truecolor, palette] = parseAnsi('\u001b[38;2;1;2;3mA\u001b[38;5;196mB\u001b[0m')
    expect(truecolor?.color).toBe('rgb(1, 2, 3)')
    expect(palette?.color).toBe('rgb(255, 0, 0)')
  })

  it('drops escape sequences that are not colors', () => {
    const text = parseAnsi('a\u001b[2Kb\u001b[?25lc')
      .map((piece) => piece.text)
      .join('')
    expect(text).toBe('abc')
  })

  it('keeps the file.go:line:column links inside colored text', () => {
    const segments = styledSegments('\u001b[31m./main.go:4:2: boom\u001b[0m')
    const link = segments.find((segment) => segment.location)
    expect(link?.location).toEqual({ file: './main.go', line: 4, column: 2 })
    expect(link?.color).toBe('var(--ansi-red)')
  })
})

describe('carriage return', () => {
  it('rewrites the line, so a progress bar stays on one line', () => {
    expect(feed(['stdout', '10%\r50%\r100%\ndone\n'])).toEqual(['100%', 'done'])
  })

  it('works across chunks and keeps the cursor at the start between them', () => {
    expect(
      feed(['stdout', 'loading 1/3\r'], ['stdout', 'loading 2/3\r'], ['stdout', 'ok\n'])
    ).toEqual(['okading 2/3'])
  })

  it('replaces the line when the new text is longer or colored', () => {
    expect(feed(['stdout', 'ab\rabcdef\n'])).toEqual(['abcdef'])
    expect(feed(['stdout', '\u001b[32m50%\u001b[0m\r\u001b[32m100%\u001b[0m\n'])).toEqual([
      '\u001b[32m100%\u001b[0m'
    ])
  })

  it('still treats CRLF as a normal line break', () => {
    expect(feed(['stdout', 'one\r\ntwo\r\n'])).toEqual(['one', 'two'])
  })

  it('does not mix the two streams', () => {
    expect(feed(['stdout', 'a\r'], ['stderr', 'b\n'])).toEqual(['a', 'b'])
  })
})

describe('Output panel', () => {
  const show = (text: string) => {
    const lines: RawRunLine[] = [{ kind: 'stdout', text }]
    runLines.set(lines)
    return render(OutputPanel)
  }

  it('shows colored text with styles and never interprets HTML', () => {
    const { container } = show('\u001b[1;32mPASS\u001b[0m <b>not bold</b>')
    const colored = screen.getByText('PASS')
    expect(colored.tagName).toBe('SPAN')
    expect(colored.style.color).toBe('var(--ansi-green)')
    expect(colored.className).toContain('ansi-bold')
    expect(container.querySelector('b')).toBeNull()
    expect(container.textContent).toContain('<b>not bold</b>')
  })

  it('keeps the places clickable', () => {
    show('\u001b[31m./main.go:7:3: oops\u001b[0m')
    expect(screen.getByRole('button', { name: /main\.go:7:3/ })).toBeTruthy()
  })
})
