import Anser from 'anser'

/** A run of text with the ANSI style (colors, bold...) in force. Never HTML: the panel renders text. */
export interface AnsiPiece {
  text: string
  /** CSS color: a theme variable (`var(--ansi-red)`) or `rgb(...)`. */
  color?: string
  background?: string
  /** Subset of bold, dim, italic, underline, strikethrough. */
  decorations: string[]
}

const NAMED = /^ansi-((?:bright-)?(?:black|red|green|yellow|blue|magenta|cyan|white))$/
const PALETTE = /^ansi-palette-(\d{1,3})$/
const CUBE_STEPS = [0, 95, 135, 175, 215, 255]
const DECORATIONS = new Set(['bold', 'dim', 'italic', 'underline', 'strikethrough'])
const NAMES = ['black', 'red', 'green', 'yellow', 'blue', 'magenta', 'cyan', 'white']

/** Colors 16-255 of the xterm palette: a 6x6x6 cube, then 24 grays. */
const paletteColor = (index: number): string => {
  if (index < 8) return `var(--ansi-${NAMES[index]})`
  if (index < 16) return `var(--ansi-bright-${NAMES[index - 8]})`
  if (index >= 232) {
    const gray = 8 + (index - 232) * 10
    return `rgb(${gray}, ${gray}, ${gray})`
  }
  const cube = index - 16
  const [r, g, b] = [Math.floor(cube / 36), Math.floor(cube / 6) % 6, cube % 6]
  return `rgb(${CUBE_STEPS[r]}, ${CUBE_STEPS[g]}, ${CUBE_STEPS[b]})`
}

const TRUECOLOR = /^(\d{1,3}), (\d{1,3}), (\d{1,3})$/

/** Turns anser's color description into a CSS color, or undefined for "default". */
const cssColor = (name: string | null, truecolor: string | null): string | undefined => {
  if (!name) return undefined
  const named = NAMED.exec(name)
  if (named) return `var(--ansi-${named[1]})`
  const palette = PALETTE.exec(name)
  if (palette) return paletteColor(Math.min(Number(palette[1]), 255))
  if (name === 'ansi-truecolor' && truecolor && TRUECOLOR.test(truecolor)) {
    return `rgb(${truecolor})`
  }
  return undefined
}

/** Splits text with ANSI SGR codes into styled pieces. Other escape sequences are dropped. */
export const parseAnsi = (text: string): AnsiPiece[] => {
  if (!text.includes('\u001b')) return [{ text, decorations: [] }]
  return Anser.ansiToJson(text, { use_classes: true, remove_empty: true }).map((chunk) => ({
    text: chunk.content,
    color: cssColor(chunk.fg, chunk.fg_truecolor),
    background: cssColor(chunk.bg, chunk.bg_truecolor),
    decorations: chunk.decorations.filter((name) => DECORATIONS.has(name))
  }))
}
