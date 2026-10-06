import { describe, expect, it } from 'vitest'
import type { SearchOptions } from '../domainSearch'
import { findInText, isValidQuery, replaceInText } from './matcher'

const plain: SearchOptions = { caseSensitive: false, wholeWord: false, regex: false }
const withOptions = (change: Partial<SearchOptions>): SearchOptions => ({ ...plain, ...change })

describe('findInText', () => {
  it('finds literal text without caring about case, with 1-based columns', () => {
    const found = findInText('Foo foo\nxfoo\n', 'foo', plain)
    expect(found.map((m) => [m.line, m.column, m.length])).toEqual([
      [1, 1, 3],
      [1, 5, 3],
      [2, 2, 3]
    ])
  })

  it('respects match case', () => {
    expect(findInText('Foo foo', 'foo', withOptions({ caseSensitive: true }))).toHaveLength(1)
  })

  it('whole word understands accents and underscores', () => {
    const found = findInText('año años año_x ñaño año', 'año', withOptions({ wholeWord: true }))
    expect(found.map((m) => m.column)).toEqual([1, 21])
  })

  it('treats a literal query as text, not as a pattern', () => {
    expect(findInText('a.b axb', 'a.b', plain)).toHaveLength(1)
    expect(findInText('a.b axb', 'a.b', withOptions({ regex: true }))).toHaveLength(2)
  })

  it('counts UTF-16 units, as the editor does', () => {
    const [match] = findInText('😀 ñandú', 'ñandú', plain)
    expect(match?.column).toBe(4)
  })

  it('validates regular expressions', () => {
    expect(isValidQuery('(', withOptions({ regex: true }))).toBe(false)
    expect(isValidQuery('(', plain)).toBe(true)
    expect(isValidQuery('', plain)).toBe(false)
  })
})

describe('replaceInText', () => {
  it('replaces every literal occurrence and counts them', () => {
    expect(replaceInText('Foo foo FOO', 'foo', plain, 'bar')).toEqual({
      text: 'bar bar bar',
      count: 3
    })
  })

  it('does not expand dollars in a literal replacement', () => {
    expect(replaceInText('a.b', 'a.b', plain, '$1').text).toBe('$1')
  })

  it('expands groups, the whole match and $$ for a regex', () => {
    const options = withOptions({ regex: true, caseSensitive: true })
    const done = replaceInText('x = 1\ny = 22', '(\\w) = (\\d+)', options, '$2 <- $1 [$&] $$ $3')
    expect(done.text).toBe('1 <- x [x = 1] $ $3\n22 <- y [y = 22] $ $3')
    expect(done.count).toBe(2)
  })

  it('replaces only the occurrence at a line and column', () => {
    const done = replaceInText('é cat cat\ncat', 'cat', plain, 'dog', { line: 1, column: 7 })
    expect(done).toEqual({ text: 'é cat dog\ncat', count: 1 })
    expect(replaceInText('cat', 'cat', plain, 'dog', { line: 1, column: 2 }).count).toBe(0)
  })

  it('keeps CRLF line ends and lets $ match before them', () => {
    const done = replaceInText('one\r\ntwo one\r\n', 'one$', withOptions({ regex: true }), '1')
    expect(done.text).toBe('1\r\ntwo 1\r\n')
  })

  it('ignores empty matches', () => {
    expect(replaceInText('abc', 'x*', withOptions({ regex: true }), '-')).toEqual({
      text: 'abc',
      count: 0
    })
  })
})
