/** The piece of a line that matters here: its text and whether the cursor waits at column 0. */
export interface CursorLine {
  text: string
  /** True after a `\r`: the next text overwrites the line from its start. */
  home?: boolean
}

const ESCAPE = '\u001b'

/**
 * What a terminal shows after `part` is written at column 0: it covers the start of the line and
 * the rest of the old text stays (so "Loading... 100%" then "\rDone" gives "Doneing... 100%").
 * Text with color codes is replaced whole: its visible length is not its length.
 */
const overwrite = (old: string, part: string): string => {
  if (part.length >= old.length || old.includes(ESCAPE) || part.includes(ESCAPE)) return part
  return part + old.slice(part.length)
}

/**
 * Writes `piece` (no `\n` in it) at the end of the line, the way a terminal does: every `\r`
 * sends the cursor back to column 0, so progress bars rewrite their line instead of
 * piling up.
 */
export const writeOver = <T extends CursorLine>(line: T, piece: string): T => {
  let text = line.text
  let home = line.home === true
  piece.split('\r').forEach((part, index) => {
    if (index > 0) home = true
    if (part === '') return
    text = home ? overwrite(text, part) : text + part
    home = false
  })
  const next: T = { ...line, text }
  if (home) next.home = true
  else delete next.home
  return next
}
