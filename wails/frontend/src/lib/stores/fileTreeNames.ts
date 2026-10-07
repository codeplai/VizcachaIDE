// Names for copies: `name (copy).ext`, `name (copy 2).ext` (English) or `nombre (copia).ext` (Spanish).
import { get } from 'svelte/store'
import { locale } from '../i18n'

/** The word of the suffix for the UI language. */
const copyWord = (): string => (get(locale)?.startsWith('es') ? 'copia' : 'copy')

/** Splits "a.tar.gz" into "a.tar" and ".gz"; folders and dotfiles have no extension. */
const splitName = (name: string, isDir: boolean): [string, string] => {
  const dot = name.lastIndexOf('.')
  return isDir || dot <= 0 ? [name, ''] : [name.slice(0, dot), name.slice(dot)]
}

/**
 * A name for a copy of `name` that none of `taken` uses (case-insensitive). The plain name is kept
 * when it is free and `inPlace` is false (pasting into another folder).
 */
export const copyName = (
  name: string,
  isDir: boolean,
  taken: string[],
  inPlace: boolean
): string => {
  const used = new Set(taken.map((other) => other.toLowerCase()))
  if (!inPlace && !used.has(name.toLowerCase())) return name
  const [stem, extension] = splitName(name, isDir)
  const word = copyWord()
  for (let n = 1; ; n++) {
    const suffix = n === 1 ? `(${word})` : `(${word} ${n})`
    const candidate = `${stem} ${suffix}${extension}`
    if (!used.has(candidate.toLowerCase())) return candidate
  }
}
