// Inlay hints: the inferred types and parameter names the language server offers, drawn inside
// the code without changing it (docs/PLAN_RUST.md section 9.1). Generic: any language whose
// server answers has them; one that does not (pylsp) simply gives none.
import { StateEffect, StateField, type EditorState, type Extension } from '@codemirror/state'
import {
  Decoration,
  EditorView,
  ViewPlugin,
  WidgetType,
  type DecorationSet,
  type ViewUpdate
} from '@codemirror/view'
import type { InlayHint } from '../domain'
import type { DocumentContext, LanguageApi } from './documentContext'

export const INLAY_DELAY_MS = 300
/**
 * A server still indexing answers with no hints: ask again, for about three minutes. rust-analyzer
 * can take over a minute on a slow or busy PC, and a project without problems sends no diagnostics
 * (nor a new status, when another project already made the server ready) to trigger a refresh.
 */
export const INLAY_RETRY_MS = [2000, 4000, 8000, 16000, ...Array<number>(6).fill(30000)]

class InlayWidget extends WidgetType {
  constructor(readonly hint: InlayHint) {
    super()
  }
  override eq(other: InlayWidget): boolean {
    const { hint } = this
    return (
      other.hint.label === hint.label &&
      other.hint.kind === hint.kind &&
      other.hint.paddingLeft === hint.paddingLeft &&
      other.hint.paddingRight === hint.paddingRight
    )
  }
  override toDOM(): HTMLElement {
    const span = document.createElement('span')
    span.className = `cm-inlay-hint cm-inlay-hint-${this.hint.kind}`
    span.textContent = this.hint.label
    if (this.hint.paddingLeft) span.classList.add('cm-inlay-pad-left')
    if (this.hint.paddingRight) span.classList.add('cm-inlay-pad-right')
    return span
  }
  override ignoreEvent(): boolean {
    return true
  }
}

export const inlayTheme = EditorView.baseTheme({
  '.cm-inlay-hint': {
    color: 'var(--muted)',
    fontSize: '0.85em',
    userSelect: 'none',
    pointerEvents: 'none'
  },
  '.cm-inlay-pad-left': { marginLeft: '0.3em' },
  '.cm-inlay-pad-right': { marginRight: '0.3em' }
})

const setHints = StateEffect.define<DecorationSet>()

/** The hints on screen. They follow the text while it is edited, until fresher ones arrive. */
const hintsField = StateField.define<DecorationSet>({
  create: () => Decoration.none,
  update: (hints, transaction) => {
    for (const effect of transaction.effects) if (effect.is(setHints)) return effect.value
    return transaction.docChanged ? hints.map(transaction.changes) : hints
  },
  provide: (field) => EditorView.decorations.from(field)
})

/** The offset of a 1-based column counted in characters (not UTF-16 units), clamped to its line. */
const offsetAt = (state: EditorState, line: number, column: number): number | null => {
  if (line < 1 || line > state.doc.lines) return null
  const { from, text } = state.doc.line(line)
  let offset = 0
  for (let n = 1; n < column && offset < text.length; n++) {
    offset += (text.codePointAt(offset) ?? 0) > 0xffff ? 2 : 1
  }
  return from + offset
}

const toDecorations = (state: EditorState, hints: InlayHint[]): DecorationSet => {
  const ranges = hints.flatMap((hint) => {
    const at = offsetAt(state, hint.line, hint.column)
    if (at === null || !hint.label) return []
    return [Decoration.widget({ widget: new InlayWidget(hint), side: 1 }).range(at)]
  })
  return Decoration.set(ranges, true)
}

export interface InlayHints {
  extension: Extension
  /** Turns the hints on or off; effective at once, in every open editor state. */
  setEnabled: (enabled: boolean) => void
  /**
   * Asks again: a language server that was still loading answered with no hints, and its new
   * diagnostics (or its ready status) say it has analysed the file now.
   */
  refresh: () => void
}

/** The lines on screen, as a request for the server (end column 1: whole last line). */
const visibleLines = (view: EditorView, file: string) => {
  const ranges = view.visibleRanges
  const first = ranges.at(0)?.from ?? view.viewport.from
  const last = ranges.at(-1)?.to ?? view.viewport.to
  const start = view.state.doc.lineAt(first).number
  const end = view.state.doc.lineAt(last).number
  return {
    start: { file, line: start, column: 1 },
    end: { file, line: end, column: 1 }
  }
}

/** What the plugin of every editor state shares: whether hints are on, and who to tell. */
interface Switchboard {
  enabled: boolean
  listeners: Set<() => void>
}

/** Asks for the hints of the lines on screen, after edits and scrolling settle. */
class InlayRequests {
  private timer: ReturnType<typeof setTimeout> | undefined
  private ticket = 0
  private retries = 0
  private destroyed = false
  private readonly onToggle = (): void => this.restart()

  constructor(
    private readonly view: EditorView,
    private readonly language: LanguageApi,
    private readonly file: DocumentContext,
    private readonly board: Switchboard
  ) {
    board.listeners.add(this.onToggle)
    this.restart()
  }

  update(update: ViewUpdate): void {
    if (update.docChanged || update.viewportChanged) this.restart()
  }

  destroy(): void {
    this.destroyed = true
    this.board.listeners.delete(this.onToggle)
    clearTimeout(this.timer)
    this.ticket++
  }

  private restart(): void {
    clearTimeout(this.timer)
    this.ticket++
    this.retries = 0
    if (!this.board.enabled) {
      this.show(Decoration.none)
      return
    }
    this.timer = setTimeout(() => void this.ask(), INLAY_DELAY_MS)
  }

  private async ask(): Promise<void> {
    const path = this.file.path()
    if (!path) return
    const ticket = ++this.ticket
    try {
      await this.file.flush()
      const hints = await this.language.inlayHints(visibleLines(this.view, path))
      const stale = ticket !== this.ticket || !this.board.enabled || this.file.path() !== path
      if (stale) return
      this.show(toDecorations(this.view.state, hints))
      const wait = INLAY_RETRY_MS[this.retries]
      if (hints.length === 0 && wait !== undefined) {
        this.retries++
        this.timer = setTimeout(() => void this.ask(), wait)
      }
    } catch {
      // a failed query only means no hints this time
    }
  }

  private show(hints: DecorationSet): void {
    // The effect is dispatched outside an update, so never from inside one.
    queueMicrotask(() => {
      if (!this.destroyed) this.view.dispatch({ effects: setHints.of(hints) })
    })
  }
}

export const inlayHints = (language: LanguageApi, file: DocumentContext): InlayHints => {
  const board: Switchboard = { enabled: true, listeners: new Set() }
  const plugin = ViewPlugin.define((view) => new InlayRequests(view, language, file, board))
  const setEnabled = (value: boolean): void => {
    if (value === board.enabled) return
    board.enabled = value
    board.listeners.forEach((listener) => listener())
  }
  const refresh = (): void => {
    if (board.enabled) board.listeners.forEach((listener) => listener())
  }
  return { extension: [hintsField, plugin, inlayTheme], setEnabled, refresh }
}
