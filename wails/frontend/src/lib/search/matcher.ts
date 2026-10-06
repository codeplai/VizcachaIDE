// Finds and replaces text the way the backend does (internal/app/search_match.go), for the files
// that are open in the editor (their text may not be saved yet). Both sides must agree on the
// options, the columns (UTF-16) and the `$1`, `$&`, `$$` expansion of a regex replacement.
import type { SearchMatch, SearchOptions } from '../domainSearch'

const WORD = '[\\p{L}\\p{N}_]'
const WORD_FALLBACK = '[A-Za-z0-9_\\u00C0-\\u024F]'

/** Compiles the search; throws a SyntaxError when the regular expression is not valid. */
export const buildRegExp = (query: string, options: SearchOptions): RegExp => {
  const source = options.regex ? query : query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const flags = options.caseSensitive ? 'g' : 'gi'
  const wrap = (word: string): string =>
    options.wholeWord ? `(?<!${word})(?:${source})(?!${word})` : source
  try {
    return new RegExp(wrap(WORD), `${flags}u`)
  } catch (error) {
    if (!options.regex) throw error
    return new RegExp(wrap(WORD_FALLBACK), flags) // patterns that only work without the u flag
  }
}

/** True when the query is a usable search (non-empty and, for a regex, valid). */
export const isValidQuery = (query: string, options: SearchOptions): boolean => {
  if (query === '') return false
  try {
    buildRegExp(query, options)
    return true
  } catch {
    return false
  }
}

const MAX_TEXT = 300

interface Found {
  start: number
  end: number
  match: RegExpExecArray
}

const occurrences = (re: RegExp, body: string): Found[] => {
  const found: Found[] = []
  re.lastIndex = 0
  for (let match = re.exec(body); match; match = re.exec(body)) {
    if (match[0] === '') re.lastIndex++
    else found.push({ start: match.index, end: match.index + match[0].length, match })
  }
  return found
}

/** Expands `$1`..`$99`, `$&` and `$$` of a regex replacement; other `$` stay as they are. */
export const expandReplacement = (template: string, match: RegExpExecArray): string => {
  const groups = match.length - 1
  let out = ''
  for (let i = 0; i < template.length; i++) {
    const next = template[i + 1] ?? ''
    if (template[i] !== '$') {
      out += template[i]
    } else if (next === '$' || next === '&') {
      out += next === '$' ? '$' : match[0]
      i++
    } else {
      const digits = /^\d\d?/.exec(template.slice(i + 1))?.[0] ?? ''
      const two = digits.length === 2 && Number(digits) <= groups ? digits : digits.slice(0, 1)
      const group = Number(two)
      if (two === '' || group < 1 || group > groups) {
        out += '$'
      } else {
        out += match[group] ?? ''
        i += two.length
      }
    }
  }
  return out
}

const splitLine = (line: string): [string, string] =>
  line.endsWith('\r') ? [line.slice(0, -1), '\r'] : [line, '']

/** Every occurrence in `text`, line by line (a match never spans lines). */
export const findInText = (text: string, query: string, options: SearchOptions): SearchMatch[] => {
  const re = buildRegExp(query, options)
  const matches: SearchMatch[] = []
  text.split('\n').forEach((line, index) => {
    const [body] = splitLine(line)
    for (const { start, end } of occurrences(re, body)) {
      matches.push({
        line: index + 1,
        column: start + 1,
        length: end - start,
        text: body.length > MAX_TEXT ? body.slice(0, MAX_TEXT) : body
      })
    }
  })
  return matches
}

export interface ReplaceAt {
  line: number
  column: number
}

/** Replaces every occurrence, or only the one that starts at `at`. Counts what it replaced. */
export const replaceInText = (
  text: string,
  query: string,
  options: SearchOptions,
  replacement: string,
  at?: ReplaceAt
): { text: string; count: number } => {
  const re = buildRegExp(query, options)
  let count = 0
  const lines = text.split('\n').map((line, index) => {
    if (at && at.line !== index + 1) return line
    const [body, ending] = splitLine(line)
    let out = ''
    let last = 0
    for (const { start, end, match } of occurrences(re, body)) {
      if (at && at.column !== start + 1) continue
      out +=
        body.slice(last, start) +
        (options.regex ? expandReplacement(replacement, match) : replacement)
      last = end
      count++
    }
    return out + body.slice(last) + ending
  })
  return { text: lines.join('\n'), count }
}

/**
 * The column of a match counted in characters, as the editor's locations are: search columns are
 * UTF-16 units (they index the line text), and an emoji before the match counts twice there.
 */
export const characterColumn = (match: SearchMatch): number =>
  Array.from(match.text.slice(0, match.column - 1)).length + 1
