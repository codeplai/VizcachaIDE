// Minimal visual marks for the W0 skeleton: debug line, breakpoints, squiggles and the inline hint.
// Track F1 owns this folder and replaces these with the full extensions (lint, gutters, hover...).
import { StateEffect, StateField, type EditorState, type Range } from '@codemirror/state'
import {
  Decoration,
  EditorView,
  GutterMarker,
  WidgetType,
  gutter,
  type DecorationSet
} from '@codemirror/view'

export interface ProblemMark {
  line: number
  /** 1-based columns, `to` exclusive. */
  from: number
  to: number
  title: string
  detail: string
}

export interface EditorMarks {
  debugLine: number | null
  breakpoints: number[]
  problems: ProblemMark[]
}

export const emptyMarks: EditorMarks = { debugLine: null, breakpoints: [], problems: [] }

export const setMarks = StateEffect.define<EditorMarks>()

export const marksField = StateField.define<EditorMarks>({
  create: () => emptyMarks,
  update: (value, transaction) => {
    for (const effect of transaction.effects) if (effect.is(setMarks)) return effect.value
    return value
  }
})

class HintWidget extends WidgetType {
  constructor(
    readonly title: string,
    readonly detail: string
  ) {
    super()
  }
  override eq(other: HintWidget): boolean {
    return other.title === this.title && other.detail === this.detail
  }
  override toDOM(): HTMLElement {
    const box = document.createElement('div')
    box.className = 'cm-inline-hint'
    box.append(this.title + ' ')
    const detail = document.createElement('span')
    detail.textContent = this.detail
    box.append(detail)
    return box
  }
}

const inDocument = (state: EditorState, line: number): boolean =>
  line >= 1 && line <= state.doc.lines

const problemDecorations = (state: EditorState, problem: ProblemMark): Range<Decoration>[] => {
  if (!inDocument(state, problem.line)) return []
  const line = state.doc.line(problem.line)
  const from = Math.min(line.from + problem.from - 1, line.to)
  const to = Math.min(line.from + problem.to - 1, line.to)
  const squiggle = Decoration.mark({ class: 'cm-problem' }).range(from, Math.max(to, from))
  const hint = Decoration.widget({
    widget: new HintWidget(problem.title, problem.detail),
    block: true,
    side: 1
  }).range(line.to)
  return to > from ? [squiggle, hint] : [hint]
}

const buildDecorations = (state: EditorState): DecorationSet => {
  const marks = state.field(marksField)
  const ranges: Range<Decoration>[] = marks.problems.flatMap((p) => problemDecorations(state, p))
  if (marks.debugLine !== null && inDocument(state, marks.debugLine)) {
    const line = state.doc.line(marks.debugLine)
    ranges.push(Decoration.line({ class: 'cm-debug-line' }).range(line.from))
  }
  return Decoration.set(ranges, true)
}

export const marksDecorations = EditorView.decorations.compute(
  [marksField, 'doc'],
  buildDecorations
)

class BreakpointMarker extends GutterMarker {
  override toDOM(): HTMLElement {
    const dot = document.createElement('span')
    dot.className = 'cm-bp'
    return dot
  }
}

const breakpointMarker = new BreakpointMarker()

/** Gutter column with the breakpoint dots; clicking a row asks the caller to toggle it. */
export const breakpointGutter = (onToggle: (line: number) => void) =>
  gutter({
    class: 'cm-bpgutter',
    lineMarker: (view, block) => {
      const number = view.state.doc.lineAt(block.from).number
      return view.state.field(marksField).breakpoints.includes(number) ? breakpointMarker : null
    },
    lineMarkerChange: (update) =>
      update.transactions.some((tr) => tr.effects.some((effect) => effect.is(setMarks))),
    initialSpacer: () => breakpointMarker,
    domEventHandlers: {
      mousedown: (view, block) => {
        onToggle(view.state.doc.lineAt(block.from).number)
        return true
      }
    }
  })
