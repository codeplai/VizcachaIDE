import { describe, expect, it } from 'vitest'
import type { TextEdit } from './domain'
import { applyChanges, editsToChanges, runesOfUnits, unitsOfRunes } from './textEdits'

const edit = (
  line: number,
  from: number,
  to: number,
  newText: string,
  endLine = line
): TextEdit => ({
  range: {
    start: { file: 'a.go', line, column: from },
    end: { file: 'a.go', line: endLine, column: to }
  },
  newText
})

const apply = (text: string, edits: TextEdit[]): string =>
  applyChanges(text, editsToChanges(text, edits))

describe('rune columns', () => {
  it('counts an emoji once in runes and twice in UTF-16 units', () => {
    expect(unitsOfRunes('😀ab', 1)).toBe(2)
    expect(runesOfUnits('😀ab', 2)).toBe(1)
    expect(unitsOfRunes('año', 2)).toBe(2)
    expect(unitsOfRunes('ab', 9)).toBe(2)
  })
})

describe('editsToChanges', () => {
  it('renames every use, whatever the order of the edits', () => {
    const text = 'func greet() {}\ngreet()\n'
    expect(apply(text, [edit(2, 1, 6, 'hello'), edit(1, 6, 11, 'hello')])).toBe(
      'func hello() {}\nhello()\n'
    )
  })

  it('reads columns in runes, so an emoji before the name does not shift the edit', () => {
    const text = 'print("😀 año", saludar(1))\n'
    // saludar starts at rune column 16 (the emoji is one rune, two UTF-16 units).
    expect(apply(text, [edit(1, 16, 23, 'saludo')])).toBe('print("😀 año", saludo(1))\n')
  })

  it('splits a replace-the-whole-file edit into the lines that really change', () => {
    const text = 'def saludar(n):\n    return n\n\nsaludar(1)\n'
    const whole = edit(1, 1, 1, 'def saludo(n):\n    return n\n\nsaludo(1)\n', 5)
    const changes = editsToChanges(text, [whole])
    expect(changes).toHaveLength(2)
    expect(applyChanges(text, changes)).toBe('def saludo(n):\n    return n\n\nsaludo(1)\n')
  })

  it('drops edits that change nothing', () => {
    expect(editsToChanges('abc\n', [edit(1, 1, 4, 'abc')])).toEqual([])
  })

  it('throws on overlapping edits and on positions outside the text', () => {
    expect(() => editsToChanges('abcdef\n', [edit(1, 1, 4, 'x'), edit(1, 3, 5, 'y')])).toThrow()
    expect(() => editsToChanges('abc\n', [edit(1, 9, 10, 'x')])).toThrow()
    expect(() => editsToChanges('abc\n', [edit(7, 1, 2, 'x')])).toThrow()
  })

  it('lets an edit end at the end of its line (also with CRLF)', () => {
    expect(apply('a\r\nb\r\n', [edit(1, 2, 2, '!'), edit(2, 2, 2, '?')])).toBe('a!\r\nb?\r\n')
  })
})
